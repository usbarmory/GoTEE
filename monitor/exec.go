// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package monitor provides supervisor support for TamaGo unikernels to allow
// scheduling of Secure user mode or NonSecure system mode executables.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package monitor

import (
	"errors"
	"log"
	"net/rpc"
	"runtime"
	"sync"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/csu"
	"github.com/f-secure-foundry/tamago/soc/imx6/tzasc"
)

const (
	UserMode   = (0b111 << 6) | arm.USR_MODE
	SystemMode = (0b111 << 6) | arm.SYS_MODE
)

var systemVectorTable = arm.SystemVectorTable()

var monitorVectorTable = arm.VectorTable{
	Reset:         resetMonitor,
	Undefined:     undefinedMonitor,
	Supervisor:    supervisorMonitor,
	PrefetchAbort: prefetchAbortMonitor,
	DataAbort:     dataAbortMonitor,
	IRQ:           irqMonitor,
	FIQ:           fiqMonitor,
}

var mux sync.Mutex

// defined in exec.s
func resetMonitor()
func undefinedMonitor()
func supervisorMonitor()
func prefetchAbortMonitor()
func dataAbortMonitor()
func irqMonitor()
func fiqMonitor()

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
	R0   uint32
	R1   uint32
	R2   uint32
	R3   uint32
	R4   uint32
	R5   uint32
	R6   uint32
	R7   uint32
	R8   uint32
	R9   uint32
	R10  uint32
	R11  uint32
	R12  uint32
	R13  uint32 // SP
	R14  uint32 // LR
	R15  uint32 // PC
	SPSR uint32
	CPSR uint32

	ExceptionVector int

	VFP   []uint64 // d0-d31
	FPSCR uint32
	FPEXC uint32

	// Memory is the executable allocated RAM
	Memory *dma.Region

	// Handler, if not nil, handles user syscalls
	Handler func(ctx *ExecCtx) error

	// Server, if not nil, serves RPC calls over syscalls
	Server *rpc.Server

	// Debug controls activation of debug logs
	Debug bool

	// TrustZone configuration
	ns bool

	// executing g stack pointer
	g_sp uint32

	// Write() buffer
	buf []byte
}

// Print logs the execution context user registers.
func (ctx *ExecCtx) Print() {
	cpsr, spsr := ctx.Mode()

	log.Printf("   r0:%.8x  r1:%.8x  r2:%.8x  r3:%.8x", ctx.R0, ctx.R1, ctx.R2, ctx.R3)
	log.Printf("   r4:%.8x  r5:%.8x  r6:%.8x  r7:%.8x", ctx.R4, ctx.R5, ctx.R6, ctx.R7)
	log.Printf("   r8:%.8x  r9:%.8x r10:%.8x r11:%.8x cpsr:%.8x (%s)", ctx.R8, ctx.R9, ctx.R10, ctx.R11, ctx.CPSR, arm.ModeName(cpsr))
	log.Printf("  r12:%.8x  sp:%.8x  lr:%.8x  pc:%.8x spsr:%.8x (%s)", ctx.R12, ctx.R13, ctx.R14, ctx.R15, ctx.SPSR, arm.ModeName(spsr))
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

func (ctx *ExecCtx) schedule() (err error) {
	mux.Lock()
	defer mux.Unlock()

	// set monitor handlers
	arm.SetVectorTable(monitorVectorTable)

	// execute applet
	Exec(ctx)

	// restore default handlers
	arm.SetVectorTable(systemVectorTable)

	mode, _ := ctx.Mode()

	switch mode {
	case arm.MON_MODE, arm.SVC_MODE:
		return
	default:
		if ctx.Debug {
			ctx.Print()
		}

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
	for {
		if err = ctx.schedule(); err != nil {
			break
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

// Load returns an execution context initialized for the argument entry point
// and memory region, the secure flag controls whether the context belongs to a
// secure partition (e.g. TrustZone Secure World) or a non-secure one (e.g.
// TrustZone Normal World).
//
// In case of a non-secure execution context, the memory is configured as
// NonSecure by means of MMU NS bit and memory controller region configuration.
// Any additional peripheral restrictions are up to the caller.
func Load(entry uint32, mem *dma.Region, secure bool) (ctx *ExecCtx, err error) {
	ctx = &ExecCtx{
		R15:    entry,
		VFP:    make([]uint64, 32),
		Memory: mem,
		Server: rpc.NewServer(),
		ns:     !secure,
	}

	if secure {
		ctx.Handler = SecureHandler
	} else {
		ctx.Handler = NonSecureHandler
	}

	memAttr := arm.TTE_CACHEABLE | arm.TTE_BUFFERABLE | arm.TTE_SECTION

	// redundant enforcement of Region 0 (entire memory space) defaults
	if tzasc.EnableRegion(0, 0, 0, (1<<tzasc.SP_SW_RD)|(1<<tzasc.SP_SW_WR)); err != nil {
		return
	}

	if ctx.ns && imx6.Native {
		// allow NonSecure World R/W access to its own memory
		if err = tzasc.EnableRegion(1, mem.Start, mem.Size, (1<<tzasc.SP_NW_RD)|(1<<tzasc.SP_NW_WR)); err != nil {
			return
		}
	}

	if ctx.ns {
		// The NS bit is required to ensure that cache lines are kept
		// separate.
		memAttr |= arm.TTE_AP_001<<10 | arm.TTE_NS
		ctx.SPSR = SystemMode
	} else {
		memAttr |= arm.TTE_AP_011 << 10
		ctx.SPSR = UserMode
	}

	// Cortex-A7 master needs CP15SDISABLE low for arm.set_ttbr0
	if secure, lock, err := csu.GetAccess(0); !secure && !lock && err == nil {
		csu.SetAccess(0, true, false)
		defer csu.SetAccess(0, false, false)
	}

	imx6.ARM.ConfigureMMU(mem.Start, mem.Start+uint32(mem.Size), memAttr)

	return
}
