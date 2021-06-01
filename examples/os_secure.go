// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
	_ "unsafe"

	"github.com/f-secure-foundry/GoTEE/monitor"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/csu"
	"github.com/f-secure-foundry/tamago/soc/imx6/dcp"
	"github.com/f-secure-foundry/tamago/soc/imx6/tzasc"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = KernelStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = KernelSize

//go:embed ta
var taELF []byte

//go:embed os_nonsecure
var osELF []byte

// RPC is an example receiver for user mode <--> system RPC over system calls.
type RPC struct{}

func (r *RPC) Echo(in string, out *string) error {
	*out = in
	return nil
}

func (r *RPC) LED(led LEDStatus, _ *string) error {
	if !imx6.Native {
		return errors.New("unsupported")
	}

	switch led.Name {
	case "white", "White", "WHITE":
		return errors.New("LED is secure only")
	case "blue", "Blue", "BLUE":
		return usbarmory.LED(led.Name, led.On)
	default:
		return errors.New("invalid LED")
	}

	return nil
}

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	if !imx6.Native {
		return
	}

	if err := imx6.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("WARNING: error setting ARM frequency: %v", err))
	}

	// white LED is used by Trusted OS example
	usbarmory.LED("white", true)
	// blue LED is requested by Trusted Applet
	usbarmory.LED("blue", false)

	debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
	<-debugConsole
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	mode := arm.ModeName(int(ctx.SPSR) & 0x1f)
	ns := ctx.NonSecure()

	log.Printf("PL1 starting mode:%s ns:%v sp:%#.8x pc:%#.8x", mode, ns, ctx.R13, ctx.R15)

	err := ctx.Run()

	if wg != nil {
		wg.Done()
	}

	log.Printf("PL1 stopped mode:%s ns:%v sp:%#.8x lr:%#.8x pc:%#.8x err:%v", mode, ns, ctx.R13, ctx.R14, ctx.R15, err)
}

func loadApplet() (ta *monitor.ExecCtx) {
	var err error

	log.Printf("PL1 %s/%s (%s) • TEE system/monitor (Secure World)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	if ta, err = monitor.Load(taELF, AppletStart, AppletSize, true); err != nil {
		log.Fatalf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded applet addr:%#x size:%d entry:%#x", ta.Memory.Start, len(taELF), ta.R15)
	}

	// register example RPC receiver
	ta.Server.Register(&RPC{})
	ta.Debug = true

	return
}

func loadNormalWorld(lock bool) (os *monitor.ExecCtx) {
	var err error

	if lock {
		// Move DMA region prevent free NonSecure access, alternatively
		// iRAM/OCRAM (default DMA region) can be locked down.
		//
		// This is necessary as iRAM/OCRAM is outside TZASC control.
		dma.Init(dmaStart, dmaSize)
	}

	if os, err = monitor.Load(osELF, NonSecureStart, NonSecureSize, false); err != nil {
		log.Fatalf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start, len(osELF), os.R15)
	}

	os.Debug = true

	// grant NonSecure access to CP10 and CP11
	monitor.NonSecureAccess(1<<11 | 1<<10)

	if !imx6.Native {
		return
	}

	// For readability purposes this example does not check for csu/tzasc
	// errors (which only traps invalid arguments being passed) are not
	// checked. You, however, should.

	csu.Init()

	// grant NonSecure access to all peripherals
	for i := csu.CSL_MIN; i < csu.CSL_MAX; i++ {
		csu.SetSecurityLevel(i, 0, csu.SEC_LEVEL_0, false)
		csu.SetSecurityLevel(i, 1, csu.SEC_LEVEL_0, false)
	}

	// TZASC NonSecure World R/W access
	tzasc.EnableRegion(1, NonSecureStart, NonSecureSize, (1<<tzasc.SP_NW_RD)|(1<<tzasc.SP_NW_WR))

	if !lock {
		return
	}

	// set all masters to NonSecure
	for i := csu.SA_MIN; i < csu.SA_MAX; i++ {
		csu.SetAccess(i, false, false)
	}

	// set Cortex-A7 master as Secure
	csu.SetAccess(0, true, false)

	// restrict ROMCP
	csu.SetSecurityLevel(13, 0, csu.SEC_LEVEL_4, false)

	// restrict TZASC
	csu.SetSecurityLevel(16, 1, csu.SEC_LEVEL_4, false)

	// restrict LEDs (GPIO4, IOMUXC)
	csu.SetSecurityLevel(2, 1, csu.SEC_LEVEL_4, false)
	csu.SetSecurityLevel(6, 1, csu.SEC_LEVEL_4, false)

	// restrict DCP
	csu.SetSecurityLevel(34, 0, csu.SEC_LEVEL_4, false)
	csu.SetAccess(14, true, false)

	return
}

func main() {
	var wg sync.WaitGroup

	ta := loadApplet()
	os := loadNormalWorld(false)

	// test concurrent execution of:
	//   Secure    World PL1 (system/monitor mode) - secure OS (this program)
	//   Secure    World PL0 (user mode)           - trusted applet
	//   NonSecure World PL1                       - main OS
	wg.Add(2)
	go run(ta, &wg)
	go run(os, &wg)

	go func() {
		log.Printf("PL1 will sleep until applet and kernel are done")

		for i := 0; i < 60; i++ {
			time.Sleep(1 * time.Second)
			log.Printf("PL1 says %d missisipi", i+1)
		}
	}()
	wg.Wait()

	usbarmory.LED("blue", false)

	if imx6.Native {
		// re-launch NonSecure World with peripheral restrictions
		os := loadNormalWorld(true)

		log.Printf("PL1 re-launching kernel with TrustZone restrictions")
		run(os, nil)

		// test restricted peripheral in Secure World
		log.Printf("PL1 in Secure World is about to perform DCP key derivation")

		k, err := dcp.DeriveKey(make([]byte, 8), make([]byte, 16), -1)

		if err != nil {
			log.Printf("PL1 in Secure World World failed to use DCP (%v)", err)
		} else {
			log.Printf("PL1 in Secure World World successfully used DCP (%x)", k)
		}
	}

	log.Printf("PL1 says goodbye")
}
