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
	"time"
	_ "unsafe"

	"github.com/f-secure-foundry/GoTEE"
	"github.com/f-secure-foundry/GoTEE/monitor"

	"github.com/f-secure-foundry/tamago/soc/imx6"

	_ "github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
)

var Build string
var Revision string

var banner string

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = tee.KernelStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = tee.KernelSize

//go:embed ta
var appletELF []byte

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	if !imx6.Native {
		return
	}

	if err := imx6.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("WARNING: error setting ARM frequency: %v", err))
	}
}

func main() {
	log.Printf("PL1 %s/%s (%s) • TEE system/supervisor", runtime.GOOS, runtime.GOARCH, runtime.Version())

	applet := monitor.Load(appletELF)
	applet.Handler = monitor.Handler
	applet.Debug = true

	log.Printf("PL1 loaded applet addr:%#x size:%d entry:%#x", applet.Memory.Start, len(appletELF), applet.R15)
	go applet.Run()

	for {
		time.Sleep(1 * time.Second)
		log.Printf("PL1 is sleeping in system mode")
	}
}
