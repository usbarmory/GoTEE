// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/f-secure-foundry/GoTEE"
	"github.com/f-secure-foundry/GoTEE/applet"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = tee.AppletStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = tee.AppletSize

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = tee.AppletStackOffset

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)
}

func testInvalidAccess() {
	pl1RamStart := uint32(0x90010000)

	log.Printf("PL0 about to read PL1 memory at %#x", pl1RamStart)

	mem := (*uint32)(unsafe.Pointer(uintptr(pl1RamStart)))
	val := atomic.LoadUint32(mem)
	res := "success (shouldn't happen)"

	if val != 0xe59a1008 {
		res = "fail (expected, but you should never see this)"
	}

	log.Printf("PL0 read PL1 memory %#x: %#x (%s)", pl1RamStart, val, res)
}

func testAbort() {
	var p *byte

	log.Printf("PL0 is about to trigger data abort")
	*p = 0xab
}

func main() {
	log.Printf("PL0 %s/%s (%s) • TEE user applet", runtime.GOOS, runtime.GOARCH, runtime.Version())
	log.Printf("PL0 will sleep for 5 seconds")

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		log.Printf("PL0 says %d missisipi", i+1)
	}

	// test memory protection
	testInvalidAccess()

	// this should be unreachable

	// test exception handling
	testAbort()

	// terminate applet
	applet.Exit()
}
