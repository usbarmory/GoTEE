// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package monitor provides supervisor support for TamaGo unikernels to allow
// scheduling of isolated supervisor executables.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package monitor

import (
	"errors"
	"fmt"
	"net/rpc"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/riscv"
	"github.com/usbarmory/tamago/soc/sifive/fu540"
)

// RISC-V privilege levels
const (
	User       = 0b00 // U
	Supervisor = 0b01 // S
	Machine    = 0b11 // M
)

var mux sync.Mutex

// defined in exec_riscv64.s
func monitor()

func init() {
	if err := fu540.RV64.WritePMP(0, (1<<64)-1, true, true, true, riscv.PMP_A_TOR, false); err != nil {
		panic("could not set PMP default entry")
	}
}

// Exec allows execution of an executable in supervisor mode. The execution is
// isolated from the invoking Go runtime, yielding back to it is supported
// through exceptions (e.g. syscalls through ECALL).
//
// The execution context pointer allows task initialization and it is updated
// with the program state at return, it can therefore be passed again to resume
// the task.
func Exec(ctx *ExecCtx)

// ExecCtx represents a executable initialization or returning state.
type ExecCtx struct {
	X1  uint64 // RA
	X2  uint64 // SP
	X3  uint64 // GP
	X4  uint64 // TP
	X5  uint64 // T0
	X6  uint64 // T1
	X7  uint64 // T2
	X8  uint64 // S0/FP
	X9  uint64 // S1
	X10 uint64 // A0
	X11 uint64 // A1
	X12 uint64 // A2
	X13 uint64 // A3
	X14 uint64 // A4
	X15 uint64 // A5
	X16 uint64 // A6
	X17 uint64 // A7
	X18 uint64 // S2
	X19 uint64 // S3
	X20 uint64 // S4
	X21 uint64 // S5
	X22 uint64 // S6
	X23 uint64 // S7
	X24 uint64 // S8
	X25 uint64 // S9
	X26 uint64 // S10
	X27 uint64 // S11 (g)
	X28 uint64 // T3
	X29 uint64 // T4
	X30 uint64 // T5
	X31 uint64 // T6

	// Program Counter
	PC uint64
	// Machine Exception Program Counter
	MEPC uint64
	// Machine Cause
	MCAUSE uint64

	// floating-point registers
	F [32]uint64 // F0-F31

	// Memory is the executable allocated RAM
	Memory *dma.Region

	// PMP, if not nil, handles physical memory protection for the
	// environment invoking the execution context.
	//
	// The integer argument is passed as first available PMP entry, the PMP
	// function must not modify previous entries as they are used at
	// scheduling to grant execution context access to its own memory.
	PMP func(ctx *ExecCtx, pmpEntry int) error

	// Handler, if not nil, handles syscalls
	Handler func(ctx *ExecCtx) error

	// Server, if not nil, serves RPC calls over syscalls
	Server *rpc.Server

	// Lockstep, if not nil, enables delayed execution of a redundant
	// execution context (see Shadow) for fault detection.
	//
	// The Shadow context yields at each monitor call for comparison (see
	// Equal), in case of a mismatch the primary context Run() raises an
	// error.
	//
	// The Lockstep function is responsible for virtual addressing
	// re-configuration (see arm.ConfigureMMU) to redirect each context to
	// its physical memory.
	Lockstep func(shadow bool)
	// Shadow represents the redundant execution context allocated for
	// lockstep execution, it is set by Run() when Lockstep is not nil.
	Shadow *ExecCtx

	// execution state
	run bool
	// stopped will be closed once the context has stopped running.
	stopped chan struct{}
	// trusted applet flag
	secure bool
	// executing g stack pointer
	g_sp uint64
	// shadow state
	isShadow bool

	// Read() buffer
	in []byte
	// Write() buffer
	out []byte
}

// String returns the string form of the execution context registers.
func (ctx *ExecCtx) String() string {
	var sb strings.Builder

	code, _ := ctx.Cause()

	fmt.Fprintf(&sb, "\n")
	fmt.Fprintf(&sb, "   ra:%.16x  sp:%.16x  gp:%.16x  tp:%.16x\n", ctx.X1, ctx.X2, ctx.X3, ctx.X4)
	fmt.Fprintf(&sb, "   t0:%.16x  t1:%.16x  t2:%.16x  s0:%.16x\n", ctx.X5, ctx.X6, ctx.X7, ctx.X8)
	fmt.Fprintf(&sb, "   s1:%.16x  a0:%.16x  a1:%.16x  a2:%.16x\n", ctx.X9, ctx.X10, ctx.X11, ctx.X12)
	fmt.Fprintf(&sb, "   a3:%.16x  a4:%.16x  a5:%.16x  a6:%.16x\n", ctx.X13, ctx.X14, ctx.X15, ctx.X16)
	fmt.Fprintf(&sb, "   a7:%.16x  s2:%.16x  s3:%.16x  s4:%.16x\n", ctx.X17, ctx.X18, ctx.X19, ctx.X20)
	fmt.Fprintf(&sb, "   s5:%.16x  s6:%.16x  s7:%.16x  s8:%.16x\n", ctx.X21, ctx.X22, ctx.X23, ctx.X24)
	fmt.Fprintf(&sb, "   s9:%.16x s10:%.16x s11:%.16x  t3:%.16x\n", ctx.X25, ctx.X26, ctx.X27, ctx.X28)
	fmt.Fprintf(&sb, "   t4:%.16x  t5:%.16x  t6:%.16x  pc:%.16x err:%d\n", ctx.X29, ctx.X30, ctx.X31, ctx.PC, code)

	return sb.String()
}

