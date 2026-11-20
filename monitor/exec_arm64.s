// Copyright (c) The GoTEE authors. All Rights Reserved.
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
// The exception handling uses a Thread ID Register (TPIDR_EL1) in a similar
// manner to Go own use of TLS (https://golang.org/src/runtime/tls_arm64.s).
//
// The handler must save and restore the following registers between context
// switches:
//
//  • General purpose registers:
//
//    The r0-r31 registers of the caller mode are saved/restored.
//
//  • Data register of shared coprocessors (e.g VFP/FPU):
//
//    The v0-v31, FPSR and FPCR registers are saved/restored.

// func Exec(ctx *ExecCtx)
TEXT ·Exec(SB),$0-8
	// get argument pointer
	MOVD	ctx+0(FP), R0

	// R28 R29 R30 FPCR

	// save caller registers
	MOVD	FPCR, R1
	STP	(g, R1), (0*16)+ExecCtx_host(R0)
	STP	(R29, R30), (1*16)+ExecCtx_host(R0)
	MOVD	DAIF, R1
	MOVD	R1, (2*16)+ExecCtx_host(R0)

	// mask all interrupts
	MSR	$0b1111, DAIFSet

	// clear last exception
	MOVD	$0, R1
	MSR	R1, FAR_EL1
	MSR	R1, ESR_EL1

	// save g stack pointer
	MOVD	RSP, R1
	MOVD	R1, ExecCtx_g_sp(R0)

	// restore SP
	MOVD	ExecCtx_R31(R0), R1
	MSR	R1, SP_EL0

	// save context pointer as Thread ID (TPIDR_EL1)
	MSR	R0, TPIDR_EL1

	// save mode
	MOVD	ExecCtx_SPSR(R0), R1
	MOVD	$0xf00003c0, R2
	AND	R2, R1, R1
	MSR	R1, SPSR_EL1

	// restore VFP registers
	ADD	$ExecCtx_VFP, R0, R1
	VLD1.P	64(R1), [V0.B16, V1.B16, V2.B16, V3.B16]
	VLD1.P	64(R1), [V4.B16, V5.B16, V6.B16, V7.B16]
	VLD1.P	64(R1), [V8.B16, V9.B16, V10.B16, V11.B16]
	VLD1.P	64(R1), [V12.B16, V13.B16, V14.B16, V15.B16]
	VLD1.P	64(R1), [V16.B16, V17.B16, V18.B16, V19.B16]
	VLD1.P	64(R1), [V20.B16, V21.B16, V22.B16, V23.B16]
	VLD1.P	64(R1), [V24.B16, V25.B16, V26.B16, V27.B16]
	VLD1.P	64(R1), [V28.B16, V29.B16, V30.B16, V31.B16]
	MOVD	ExecCtx_FPSR(R0), R1
	MOVD	R1, FPSR
	MOVD	ExecCtx_FPCR(R0), R1
	MOVD	R1, FPCR

	// restore general purpose registers
	LDP	ExecCtx_R2(R0), (R2, R3)
	LDP	ExecCtx_R4(R0), (R4, R5)
	LDP	ExecCtx_R6(R0), (R6, R7)
	LDP	ExecCtx_R8(R0), (R8, R9)
	LDP	ExecCtx_R10(R0), (R10, R11)
	LDP	ExecCtx_R12(R0), (R12, R13)
	LDP	ExecCtx_R14(R0), (R14, R15)
	LDP	ExecCtx_R16(R0), (R16, R17)
	LDP	ExecCtx_R18(R0), (R18_PLATFORM, R19)
	LDP	ExecCtx_R20(R0), (R20, R21)
	LDP	ExecCtx_R22(R0), (R22, R23)
	LDP	ExecCtx_R24(R0), (R24, R25)
	LDP	ExecCtx_R26(R0), (R26, R27)
	LDP	ExecCtx_R28(R0), (g, R29)
	MOVD	ExecCtx_R30(R0), R30

	// restore PC
	MOVD	ExecCtx_PC(R0), R1
	MSR	R1, ELR_EL1

	// restore R1, R0
	LDP	ExecCtx_R0(R0), (R0, R1)

	ERET

