// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !syscall_nanotime

package applet

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
)

var ARM = &arm.CPU{}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return ARM.GetTime()
}

func initTimers() {
	ARM.InitGenericTimers(0, 0)
}
