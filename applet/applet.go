// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package applet provides user mode initialization for bare metal Go
// unikernels written using the TamaGo framework.
//
// The package supports trusted applet execution under a GoTEE compatible
// supervisor, linking essential runtime functions with required system calls.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package applet

import (
	_ "unsafe"

	"github.com/f-secure-foundry/GoTEE/syscall"

	"github.com/f-secure-foundry/tamago/arm"
)

var ARM = &arm.CPU{}

func init() {
	if ARM.Mode() != arm.USR_MODE {
		panic("unexpected processor mode")
	}
}

//go:linkname printk runtime.printk
func printk(c byte) {
	syscall.Print(c)
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	// A more efficient version is (as tamago allows PL0 access to generic
	// counters):
	//	return int64(ARM.TimerFn() * ARM.TimerMultiplier)
	//
	// But to stress test things and have nicer non-interleaved logging we
	// keep the more demanding syscall for now.
	return syscall.Nanotime()
}

//go:linkname initRNG runtime.initRNG
func initRNG() {
	// no initialization required in user mode
}

//go:linkname hwinit runtime.hwinit
func hwinit() {
	ARM.InitGenericTimers(0, 0)
}

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
	syscall.GetRandom(b, uint(len(b)))
}

// Exit signals the applet termination to its supervisor.
func Exit() {
	syscall.Exit()
}
