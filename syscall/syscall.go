// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package syscall

// defined in syscall.s

// svc triggers a supervisor exception
func svc()

// Exit terminates the user mode executable scheduling.
func Exit()

// Write prints a single character on standard output.
func Write(c byte)

// Nanotime returns the system time in nanoseconds.
func Nanotime() int64

// GetRandom fills a byte array with random values.
func GetRandom(b []byte, n uint)
