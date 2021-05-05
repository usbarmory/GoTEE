// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package applet

import (
	_ "unsafe"

	"github.com/f-secure-foundry/GoTEE/syscall"
)

//go:linkname printk runtime.printk
func printk(c byte) {
	syscall.Write(c)
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return syscall.Utime()
}

//go:linkname initRNG runtime.initRNG
func initRNG() {
	// no initialization required in user mode
}

//go:linkname hwinit runtime.hwinit
func hwinit() {
	// no initialization required in user mode
}

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
	syscall.GetRandom(b, uint(len(b)))
}

// Exit signals the applet termination to its supervisor.
func Exit() {
	syscall.Exit()
}
