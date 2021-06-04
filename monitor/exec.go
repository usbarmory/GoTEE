// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package monitor provides supervisor support for TamaGo unikernels
// to allow scheduling of user mode executables.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package monitor

import (
	"fmt"
	"log"
	"net/rpc"
	"runtime"
	"sync"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

const (
	UserMode   = (0b111 << 6) | arm.USR_MODE
	SystemMode = (0b111 << 6) | arm.SYS_MODE
)

var systemVectorTable = arm.SystemVectorTable()

var monitorVectorTable = arm.VectorTable{
	Reset:         monitor,
	Undefined:     monitor,
	Supervisor:    monitor,
	PrefetchAbort: monitor,
	DataAbort:     monitor,
	IRQ:           monitor,
	FIQ:           monitor,
}

var mux sync.Mutex

// defined in exec.s
func monitor()

// Exec allows execution of an executable in user mode. The execution is
// isolated from the invoking Go runtime as user mode can yield back to it
// through exceptions (e.g. syscalls through SVC).
//
// The execution context pointer allows task initialization and it is updated
// with the user mode program state at return, it can therefore be passed again
// to resume the task.
func Exec(ctx *ExecCtx)

// ExecCtx represents a user mode executable initialization or returning state.
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

	VFP   []uint64 // d0-d31
	FPSCR uint32

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
	log.Printf("\t r0:%.8x   r1:%.8x   r2:%.8x   r3:%.8x", ctx.R0, ctx.R1, ctx.R2, ctx.R3)
	log.Printf("\t r4:%.8x   r5:%.8x   r6:%.8x   r7:%.8x", ctx.R4, ctx.R5, ctx.R6, ctx.R7)
	log.Printf("\t r8:%.8x   r9:%.8x  r10:%.8x  r11:%.8x", ctx.R8, ctx.R9, ctx.R10, ctx.R11)
	log.Printf("\tr12:%.8x   sp:%.8x   lr:%.8x   pc:%.8x  spsr:%.8x", ctx.R12, ctx.R13, ctx.R14, ctx.R15, ctx.SPSR)
}

// NonSecure returns whether the execution context is loaded as non-secure.
func (ctx *ExecCtx) NonSecure() bool {
	return ctx.ns
}

// Mode returns the processor mode.
func (ctx *ExecCtx) ExceptionMode() int {
	return int(ctx.CPSR & 0x1f)
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

	mode := ctx.ExceptionMode()

	switch mode {
	case arm.MON_MODE, arm.SVC_MODE:
		return
	default:
		if ctx.Debug {
			ctx.Print()
		}

		return fmt.Errorf("exception mode %s", arm.ModeName(mode))
	}

	return
}

// Run starts the execution context and handles user mode system calls. The
// function yields to the invoking Go runtime only when exceptions are issued
// in user mode.
//
// The function handles system calls (see Handler()) and only returns if
// non-supervisor exceptions are issued or when `SYS_EXIT` system call is
// handled.
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

// Load returns an execution context initialized for the argument ELF
// executable, the secure flag controls whether the context belongs to a secure
// partition (e.g. TrustZone Secure World) or not (e.g. TrustZone Normal
// World).
func Load(elf []byte, start uint32, size int, secure bool) (ctx *ExecCtx, err error) {
	mem := &dma.Region{
		Start: start,
		Size:  size,
	}

	mem.Init()
	mem.Reserve(mem.Size, 0)

	entry, err := parseELF(mem, elf)

	if err != nil {
		return
	}

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

	if ctx.ns {
		// The NS bit is required to ensure that cache lines are kept
		// separate.
		memAttr |= arm.TTE_AP_001<<10 | arm.TTE_NS
		ctx.SPSR = SystemMode
	} else {
		memAttr |= arm.TTE_AP_011 << 10
		ctx.SPSR = UserMode
	}

	imx6.ARM.ConfigureMMU(start, start+uint32(size), memAttr)

	return
}
