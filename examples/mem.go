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

// TODO: for now we take a lazy approach of allocating 32MB for each kernel
const (
	KernelStart = 0x80000000
	KernelSize  = 0x02000000

	AppletStart = 0x82000000
	AppletSize  = 0x02000000

	NonSecureStart = 0x84000000
	NonSecureSize  = 0x02000000
)

func testInvalidAccess() {
	pl1TextStart := KernelStart + uint32(0x10000)
	mem := (*uint32)(unsafe.Pointer(uintptr(pl1TextStart)))

	log.Printf("PL0 is about to read PL1 memory at %#x", pl1TextStart)
	val := atomic.LoadUint32(mem)

	res := "success - FIXME: shouldn't happen"

	if val != 0xe59a1008 {
		res = "fail (expected, but you should never see this)"
	}

	log.Printf("PL0 read PL1 memory %#x: %#x (%s)", pl1TextStart, val, res)
}
