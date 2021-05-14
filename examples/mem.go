// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

// TODO: for now we take a lazy approach of allocating 32MB for each kernel

const (
	KernelStart = 0x80000000
	KernelSize  = 0x2000000

	AppletStart = 0x82000000
	AppletSize  = 0x2000000

	NonSecureStart = 0x84000000
	NonSecureSize  = 0x2000000
)
