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

// Exec allows execution of a bare metal executable in user mode. The execution
// is isolated from the invoking Go unikernel runtime, which can be returned to
// with a a supervisor (SVC) exception or through timer interrupts (TODO).
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
	R13  uint32
	R14  uint32
	R15  uint32
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

// Mode returns the processor mode.
func (ctx *ExecCtx) ExceptionMode() int {
	return int(ctx.CPSR & 0x1f)
}

func (ctx *ExecCtx) schedule() (err error) {
	// set monitor handlers
	arm.SetVectorTable(MonitorVectorTable)

	// execute applet
	Exec(ctx)

	// restore default handlers
	arm.SetVectorTable(systemVectorTable)

	if mode := ctx.ExceptionMode(); mode != arm.SVC_MODE {
		return fmt.Errorf("exception mode %s", arm.ModeName(mode))
	}

	return
}

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
		log.Printf("stopped PL0 task sp:%#.8x lr:%#.8x pc:%#.8x err:%v", ctx.R13, ctx.R14, ctx.R15, err)
	}

	return
}

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