// Secure returns whether the execution context is loaded as trusted applet.
func (ctx *ExecCtx) Secure() bool {
	return ctx.secure
}

// Cause returns the trap event.
func (ctx *ExecCtx) Cause() (code uint64, irq bool) {
	code = (ctx.MCAUSE &^ (1 << 63))
	irq = (ctx.MCAUSE >> 63) == 1
	return
}

func (ctx *ExecCtx) schedule() (err error) {
	var pmpEntry int

	mux.Lock()
	defer mux.Unlock()

	// set monitor handlers
	fu540.RV64.SetExceptionHandler(monitor)

	// grant execution context access to its own memory
	if pmpEntry, err = ctx.pmp(); err != nil {
		return
	}

	if ctx.PMP != nil {
		// set up application physical memory protection
		if err = ctx.PMP(ctx, pmpEntry); err != nil {
			return
		}
	}

	// execute applet
	Exec(ctx)

	// restore default handlers
	fu540.RV64.SetExceptionHandler(riscv.DefaultExceptionHandler)

	code, irq := ctx.Cause()

	if code != riscv.EnvironmentCallFromS || irq {
		return fmt.Errorf("%x", code)
	}

	return
}

// Run starts the execution context and handles system or monitor calls. The
// execution yields back to the invoking Go runtime only when exceptions are
// caught.
//
// The function invokes the context Handler() and returns when an unhandled
// exception, or any other error, is raised.
func (ctx *ExecCtx) Run() (err error) {
	ctx.run = true
	ctx.stopped = make(chan struct{})
	defer close(ctx.stopped)
	if ctx.Lockstep != nil {
		switch {
		case !ctx.isShadow && ctx.Shadow == nil:
			shadow := *ctx
			shadow.Handler = lockstepHandler
			shadow.isShadow = true

			ctx.Shadow = &shadow
		case ctx.isShadow:
			ctx.Lockstep(true)
			defer ctx.Lockstep(false)
		}
	}

	for ctx.run {
		if err = ctx.schedule(); err != nil {
			break
		}

		if ctx.Shadow != nil {
			if err = ctx.Shadow.Run(); err != nil {
				break
			}

			if !Equal(ctx, ctx.Shadow) {
				err = errors.New("lockstep failure")
				break
			}
		}

		if ctx.Handler != nil {
			if err = ctx.Handler(ctx); err != nil {
				break
			}
		}

		runtime.Gosched()
	}

	return
}

// Stop stops the execution context.
func (ctx *ExecCtx) Stop() {
	mux.Lock()
	defer mux.Unlock()

	ctx.run = false
}

// Done returns a channel which will be closed once execution context has stopped.
func (ctx *ExecCtx) Done() chan struct{} {
	return ctx.stopped
}

// Load returns an execution context initialized for the argument entry point
// and memory region.
//
// Any additional peripheral restrictions are up to the caller.
func Load(entry uint, mem *dma.Region, secure bool) (ctx *ExecCtx, err error) {
	ctx = &ExecCtx{
		PC:     uint64(entry),
		Memory: mem,
		Server: rpc.NewServer(),
		secure: secure,
	}

	if secure {
		ctx.Handler = SecureHandler
	} else {
		ctx.Handler = NonSecureHandler
	}

	return
}

// pmp grants context access to its own memory.
func (ctx *ExecCtx) pmp() (pmpEntry int, err error) {
	if err = fu540.RV64.WritePMP(pmpEntry, uint64(ctx.Memory.Start()), false, false, false, riscv.PMP_A_OFF, false); err != nil {
		return
	}
	pmpEntry += 1

	if err = fu540.RV64.WritePMP(pmpEntry, uint64(ctx.Memory.End()), true, true, true, riscv.PMP_A_TOR, false); err != nil {
		return
	}
	pmpEntry += 1

	return
}

// Equal returns whether a and b have the same register state, it is meant to
// be used for lockstep verification of primary and shadow execution contexts.
func Equal(a, b *ExecCtx) bool {
	return (a.X1 == b.X1 &&
		a.X2 == b.X2 &&
		a.X3 == b.X3 &&
		a.X4 == b.X4 &&
		a.X5 == b.X5 &&
		a.X6 == b.X6 &&
		a.X7 == b.X7 &&
		a.X8 == b.X8 &&
		a.X9 == b.X9 &&
		a.X10 == b.X10 &&
		a.X11 == b.X11 &&
		a.X12 == b.X12 &&
		a.X13 == b.X13 &&
		a.X14 == b.X14 &&
		a.X15 == b.X15 &&
		a.X16 == b.X16 &&
		a.X17 == b.X17 &&
		a.X18 == b.X18 &&
		a.X19 == b.X19 &&
		a.X20 == b.X20 &&
		a.X21 == b.X21 &&
		a.X22 == b.X22 &&
		a.X23 == b.X23 &&
		a.X24 == b.X24 &&
		a.X25 == b.X25 &&
		a.X26 == b.X26 &&
		a.X27 == b.X27 &&
		a.X28 == b.X28 &&
		a.X29 == b.X29 &&
		a.X30 == b.X30 &&
		a.X31 == b.X31 &&
		a.PC == b.PC &&
		a.MCAUSE == b.MCAUSE &&
		a.F == b.F &&
		slices.Equal(a.in, b.in))
}
