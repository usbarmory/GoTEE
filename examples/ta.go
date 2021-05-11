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
	"github.com/f-secure-foundry/GoTEE/syscall"
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

func testRNG(n int) {
	buf := make([]byte, n)
	syscall.GetRandom(buf, uint(n))
	log.Printf("PL0 obtained %d random bytes from PL1: %x", n, buf)
}

func testRPC() {
	rpcClient := syscall.NewClient()
	defer rpcClient.Close()

	res := ""
	req := "hello"
	log.Printf("PL0 requests echo via RPC: %s", req)

	err := rpcClient.Call("Receiver.Echo", req, &res)

	if err != nil {
		log.Printf("PL0 received RPC error: %v", err)
	} else {
		log.Printf("PL0 received echo via RPC: %s", res)
	}
}

func testInvalidAccess() {
	pl1TextStart := tee.KernelStart + uint32(0x10000)
	mem := (*uint32)(unsafe.Pointer(uintptr(pl1TextStart)))

	log.Printf("PL0 is about to read PL1 memory at %#x", pl1TextStart)
	val := atomic.LoadUint32(mem)

	res := "success (shouldn't happen)"

	if val != 0xe59a1008 {
		res = "fail (expected, but you should never see this)"
	}

	log.Printf("PL0 read PL1 memory %#x: %#x (%s)", pl1TextStart, val, res)
}

func testAbort() {
	var p *byte

	log.Printf("PL0 is about to trigger data abort")
	*p = 0xab
}

func main() {
	log.Printf("PL0 %s/%s (%s) • TEE user applet", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// test syscall interface
	testRNG(16)

	// test RPC interface
	testRPC()

	log.Printf("PL0 will sleep for 5 seconds")

	// test concurrent execution of PL1 applet and PL0 supervisor
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
