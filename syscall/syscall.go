// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package syscall provides support for system call for TamaGo unikernels
// launched in supervised mode through monitor.Exec (see monitor package).
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go on ARM/RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package syscall

// defined in syscall_*.s

// Supervisors triggers a supervisor call (SWI/SVC).
func Supervisor()

// Exit terminates the execution context scheduling through a system call to
// the supervisor.
func Exit()

// Print prints a single character on standard output through a system call to
// the supervisor.
func Print(c byte)

// Nanotime returns the system time in nanoseconds through a system call to the
// supervisor.
func Nanotime() int64

// GetRandom fills a byte array with random values through a system call to the
// supervisor.
func GetRandom(b []byte, n uint) {
	Write(SYS_GETRANDOM, b, n)
}

// Read requests a transfer of n bytes into p from the supervisor through the
// syscall specified as first argument. It can be used to implement syscalls
// that require request/responses data streams, along with Write().
//
// The underlying connection used by the RPC client (see NewClient()) is an
// example of such implementation.
func Read(trap uint, p []byte, n uint) uint

// Write requests a transfer of n bytes from p to the supervisor through the
// syscall specified as first argument. It can be used to implement syscalls
// that require request/responses data streams, along with Read().
//
// The underlying connection used by the RPC client (see NewClient()) is an
// example of such implementation.
func Write(trap uint, p []byte, n uint)
