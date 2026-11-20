// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package monitor provides supervisor support for TamaGo unikernels to allow
// scheduling of Secure user mode executables.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
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

	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/dma"
)

var mux sync.Mutex

// defined in exec_arm64.s
func monitor()

func init() {
	// set lower handlers
	cpu := arm64.CPU{}
	cpu.SetLowerVectorTable(monitor, monitor)
}

// Exec allows execution of an executable in Secure user mode.  The execution
// is isolated from the invoking Go runtime, yielding back to it is supported
// through exceptions (e.g. syscalls through SVC).
//
// The execution context pointer allows task initialization and it is updated
// with the program state at return, it can therefore be passed again to resume
// the task.
func Exec(ctx *ExecCtx)

// ExecCtx represents a executable initialization or returning state.
type ExecCtx struct {
	R0  uint64
	R1  uint64
	R2  uint64
	R3  uint64
	R4  uint64
	R5  uint64
	R6  uint64
	R7  uint64
	R8  uint64
	R9  uint64
	R10 uint64
	R11 uint64
	R12 uint64
	R13 uint64
	R14 uint64
	R15 uint64
	R16 uint64
	R17 uint64
	R18 uint64
	R19 uint64
	R20 uint64
	R21 uint64
	R22 uint64
	R23 uint64
	R24 uint64
	R25 uint64
	R26 uint64
	R27 uint64
	R28 uint64 // g
	R29 uint64 // FP
	R30 uint64 // LR
	R31 uint64 // SP

	// Program Counter
	PC uint64

	// Fault Address Register
	FAR uint64
	// Exception Syndrome Register
	ESR uint64
	// SPSR (Saved Program Status Register) is the PSTATE of the execution
	// context as it raised the exception.
	SPSR uint64

	// Floating-point registers
	VFP  [8][8]uint64 // v0-v31
	FPSR uint64
	FPCR uint64

	// Memory is the executable allocated RAM
	Memory *dma.Region

	// TODO: use ASID

	// MMU, if not nil, is called before each execution context Schedule()
	// or Write() to allow virtual addressing re-configuration as needed
	// (see arm64.ConfigureMMU).
	MMU func()

	// Handler, if not nil, handles context switch calls
	Handler func(ctx *ExecCtx) error

	// Server, if not nil, serves RPC calls over syscalls
	Server *rpc.Server

	// Shadow represents a redundant execution context for opportunistic
	// soft lockstep, it is meant to be created with Clone() and once set
	// enables its delayed lockstep execution for fault detection.
	//
	// When set Run() will Schedule() the primary and Shadow context for
	// opportunistic comparison, in case of a mismatch (see Equal) the
	// primary context Run() raises an error.
	Shadow *ExecCtx

	// execution state
	run bool
	// stopped will be closed once the context has stopped running.
	stopped chan struct{}
	// TrustZone configuration
	ns bool
	// executing g stack pointer
	g_sp uint64
	// executing g registers saved by Exec
	host [6]uint64

	// Read() buffer
	in []byte
	// Write() buffer
	out []byte
}

// String returns the string form of the execution context registers.
func (ctx *ExecCtx) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\n")
	fmt.Fprintf(&sb, "   r0:%.16x  r1:%.16x  r2:%.16x  r3:%.16x\n", ctx.R0, ctx.R1, ctx.R2, ctx.R3)
	fmt.Fprintf(&sb, "   r4:%.16x  r5:%.16x  r6:%.16x  r7:%.16x\n", ctx.R4, ctx.R5, ctx.R6, ctx.R7)
	fmt.Fprintf(&sb, "   r8:%.16x  r9:%.16x r10:%.16x r11:%.16x\n", ctx.R8, ctx.R9, ctx.R10, ctx.R11)
	fmt.Fprintf(&sb, "  r12:%.16x r13:%.16x r14:%.16x r15:%.16x\n", ctx.R12, ctx.R13, ctx.R14, ctx.R15)
	fmt.Fprintf(&sb, "  r16:%.16x r17:%.16x r18:%.16x r19:%.16x\n", ctx.R16, ctx.R17, ctx.R18, ctx.R19)
	fmt.Fprintf(&sb, "  r20:%.16x r21:%.16x r22:%.16x r23:%.16x\n", ctx.R20, ctx.R21, ctx.R22, ctx.R23)
	fmt.Fprintf(&sb, "  r24:%.16x r25:%.16x r26:%.16x r27:%.16x spsr:%.16x\n", ctx.R24, ctx.R25, ctx.R26, ctx.R27, ctx.SPSR)
	fmt.Fprintf(&sb, "  r28:%.16x  fp:%.16x  lr:%.16x  sp:%.16x   pc:%.16x\n", ctx.R28, ctx.R29, ctx.R30, ctx.R31, ctx.PC)

	return sb.String()
}

