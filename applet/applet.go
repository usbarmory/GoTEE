// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package applet

import (
	_ "unsafe"

	"github.com/usbarmory/GoTEE/syscall"
)

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

// Exit signals the applet termination to its supervisor, it can be used as
// runtime.Exit to yield to monitor on runtime panic.
func Exit() {
	syscall.Exit()
}
