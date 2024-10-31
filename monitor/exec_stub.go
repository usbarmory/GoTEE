// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// stub for pkg.go.dev coverage
//go:build !tamago

// Package monitor provides supervisor support for TamaGo unikernels to allow
// scheduling of Secure user mode or NonSecure system mode (ARM) or Supervisor
// mode (RISC-V) executables.
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go on ARM/RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package monitor

// Exec allows execution of an executable in Secure user mode or NonSecure
// system mode (ARM) or Supervisor mode (RISC-V).
//
// The execution is isolated from the invoking Go runtime, yielding back to it
// is supported through exceptions (e.g. syscalls through SVC on ARM and ECALL
// on RISC-V).
//
// The execution context pointer allows task initialization and it is updated
// with the program state at return, it can therefore be passed again to resume
// the task.
func Exec(ctx *ExecCtx)

// ExecCtx represents a executable initialization or returning state.
type ExecCtx struct {}

// String returns the string form of the execution context registers.
func (ctx *ExecCtx) String()

// Secure (RISC-V) returns whether the execution context is loaded as trusted
// applet.
func (ctx *ExecCtx) Secure() bool

// NonSecure (ARM) returns whether the execution context is loaded as
// non-secure.
func (ctx *ExecCtx) NonSecure() bool

// Cause (RISC-V) returns the trap event.
func (ctx *ExecCtx) Cause() (code uint64, irq bool)

// Mode (ARM) returns the processor mode.
func (ctx *ExecCtx) Mode() (current int, saved int)

// Schedule runs the execution context until an exception is caught.
//
// Unlike Run() the function does not invoke the context Handler(), there
// exceptions and system or monitor calls are not handled.
func (ctx *ExecCtx) Schedule() (err error)

// Run starts the execution context and handles system or monitor calls. The
// execution yields back to the invoking Go runtime only when exceptions are
// caught.
//
// The function invokes the context Handler() and returns when an unhandled
// exception, or any other error, is raised.
func (ctx *ExecCtx) Run() (err error)

// Stop stops the execution context.
func (ctx *ExecCtx) Stop()

// Done returns a channel which will be closed once execution context has stopped.
func (ctx *ExecCtx) Done() chan struct{}

// Load returns an execution context initialized for the argument entry point
// and memory region
//
// ARM: the secure flag controls whether the context belongs to a secure
// partition (e.g. TrustZone Secure World) or a non-secure one (e.g.  TrustZone
// Normal World). In case of a non-secure execution context, the memory is
// configured as NonSecure by means of MMU NS bit and memory controller region
// configuration. The caller is responsible for any other required MMU
// configuration (see arm.ConfigureMMU()) or additional peripheral restrictions
// (e.g. TrustZone).
//
// RISC-V: any additional peripheral restrictions are up to the caller.
func Load(entry uint, mem *dma.Region, secure bool) (ctx *ExecCtx, err error)

// Equal returns whether a and b holds the same register state.
func Equal(a, b *ExecCtx) bool
