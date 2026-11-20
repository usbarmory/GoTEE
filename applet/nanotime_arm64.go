// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !syscall_nanotime

package applet

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/arm64"
)

var ARM64 = &arm64.CPU{}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return ARM64.GetTime()
}

func initTimers() {
	ARM64.InitGenericTimers(0, 0)
}
