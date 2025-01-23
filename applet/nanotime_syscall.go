// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build syscall_nanotime

package applet

import (
	_ "unsafe"

	"github.com/usbarmory/GoTEE/syscall"
)

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	// Supervisor request, used when PL0 has no direct access to timers or
	// when indirection is desired (e.g.  soft lockstep).
	return syscall.Nanotime()
}

func initTimers() {
	// no initialization required
}
