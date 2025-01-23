// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !syscall_nanotime

package applet

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/sifive/clint"
)

var CLINT = &clint.CLINT{
	Base:   0x2000000,
	RTCCLK: 1000000,
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return CLINT.Nanotime()
}

func initTimers() {
	// no initialization required in supervised mode
}
