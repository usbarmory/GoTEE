// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package applet provides supervised mode initialization for bare metal Go
// unikernels written using the TamaGo framework.
//
// The package supports trusted applet execution under a GoTEE compatible
// supervisor, linking essential runtime functions with required system calls.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
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
	// Direct access to timers is allowed under supervised mode, so an
	// efficient implementation of nanotime1 in supervised mode simply
	// mirrors what TamaGo does natively:
	return CLINT.Nanotime()

	// A less efficient version, in case tiemrs are not accessible, is to
	// make a supervisor request:
	//   return syscall.Nanotime()
}

//go:linkname hwinit runtime.hwinit
func hwinit() {
	// no initialization required in supervised mode
}
