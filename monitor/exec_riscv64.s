// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

#include "go_asm_riscv64.h"

// GoTEE exception handling relies on one exit point (Exec) and one return
// point (monitor), both are used for execution of a Supervisor mode execution
// context.
//
// An execution context (ExecCtx) structure is used to hold initial register
// state at execution as well as store the updated state on re-entry.
//
// The exception handling uses the RISC-V Machine Scratch Register (mscratch)
// in a similar manner to Go own use of TLS (https://golang.org/src/runtime/tls_arm.s).
//
// The handler must save and restore the following registers between Machine <>
// Supervisor mode switches:
//
//  • x1-x31 general purpose registers:
//
//    TamaGo (and therefore GoTEE) runs in Machine mode and does not use
//    Supervisor or User mode. The pre-exception x1-x31 registers are
//    saved/restored.
//
//  • f0-f31 floating point registers:
//
//    TamaGo does not use the fcsr register, therefore only pre-exception
//    f0-f31 registers are saved/restored.

// func Exec(ctx *ExecCtx)
TEXT ·Exec(SB),$0-8
	// save general purpose registers
	MOV	X1, -2*8(SP)
	MOV	X3, -3*8(SP)
	MOV	TP, -4*8(SP)
	MOV	X5, -5*8(SP)
	MOV	X6, -6*8(SP)
	MOV	X7, -7*8(SP)
	MOV	X8, -8*8(SP)
	MOV	X9, -9*8(SP)
	MOV	X10, -10*8(SP)
	MOV	X11, -11*8(SP)
	MOV	X12, -12*8(SP)
	MOV	X13, -13*8(SP)
	MOV	X14, -14*8(SP)
	MOV	X15, -15*8(SP)
	MOV	X16, -16*8(SP)
	MOV	X17, -17*8(SP)
	MOV	X18, -18*8(SP)
	MOV	X19, -19*8(SP)
	MOV	X20, -20*8(SP)
	MOV	X21, -21*8(SP)
	MOV	X22, -22*8(SP)
	MOV	X23, -23*8(SP)
	MOV	X24, -24*8(SP)
	MOV	X25, -25*8(SP)
	MOV	X26, -26*8(SP)
	MOV	g,   -27*8(SP)
	MOV	X28, -28*8(SP)
	MOV	X29, -29*8(SP)
	MOV	X30, -30*8(SP)
	MOV	X31, -31*8(SP)

	// get argument pointer
	MOV	ctx+0(FP), T0

	// save g stack pointer
	MOV	SP, ExecCtx_g_sp(T0)

	// save context pointer
	CSRW(t0, mscratch)

	// switch to supervisor mode
	MOV	$const_Supervisor << 11, T1
	CSRS(t1, mstatus)

	// restore floating-point registers
	MOVD	(1*8)+ExecCtx_F(T0), F0
	MOVD	(2*8)+ExecCtx_F(T0), F1
	MOVD	(3*8)+ExecCtx_F(T0), F2
	MOVD	(4*8)+ExecCtx_F(T0), F3
	MOVD	(5*8)+ExecCtx_F(T0), F4
	MOVD	(6*8)+ExecCtx_F(T0), F5
	MOVD	(7*8)+ExecCtx_F(T0), F6
	MOVD	(8*8)+ExecCtx_F(T0), F7
	MOVD	(9*8)+ExecCtx_F(T0), F8
	MOVD	(10*8)+ExecCtx_F(T0), F9
	MOVD	(11*8)+ExecCtx_F(T0), F10
	MOVD	(12*8)+ExecCtx_F(T0), F11
	MOVD	(13*8)+ExecCtx_F(T0), F12
	MOVD	(14*8)+ExecCtx_F(T0), F13
	MOVD	(15*8)+ExecCtx_F(T0), F14
	MOVD	(16*8)+ExecCtx_F(T0), F15
	MOVD	(17*8)+ExecCtx_F(T0), F16
	MOVD	(18*8)+ExecCtx_F(T0), F17
	MOVD	(19*8)+ExecCtx_F(T0), F18
	MOVD	(20*8)+ExecCtx_F(T0), F19
	MOVD	(21*8)+ExecCtx_F(T0), F20
	MOVD	(22*8)+ExecCtx_F(T0), F21
	MOVD	(23*8)+ExecCtx_F(T0), F22
	MOVD	(24*8)+ExecCtx_F(T0), F23
	MOVD	(25*8)+ExecCtx_F(T0), F24
	MOVD	(26*8)+ExecCtx_F(T0), F25
	MOVD	(27*8)+ExecCtx_F(T0), F26
	MOVD	(28*8)+ExecCtx_F(T0), F27
	MOVD	(29*8)+ExecCtx_F(T0), F28
	MOVD	(30*8)+ExecCtx_F(T0), F29
	MOVD	(31*8)+ExecCtx_F(T0), F30
	MOVD	(32*8)+ExecCtx_F(T0), F31

	// restore general purpose registers
	MOV	ExecCtx_X1(T0), X1
	MOV	ExecCtx_X2(T0), X2
	MOV	ExecCtx_X4(T0), TP
	MOV	ExecCtx_X3(T0), X3
	MOV	ExecCtx_X7(T0), X7
	MOV	ExecCtx_X8(T0), X8
	MOV	ExecCtx_X9(T0), X9
	MOV	ExecCtx_X10(T0), X10
	MOV	ExecCtx_X11(T0), X11
	MOV	ExecCtx_X12(T0), X12
	MOV	ExecCtx_X13(T0), X13
	MOV	ExecCtx_X14(T0), X14
	MOV	ExecCtx_X15(T0), X15
	MOV	ExecCtx_X16(T0), X16
	MOV	ExecCtx_X17(T0), X17
	MOV	ExecCtx_X18(T0), X18
	MOV	ExecCtx_X19(T0), X19
	MOV	ExecCtx_X20(T0), X20
	MOV	ExecCtx_X21(T0), X21
	MOV	ExecCtx_X22(T0), X22
	MOV	ExecCtx_X23(T0), X23
	MOV	ExecCtx_X24(T0), X24
	MOV	ExecCtx_X25(T0), X25
	MOV	ExecCtx_X26(T0), X26
	MOV	ExecCtx_X27(T0), g
	MOV	ExecCtx_X28(T0), X28
	MOV	ExecCtx_X29(T0), X29
	MOV	ExecCtx_X30(T0), X30
	MOV	ExecCtx_X31(T0), X31

	MOV	ExecCtx_PC(T0), T1
	CSRW(t1, mepc)

	// restore T1, T0
	MOV	ExecCtx_X6(T0), T1
	MOV	ExecCtx_X5(T0), T0

	MRET