TEXT ·monitor(SB),NOSPLIT|NOFRAME,$0
	// restore context pointer from Thread ID (TPIDR_EL1)
	MRS	TPIDR_EL1, R18_PLATFORM

	// save general purpose registers
	STP	(R0, R1), ExecCtx_R0(R18_PLATFORM)
	MOVD	R18_PLATFORM, R0

	// save R18, pushed by the trampoline (see arm64.SetLowerVectorTable)
	MOVD	(RSP), R1
	MOVD	R1, ExecCtx_R18(R0)
	ADD	$16, RSP

	STP	(R2, R3), ExecCtx_R2(R0)
	STP	(R4, R5), ExecCtx_R4(R0)
	STP	(R6, R7), ExecCtx_R6(R0)
	STP	(R8, R9), ExecCtx_R8(R0)
	STP	(R10, R11), ExecCtx_R10(R0)
	STP	(R12, R13), ExecCtx_R12(R0)
	STP	(R14, R15), ExecCtx_R14(R0)
	STP	(R16, R17), ExecCtx_R16(R0)
	MOVD	R19, ExecCtx_R19(R0)
	STP	(R20, R21), ExecCtx_R20(R0)
	STP	(R22, R23), ExecCtx_R22(R0)
	STP	(R24, R25), ExecCtx_R24(R0)
	STP	(R26, R27), ExecCtx_R26(R0)
	STP	(g, R29), ExecCtx_R28(R0)
	MOVD	R30, ExecCtx_R30(R0)

	// save stack pointer
	MRS	SP_EL0, R1
	MOVD	R1, ExecCtx_R31(R0)

	// save PC
	MRS	ELR_EL1, R1
	MOVD	R1, ExecCtx_PC(R0)

	// save mode
	MRS	SPSR_EL1, R1
	MOVD	R1, ExecCtx_SPSR(R0)

	// save Fault Address Register
	MRS	FAR_EL1, R1
	MOVD	R1, ExecCtx_FAR(R0)

	// save Exception Syndrome Register
	MRS	ESR_EL1, R1
	MOVD	R1, ExecCtx_ESR(R0)

	// save floating-point registers
	ADD	$ExecCtx_VFP, R0, R1
	VST1.P	[V0.B16, V1.B16, V2.B16, V3.B16], 64(R1)
	VST1.P	[V4.B16, V5.B16, V6.B16, V7.B16], 64(R1)
	VST1.P	[V8.B16, V9.B16, V10.B16, V11.B16], 64(R1)
	VST1.P	[V12.B16, V13.B16, V14.B16, V15.B16], 64(R1)
	VST1.P	[V16.B16, V17.B16, V18.B16, V19.B16], 64(R1)
	VST1.P	[V20.B16, V21.B16, V22.B16, V23.B16], 64(R1)
	VST1.P	[V24.B16, V25.B16, V26.B16, V27.B16], 64(R1)
	VST1.P	[V28.B16, V29.B16, V30.B16, V31.B16], 64(R1)
	MOVD	FPSR, R1
	MOVD	R1, ExecCtx_FPSR(R0)
	MOVD	FPCR, R1
	MOVD	R1, ExecCtx_FPCR(R0)

	// restore caller registers
	LDP	(0*16)+ExecCtx_host(R0), (g, R1)
	MOVD	R1, FPCR
	LDP	(1*16)+ExecCtx_host(R0), (R29, R30)
	MOVD	(2*16)+ExecCtx_host(R0), R1
	MSR	R1, DAIF

	// restore g stack pointer
	MOVD	ExecCtx_g_sp(R0), R1
	MOVD	R1, RSP

	RET