// Schedule runs the execution context until an exception is caught.
//
// Unlike Run() the function does not invoke the context Handler(), therefore
// exceptions and system or monitor calls are not handled.
func (ctx *ExecCtx) Schedule() (err error) {
	mux.Lock()
	defer mux.Unlock()

	// reconfigure MMU as needed
	if ctx.MMU != nil {
		ctx.MMU()
	}

	// execute context
	Exec(ctx)

	ec := (ctx.ESR >> arm64.ESR_EC) & 0x3f

	switch ec {
	case arm64.EC_UNKNOWN, arm64.EC_SYSREG:
		// left to application handler
		return
	case arm64.EC_SERROR:
		return fmt.Errorf("SError esr:%#x", ctx.ESR)
	}

	if ec != arm64.EC_SVC64 {
		return fmt.Errorf("exception ec:%#x esr:%#x far:%#x", ec, ctx.ESR, ctx.FAR)
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

	for ctx.run {
		if err = ctx.Schedule(); err != nil {
			break
		}

		if ctx.Shadow != nil {
			err = ctx.Shadow.lockstep(ctx)
			ctx.MMU()

			if err != nil {
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
//
// The caller is responsible for any other required MMU configuration (see
// arm64.ConfigureMMU()).
func Load(entry uint, mem *dma.Region, secure bool) (ctx *ExecCtx, err error) {
	if !secure {
		return nil, errors.New("unimplemented")
	}

	ctx = &ExecCtx{
		PC:     uint64(entry),
		Memory: mem,
		Server: rpc.NewServer(),
		ns:     !secure,
	}

	if secure {
		ctx.Handler = SecureHandler
	} else {
		ctx.Handler = NonSecureHandler
	}

	// set all mask bits
	ctx.SPSR = (0b111 << 6)

	return
}

// Equal returns whether a and b holds the same register state.
func Equal(a, b *ExecCtx) bool {
	return (a.R0 == b.R0 &&
		a.R1 == b.R1 &&
		a.R2 == b.R2 &&
		a.R3 == b.R3 &&
		a.R4 == b.R4 &&
		a.R5 == b.R5 &&
		a.R6 == b.R6 &&
		a.R7 == b.R7 &&
		a.R8 == b.R8 &&
		a.R9 == b.R9 &&
		a.R10 == b.R10 &&
		a.R11 == b.R11 &&
		a.R12 == b.R12 &&
		a.R13 == b.R13 &&
		a.R14 == b.R14 &&
		a.R15 == b.R15 &&
		a.R16 == b.R16 &&
		a.R17 == b.R17 &&
		a.R18 == b.R18 &&
		a.R19 == b.R19 &&
		a.R20 == b.R20 &&
		a.R21 == b.R21 &&
		a.R22 == b.R22 &&
		a.R23 == b.R23 &&
		a.R24 == b.R24 &&
		a.R25 == b.R25 &&
		a.R26 == b.R26 &&
		a.R27 == b.R27 &&
		a.R28 == b.R28 &&
		a.R29 == b.R29 &&
		a.R30 == b.R30 &&
		a.R31 == b.R31 &&
		a.PC == b.PC &&
		a.SPSR == b.SPSR &&
		a.FPSR == b.FPSR &&
		a.FPCR == b.FPCR &&
		a.VFP == b.VFP &&
		slices.Equal(a.in, b.in))
}
