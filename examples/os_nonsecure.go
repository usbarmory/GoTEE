// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	_ "unsafe"

	"github.com/f-secure-foundry/GoTEE/syscall"

	_ "github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = NonSecureStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = NonSecureSize

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	if !imx6.Native {
		return
	}

	if err := imx6.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("WARNING: error setting ARM frequency: %v", err))
	}
}

func main() {
	log.Printf("PL1 %s/%s (%s) • system/supervisor (Normal World)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// yield back to secure monitor
	syscall.Monitor()

	// this should be unreachable
	log.Printf("PL1 says goodbye (NS)")
}
