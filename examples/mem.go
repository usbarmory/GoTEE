// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"sync/atomic"
	"unsafe"
)

// This example layout allocates 32MB for each kernel.
const (
	// Secure World OS
	KernelStart = 0x80000000
	KernelSize  = 0x01f00000

	// Secure World DMA (relocated to avoid conflicts with NonSecure world)
	dmaStart = 0x81f00000
	dmaSize  = 0x00100000

	// Secure World Applet
	AppletStart = 0x82000000
	AppletSize  = 0x02000000

	// NonSecure World OS
	NonSecureStart = 0x84000000
	NonSecureSize  = 0x02000000
)

func testInvalidAccess(tag string) {
	pl1TextStart := KernelStart + uint32(0x10000)
	mem := (*uint32)(unsafe.Pointer(uintptr(pl1TextStart)))

	log.Printf("%s is about to read PL1 Secure World memory at %#x", tag, pl1TextStart)
	val := atomic.LoadUint32(mem)

	res := "success - FIXME: shouldn't happen"

	if val != 0xe59a1008 {
		res = "fail (expected, but you should never see this)"
	}

	log.Printf("%s read PL1 Secure World memory %#x: %#x (%s)", tag, pl1TextStart, val, res)
}
