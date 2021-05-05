// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package syscall

import (
	_ "unsafe"
)

// defined in syscall.s
func svc()
func Exit()
func Write(byte)
func Utime() int64
func GetRandom([]byte, uint)
