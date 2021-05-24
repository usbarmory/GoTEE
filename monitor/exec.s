// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

DATA	·lastCtx+0(SB)/4,$0
GLOBL	·lastCtx+0(SB),RODATA,$4

// func Exec(ctx *ExecCtx)
TEXT ·Exec(SB),$0-4
	// save caller registers
	MOVM.DB.W	[R0-R12, R14], (R13)	// push {r0-r12, r14}

	// get argument pointer
	ADD	$(14*4), R13, R13
	MOVW	ctx+0(FP), R0
	MOVW	R0, ·lastCtx(SB)
	SUB	$(14*4), R13, R13

	// save g stack pointer
	MOVW	R13, ExecCtx_g_sp(R0)

	// restore SP, LR
	MOVW	ExecCtx_R13(R0), R13
	MOVW	ExecCtx_R14(R0), R14

	// switch to supervisor mode
	WORD	$0xf1020013			// cps 0x13

	MOVW	ExecCtx_NonSecure(R0), R1
	AND	$1, R1, R1
	CMP	$1, R1
	BNE	restore

	// enable EA, FIQ, and NS bit in SCR
	MOVW	$13, R1
	MCR	15, 0, R1, C1, C1, 0

restore:
	// restore mode
	MOVW	ExecCtx_SPSR(R0), R1
	WORD	$0xe169f001			// msr SPSR, r1

	// restore r0-r12, r15
	WORD	$0xe8d0ffff			// ldmia r0, {r0-r15}^

#define MONITOR_EXCEPTION()							\
	/* save caller registers */						\
										\
	MOVW	·lastCtx(SB), R13						\
	WORD	$0xe8cd7fff			/* stmia r13, {r0-r14}^ */	\
	MOVW	R14, ExecCtx_R15(R13)						\
										\
	WORD	$0xe14f0000			/* mrs r0, SPSR */		\
	MOVW	R0, ExecCtx_SPSR(R13)						\
										\
	WORD	$0xe10f0000			/* mrs r0, CPSR */		\
	MOVW	R0, ExecCtx_CPSR(R13)						\
										\
	/* disable NS bit in SCR */						\
	MOVW	$0, R0								\
	MCR	15, 0, R0, C1, C1, 0						\
										\
	/* switch to System Mode */						\
	WORD	$0xf102001f			/* cps 0x1f */			\
										\
	/* restore g registers */						\
	MOVW		·lastCtx(SB), R13					\
	MOVW		ExecCtx_g_sp(R13), R13					\
	MOVM.IA.W	(R13), [R0-R12, R14]	/* pop {r0-rN, r14} */		\
										\
	/* restore PC from LR */						\
	MOVW	R14, R15							\

TEXT ·monitor(SB),NOSPLIT|NOFRAME,$0
	MONITOR_EXCEPTION()
