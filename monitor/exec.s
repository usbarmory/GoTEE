// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// GoTEE exception handling relies on one exit point (Exec) and one return
// point (monitor), both are used for:
//
//   Secure      User Mode execution and exception handling
//   NonSecure System mode execution and supervisor call handling
//
// An execution context (ExecCtx) structure is used to hold initial register
// state at execution as well as store the updated state on re-entry.
//
// The exception handling uses ARM Thread ID Register (TPIDRURO) in a similar
// manner to Go own use of TLS (https://golang.org/src/runtime/tls_arm.s).
//
// With respect to TrustZone the handler theoretically must save and restore
// the following registers between Secure <> NonSecure World switches:
//
//  • r0-r15, CPSR of System/User/Supervisor modes:
//
//    TamaGo (and therefore GoTEE) does not use Supervisor mode, System/User
//    mode share the same register banks, therefore saving and restoring the
//    registers of the mode that triggered the exception is sufficient.
//
//  • r13-r14 of Abort/Undefined/IRQ modes, r8-r14 of FIQ mode:
//
//    In GoTEE exceptions to Abort/Undefined/IRQ/FIQ modes are always handled
//    with an unrecoverable panic, therefore we do not save/restore their
//    banked registers.
//
//  • TODO: Data register of shared coprocessors (e.g VFP/FPU), example:
//      vstm  rN!, {d0-d15}
//      vstm  rN!, {d16-d31}

// func Exec(ctx *ExecCtx)
TEXT ·Exec(SB),$0-4
	// save caller registers
	MOVM.DB.W	[R0-R12, R14], (R13)	// push {r0-r12, r14}

	// get argument pointer
	ADD	$(14*4), R13, R13
	MOVW	ctx+0(FP), R0
	SUB	$(14*4), R13, R13

	// save g stack pointer
	MOVW	R13, ExecCtx_g_sp(R0)

	// restore SP, LR
	MOVW	ExecCtx_R13(R0), R13
	MOVW	ExecCtx_R14(R0), R14

	// save context pointer as Thread ID (TPIDRURO)
	MCR	15, 0, R0, C13, C0, 3

	// switch to monitor mode
	WORD	$0xf1020016			// cps 0x16

	// restore mode
	MOVW	ExecCtx_SPSR(R0), R1
	WORD	$0xe169f001			// msr SPSR, r1

	MOVW	ExecCtx_ns(R0), R1
	TST	$1, R1
	BEQ	switch

	// enable EA, FIQ, and NS bit in SCR
	MOVW	$13, R1
	MCR	15, 0, R1, C1, C1, 0

switch:
	// restore r0-r12, r15
	WORD	$0xe8d0ffff			// ldmia r0, {r0-r15}^

#define MONITOR_EXCEPTION()							\
	/* save R0 */								\
	MOVW	R0, R13								\
										\
	/* disable NS bit in SCR */						\
	MOVW	$0, R0								\
	MCR	15, 0, R0, C1, C1, 0						\
										\
	/* restore context pointer from Thread ID (TPIDRURO) */			\
	MRC	15, 0, R0, C13, C0, 3						\
										\
	/* save caller registers */						\
	WORD	$0xe8c07fff			/* stmia r0, {r0-r14}^ */	\
	MOVW	R0, R1								\
	MOVW	R13, ExecCtx_R0(R1)						\
										\
	/* save PC from LR */							\
	MOVW	R14, ExecCtx_R15(R1)						\
										\
	WORD	$0xe14f0000			/* mrs r0, SPSR */		\
	MOVW	R0, ExecCtx_SPSR(R1)						\
										\
	WORD	$0xe10f0000			/* mrs r0, CPSR */		\
	MOVW	R0, ExecCtx_CPSR(R1)						\
										\
	/* switch to System Mode */						\
	MOVW	$0x1df, R0			/* AIF masked, SYS mode */	\
	WORD	$0xe169f000			/* msr SPSR, R0 */		\
	WORD	$0xf102001f			/* cps 0x1f */			\
										\
	/* restore g registers */						\
	MOVW		ExecCtx_g_sp(R1), R13					\
	MOVM.IA.W	(R13), [R0-R12, R14]	/* pop {r0-rN, r14} */		\
										\
	/* restore PC from LR */						\
	MOVW	R14, R15							\

TEXT ·monitor(SB),NOSPLIT|NOFRAME,$0
	MONITOR_EXCEPTION()
