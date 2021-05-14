// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
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
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = KernelStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = KernelSize

//go:embed ta
var taELF []byte

//go:embed os_nonsecure
var osELF []byte

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	if !imx6.Native {
		return
	}

	if err := imx6.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("WARNING: error setting ARM frequency: %v", err))
	}

	usbarmory.LED("blue", true)

	debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
	<-debugConsole
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	mode := arm.ModeName(int(ctx.SPSR) & 0x1f)
	ns := ctx.NonSecure

	log.Printf("PL1 starting mode:%s ns:%v sp:%#.8x pc:%#.8x", mode, ns, ctx.R13, ctx.R15)

	err := ctx.Run()
	wg.Done()

	log.Printf("PL1 stopped mode:%s ns:%v sp:%#.8x lr:%#.8x pc:%#.8x err:%v", mode, ns, ctx.R13, ctx.R14, ctx.R15, err)
}

func main() {
	var wg sync.WaitGroup
	var ta *monitor.ExecCtx
	var os *monitor.ExecCtx
	var err error

	log.Printf("PL1 %s/%s (%s) • TEE system/supervisor", runtime.GOOS, runtime.GOARCH, runtime.Version())

	if ta, err = monitor.Load(taELF, AppletStart, AppletSize); err != nil {
		log.Fatalf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded applet addr:%#x size:%d entry:%#x", ta.Memory.Start, len(taELF), ta.R15)
	}

	// register receiver methods for RPC test (see rpc.go)
	ta.Server.Register(&Receiver{})
	ta.Debug = true

	// test concurrent execution of:
	//   Secure    World PL0 (supervisor mode) - secure OS (this program)
	//   Secure    World PL1 (user mode)       - trusted applet
	wg.Add(1)
	go run(ta, &wg)

	go func() {
		log.Printf("PL1 will sleep until PL0 is done")

		for i := 0; i < 60; i++ {
			time.Sleep(1 * time.Second)
			log.Printf("PL1 says %d missisipi", i+1)
		}
	}()
	wg.Wait()

	if len(osELF) == 0 {
		log.Printf("PL1 says goodbye")
		return
	}

	// test execution of:
	//   NonSecure World PL0 (supervisor mode) - main OS

	if os, err = monitor.Load(osELF, NonSecureStart, NonSecureSize); err != nil {
		log.Fatalf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start, len(osELF), os.R15)
	}

	os.NonSecure = true
	os.SPSR = monitor.SystemMode
	os.Debug = true

	// grant access to CP10 and CP11
	monitor.NonSecureAccess(1<<11 | 1<<10)

	// this will take over as we have no Secure/Normal World API yet
	run(os, &wg)

	// unreachable
	log.Printf("PL1 says goodbye")
}
