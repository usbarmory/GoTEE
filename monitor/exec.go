// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"fmt"
	"log"
	"runtime"

	"github.com/f-secure-foundry/GoTEE"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

const userModeFlags = arm.TTE_AP_011 | arm.TTE_CACHEABLE | arm.TTE_BUFFERABLE | arm.TTE_SECTION_1MB
const userModePSR uint32 = (0b111 << 6) | arm.USR_MODE

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

	// Memory is the executable allocated RAM
	Memory *dma.Region

	// Handler, if not nil, handles user syscalls
	Handler func(ctx *ExecCtx) error

	// Debug controls activation of debug logs
	Debug bool

	// executing g stack pointer
	g_sp uint32
}

func (ctx *ExecCtx) debug() {
	log.Printf("\tr0:%.8x   r1:%.8x   r2:%.8x   r3:%.8x", ctx.R0, ctx.R1, ctx.R2, ctx.R3)
	log.Printf("\tr1:%.8x   r2:%.8x   r3:%.8x   r4:%.8x", ctx.R1, ctx.R2, ctx.R3, ctx.R4)
	log.Printf("\tr5:%.8x   r6:%.8x   r7:%.8x   r8:%.8x", ctx.R5, ctx.R6, ctx.R7, ctx.R8)
	log.Printf("\tr9:%.8x  r10:%.8x  r11:%.8x  r12:%.8x", ctx.R9, ctx.R10, ctx.R11, ctx.R12)
	log.Printf("\tsp:%.8x   lr:%.8x   pc:%.8x spsr:%.8x", ctx.R13, ctx.R14, ctx.R15, ctx.SPSR)
}

func (ctx *ExecCtx) schedule() (err error) {
	// set monitor handlers
	arm.SetVectorTable(monitorVectorTable)

	// execute applet
	Exec(ctx)

	// restore default handlers
	arm.SetVectorTable(systemVectorTable)

	if mode := ctx.ExceptionMode(); mode != arm.SVC_MODE {
		if ctx.Debug {
			ctx.debug()
		}

		return fmt.Errorf("exception mode %s", arm.ModeName(mode))
	}

	return
}

// Mode returns the processor mode.
func (ctx *ExecCtx) ExceptionMode() int {
	return int(ctx.CPSR & 0x1f)
}

// Run starts the execution context and handles user mode system calls. The
// function yields to the invoking Go runtime only when exceptions are issued
// in user mode.
//
// The function handles system calls (see Handler()) and only returns if
// non-supervisor exceptions are issued or when `SYS_EXIT` system call is
// handled.
func (ctx *ExecCtx) Run() (err error) {
	if ctx.Debug {
		log.Printf("PL1 starting PL0 sp:%#.8x pc:%#.8x", ctx.R13, ctx.R15)
	}

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

	if ctx.Debug {
		log.Printf("PL1 stopped PL0 task sp:%#.8x lr:%#.8x pc:%#.8x err:%v", ctx.R13, ctx.R14, ctx.R15, err)
	}

	return
}

// Load returns an execution context initialized for the argument ELF
// executable.
func Load(elf []byte) *ExecCtx {
	imx6.ARM.ConfigureMMU(tee.AppletStart, tee.AppletStart+uint32(tee.AppletSize), userModeFlags)

	mem := &dma.Region{
		Start: tee.AppletStart,
		Size:  tee.AppletSize,
	}

	mem.Init()
	mem.Reserve(mem.Size, 0)

	return &ExecCtx{
		R13: mem.Start + uint32(mem.Size) - tee.AppletStackOffset,
		R15: parseELF(mem, elf),
		SPSR: userModePSR,
		Memory: mem,
	}
}
