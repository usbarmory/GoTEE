// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package monitor provides supervisor support for TamaGo unikernels to allow
// scheduling of Secure user mode or NonSecure system mode executables.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
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

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/arm/tzc380"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

var (
	systemVectorTable = arm.SystemVectorTable()

	monitorVectorTable = arm.VectorTable{
		Reset:         resetMonitor,
		Undefined:     undefinedMonitor,
		Supervisor:    supervisorMonitor,
		PrefetchAbort: prefetchAbortMonitor,
		DataAbort:     dataAbortMonitor,
		IRQ:           irqMonitor,
		FIQ:           fiqMonitor,
	}
)

var mux sync.Mutex

// defined in exec_arm.s
func resetMonitor()
func undefinedMonitor()
func supervisorMonitor()
func prefetchAbortMonitor()
func dataAbortMonitor()
func irqMonitor()
func fiqMonitor()

func init() {
	imx6ul.CSU.Init()
	imx6ul.TZASC.Init()

	if !imx6ul.Native {
		return
	}

	tzcAttr := (1 << tzc380.SP_SW_RD) | (1 << tzc380.SP_SW_WR)

	// redundant enforcement of Region 0 (entire memory space) defaults
	if err := imx6ul.TZASC.EnableRegion(0, 0, 0, tzcAttr); err != nil {
		panic("could not set TZASC defaults")
	}
}

// Exec allows execution of an executable in Secure user mode or NonSecure
// system mode. The execution is isolated from the invoking Go runtime,
// yielding back to it is supported through exceptions (e.g. syscalls through
// SVC).
//
// The execution context pointer allows task initialization and it is updated
// with the program state at return, it can therefore be passed again to resume
// the task.
func Exec(ctx *ExecCtx)

// ExecCtx represents a executable initialization or returning state.
type ExecCtx struct {
	R0  uint32
	R1  uint32
	R2  uint32
	R3  uint32
	R4  uint32
	R5  uint32
	R6  uint32
	R7  uint32
	R8  uint32
	R9  uint32
	R10 uint32
	R11 uint32
	R12 uint32
	R13 uint32 // SP
	R14 uint32 // LR
	R15 uint32 // PC

	// CPSR is the Current Program Status Register of the handler which
	// caught the exception raised by the execution context.
	CPSR uint32

	// SPSR (Saved Program Status Register) is the CPSR of the execution
	// context as it raised the exception.
	SPSR uint32

	ExceptionVector int

	VFP   []uint64 // d0-d31
	FPSCR uint32
	FPEXC uint32

	// Memory is the executable allocated RAM
	Memory *dma.Region

	// Domain represents the domain ID (0-15) assigned to the executable
	// Memory. The value must be overridden with distinct values if memory
	// isolation is required for parallel execution of different
	// user/secure execution contexts.
	Domain uint32

	// MMU, if not nil, is called before each execution context Schedule()
	// or Write() to allow virtual addressing re-configuration as needed
	// (see arm.ConfigureMMU).
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
	g_sp uint32

	// Read() buffer
	in []byte
	// Write() buffer
	out []byte
}

// String returns the string form of the execution context registers.
func (ctx *ExecCtx) String() string {
	var sb strings.Builder

	cpsr, spsr := ctx.Mode()

	fmt.Fprintf(&sb, "\n")
	fmt.Fprintf(&sb, "   r0:%.8x  r1:%.8x  r2:%.8x  r3:%.8x\n", ctx.R0, ctx.R1, ctx.R2, ctx.R3)
	fmt.Fprintf(&sb, "   r4:%.8x  r5:%.8x  r6:%.8x  r7:%.8x\n", ctx.R4, ctx.R5, ctx.R6, ctx.R7)
	fmt.Fprintf(&sb, "   r8:%.8x  r9:%.8x r10:%.8x r11:%.8x cpsr:%.8x (%s)\n", ctx.R8, ctx.R9, ctx.R10, ctx.R11, ctx.CPSR, arm.ModeName(cpsr))
	fmt.Fprintf(&sb, "  r12:%.8x  sp:%.8x  lr:%.8x  pc:%.8x spsr:%.8x (%s)\n", ctx.R12, ctx.R13, ctx.R14, ctx.R15, ctx.SPSR, arm.ModeName(spsr))

	return sb.String()
}

