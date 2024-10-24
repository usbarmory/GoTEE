// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package applet provides user mode initialization for bare metal Go
// unikernels written using the TamaGo framework.
//
// The package supports trusted applet execution under a GoTEE compatible
// supervisor, linking essential runtime functions with required system calls.
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go on ARM/RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package applet

import (
	_ "unsafe"

	"github.com/usbarmory/GoTEE/syscall"
)

//go:linkname hwinit runtime.hwinit
func hwinit() {
	initTimers()
}

//go:linkname printk runtime.printk
func printk(c byte) {
	syscall.Print(c)
}

//go:linkname initRNG runtime.initRNG
func initRNG() {
	// no initialization required in supervised mode
}

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
	syscall.GetRandom(b, uint(len(b)))
}

// Exit signals the applet termination to its supervisor.
func Exit() {
	syscall.Exit()
}

// Crash forces a nil pointer dereference to terminate the applet through an
// exception, it is meant to be used as runtime.Exit to yield to monitor on
// runtime panic.
func Crash(_ int32) {
	*(*int)(nil) = 0
}
