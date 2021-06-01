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
	"time"
	_ "unsafe"

	"github.com/f-secure-foundry/GoTEE/applet"
	"github.com/f-secure-foundry/GoTEE/syscall"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = AppletStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = AppletSize

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)
}

func testRNG(n int) {
	buf := make([]byte, n)
	syscall.GetRandom(buf, uint(n))
	log.Printf("PL0 obtained %d random bytes from PL1: %x", n, buf)
}

func testRPC() {
	res := ""
	req := "hello"

	log.Printf("PL0 requests echo via RPC: %s", req)
	err := syscall.Call("RPC.Echo", req, &res)

	if err != nil {
		log.Printf("PL0 received RPC error: %v", err)
	} else {
		log.Printf("PL0 received echo via RPC: %s", res)
	}
}

func testAbort() {
	var p *byte

	log.Printf("PL0 is about to trigger data abort")
	*p = 0xab
}

func main() {
	log.Printf("PL0 %s/%s (%s) • TEE user applet (Secure World)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// test syscall interface
	testRNG(16)

	// test RPC interface
	testRPC()

	log.Printf("PL0 will sleep for 5 seconds")

	ledStatus := LEDStatus{
		Name: "blue",
		On:   true,
	}

	// test concurrent execution of PL0 applet and PL1 supervisor
	for i := 0; i < 5; i++ {
		syscall.Call("RPC.LED", ledStatus, nil)
		ledStatus.On = !ledStatus.On

		time.Sleep(1 * time.Second)
		log.Printf("PL0 says %d missisipi", i+1)
	}

	// test memory protection
	testInvalidAccess("PL0")

	// this should be unreachable

	// test exception handling
	testAbort()

	// terminate applet
	applet.Exit()
}