// NonSecure returns whether the execution context is loaded as non-secure.
func (ctx *ExecCtx) NonSecure() bool {
	return ctx.ns
}

// Mode returns the processor mode.
func (ctx *ExecCtx) Mode() (current int, saved int) {
	current = int(ctx.CPSR & 0x1f)
	saved = int(ctx.SPSR & 0x1f)
	return
}

// Schedule runs the execution context until an exception is caught.
//
// Unlike Run() the function does not invoke the context Handler(), there
// exceptions and system or monitor calls are not handled.
func (ctx *ExecCtx) Schedule() (err error) {
	mux.Lock()
	defer mux.Unlock()

	// set monitor handlers
	imx6ul.ARM.SetVectorTable(monitorVectorTable)

	// reconfigure MMU as needed
	if ctx.MMU != nil {
		ctx.MMU()
	}

	// execute context
	Exec(ctx)

	// restore default handlers
	imx6ul.ARM.SetVectorTable(systemVectorTable)

	mode, _ := ctx.Mode()

	switch mode {
	case arm.IRQ_MODE, arm.FIQ_MODE, arm.SVC_MODE, arm.MON_MODE:
		return
	default:
		return errors.New(arm.ModeName(mode))
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

	ap := arm.TTE_AP_001

	if !ctx.ns {
		ap = arm.TTE_AP_011
	}

	// Set privilege level and domain access
	imx6ul.ARM.SetAccessPermissions(
		uint32(ctx.Memory.Start()), uint32(ctx.Memory.End()),
		ap, ctx.Domain,
	)

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

		// Return to next instruction when handling interrupts
		// (Table 11-3, ARM® Cortex™ -A Series Programmer’s Guide).
		switch ctx.ExceptionVector {
		case arm.IRQ, arm.FIQ:
			ctx.R15 -= 4
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
// and memory region, the secure flag controls whether the context belongs to a
// secure partition (e.g. TrustZone Secure World) or a non-secure one (e.g.
// TrustZone Normal World).
//
// In case of a non-secure execution context, the memory is configured as
// NonSecure by means of MMU NS bit and memory controller region configuration.
//
// The caller is responsible for any other required MMU configuration (see
// arm.ConfigureMMU()) or additional peripheral restrictions (e.g. TrustZone).
func Load(entry uint, mem *dma.Region, secure bool) (ctx *ExecCtx, err error) {
	ctx = &ExecCtx{
		R15:    uint32(entry),
		VFP:    make([]uint64, 32),
		Memory: mem,
		Server: rpc.NewServer(),
		ns:     !secure,
	}

	if secure {
		// enable VFP
		ctx.FPEXC = 1 << arm.FPEXC_EN
	}

	if secure {
		ctx.Handler = SecureHandler
	} else {
		ctx.Handler = NonSecureHandler
	}

	tzcAttr := (1 << tzc380.SP_NW_RD) | (1 << tzc380.SP_NW_WR)

	if ctx.ns && imx6ul.Native {
		// allow NonSecure World R/W access to its own memory
		if err = imx6ul.TZASC.EnableRegion(1, uint32(mem.Start()), uint32(mem.Size()), tzcAttr); err != nil {
			return
		}
	}

	// set all mask bits
	ctx.SPSR = (0b111 << 6)
	flags := arm.MemoryRegion

	if ctx.ns {
		// The NS bit is required to ensure that cache lines are kept
		// separate.
		ctx.SPSR |= arm.SYS_MODE
		flags |= arm.TTE_NS
	} else {
		ctx.SPSR |= arm.USR_MODE
	}

	imx6ul.ARM.SetAttributes(uint32(mem.Start()), uint32(mem.End()), flags)

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
		a.CPSR == b.CPSR &&
		a.SPSR == b.SPSR &&
		a.FPSCR == b.FPSCR &&
		a.FPEXC == b.FPEXC &&
		slices.Equal(a.VFP, b.VFP) &&
		slices.Equal(a.in, b.in))
}