TEXT ·monitor(SB),NOSPLIT|NOFRAME,$0
	// restore context pointer
	CSRRW(t0, mscratch, t0)

	// save general purpose registers
	MOV	X1, ExecCtx_X1(T0)
	MOV	X2, ExecCtx_X2(T0)
	MOV	X3, ExecCtx_X3(T0)
	MOV	TP, ExecCtx_X4(T0)
	MOV	X6, ExecCtx_X6(T0)
	MOV	X7, ExecCtx_X7(T0)
	MOV	X8, ExecCtx_X8(T0)
	MOV	X9, ExecCtx_X9(T0)
	MOV	X10, ExecCtx_X10(T0)
	MOV	X11, ExecCtx_X11(T0)
	MOV	X12, ExecCtx_X12(T0)
	MOV	X13, ExecCtx_X13(T0)
	MOV	X14, ExecCtx_X14(T0)
	MOV	X15, ExecCtx_X15(T0)
	MOV	X16, ExecCtx_X16(T0)
	MOV	X17, ExecCtx_X17(T0)
	MOV	X18, ExecCtx_X18(T0)
	MOV	X19, ExecCtx_X19(T0)
	MOV	X20, ExecCtx_X20(T0)
	MOV	X21, ExecCtx_X21(T0)
	MOV	X22, ExecCtx_X22(T0)
	MOV	X23, ExecCtx_X23(T0)
	MOV	X24, ExecCtx_X24(T0)
	MOV	X25, ExecCtx_X25(T0)
	MOV	X26, ExecCtx_X26(T0)
	MOV	g,   ExecCtx_X27(T0)
	MOV	X28, ExecCtx_X28(T0)
	MOV	X29, ExecCtx_X29(T0)
	MOV	X30, ExecCtx_X30(T0)
	MOV	X31, ExecCtx_X31(T0)

	// save floating-point registers
	MOVD	F0, (1*8)+ExecCtx_F(T0)
	MOVD	F1, (2*8)+ExecCtx_F(T0)
	MOVD	F2, (3*8)+ExecCtx_F(T0)
	MOVD	F3, (4*8)+ExecCtx_F(T0)
	MOVD	F4, (5*8)+ExecCtx_F(T0)
	MOVD	F5, (6*8)+ExecCtx_F(T0)
	MOVD	F6, (7*8)+ExecCtx_F(T0)
	MOVD	F7, (8*8)+ExecCtx_F(T0)
	MOVD	F8, (9*8)+ExecCtx_F(T0)
	MOVD	F9, (10*8)+ExecCtx_F(T0)
	MOVD	F10, (11*8)+ExecCtx_F(T0)
	MOVD	F11, (12*8)+ExecCtx_F(T0)
	MOVD	F12, (13*8)+ExecCtx_F(T0)
	MOVD	F13, (14*8)+ExecCtx_F(T0)
	MOVD	F14, (15*8)+ExecCtx_F(T0)
	MOVD	F15, (16*8)+ExecCtx_F(T0)
	MOVD	F16, (17*8)+ExecCtx_F(T0)
	MOVD	F17, (18*8)+ExecCtx_F(T0)
	MOVD	F18, (19*8)+ExecCtx_F(T0)
	MOVD	F19, (20*8)+ExecCtx_F(T0)
	MOVD	F20, (21*8)+ExecCtx_F(T0)
	MOVD	F21, (22*8)+ExecCtx_F(T0)
	MOVD	F22, (23*8)+ExecCtx_F(T0)
	MOVD	F23, (24*8)+ExecCtx_F(T0)
	MOVD	F24, (25*8)+ExecCtx_F(T0)
	MOVD	F25, (26*8)+ExecCtx_F(T0)
	MOVD	F26, (27*8)+ExecCtx_F(T0)
	MOVD	F27, (28*8)+ExecCtx_F(T0)
	MOVD	F28, (29*8)+ExecCtx_F(T0)
	MOVD	F29, (30*8)+ExecCtx_F(T0)
	MOVD	F30, (31*8)+ExecCtx_F(T0)
	MOVD	F31, (32*8)+ExecCtx_F(T0)

	// save PC
	CSRR(mepc, t1)
	ADD	$(4), T1, T1
	MOV	T1, ExecCtx_PC(T0)

	// save PC from RA
	//MOV	RA, ExecCtx_PC(T0)

	// save T0
	CSRR(mscratch, t1)
	MOV	T1, ExecCtx_X5(T0)

	// save MCAUSE
	CSRR(mcause, t1)
	MOV	T1, ExecCtx_MCAUSE(T0)

	// restore g registers
	MOV	ExecCtx_g_sp(T0), SP
	MOV	-2*8(SP), X1
	MOV	-3*8(SP), X3
	MOV	-4*8(SP), TP
	MOV	-5*8(SP), X5
	MOV	-6*8(SP), X6
	MOV	-7*8(SP), X7
	MOV	-8*8(SP), X8
	MOV	-9*8(SP), X9
	MOV	-10*8(SP), X10
	MOV	-11*8(SP), X11
	MOV	-12*8(SP), X12
	MOV	-13*8(SP), X13
	MOV	-14*8(SP), X14
	MOV	-15*8(SP), X15
	MOV	-16*8(SP), X16
	MOV	-17*8(SP), X17
	MOV	-18*8(SP), X18
	MOV	-19*8(SP), X19
	MOV	-20*8(SP), X20
	MOV	-21*8(SP), X21
	MOV	-22*8(SP), X22
	MOV	-23*8(SP), X23
	MOV	-24*8(SP), X24
	MOV	-25*8(SP), X25
	MOV	-26*8(SP), X26
	MOV	-27*8(SP), g
	MOV	-28*8(SP), X28
	MOV	-29*8(SP), X29
	MOV	-30*8(SP), X30
	MOV	-31*8(SP), X31

	JMP	(RA)
