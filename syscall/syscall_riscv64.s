// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// A7 must be set to 0 to avoid interference with SBI

// func Supervisor()
TEXT ·Supervisor(SB),$0
	MOV	$0, A7
	ECALL

	RET

// func Exit()
TEXT ·Exit(SB),$0
	MOV	$const_SYS_EXIT, A0

	MOV	$0, A7
	ECALL

	RET

// func Print(c byte)
TEXT ·Print(SB),$0-1
	MOV	$const_SYS_WRITE, A0
	MOV	c+0(FP), A1

	MOV	$0, A7
	ECALL

	RET

// func Nanotime() int64
TEXT ·Nanotime(SB),$0-8
	MOV	$const_SYS_NANOTIME, A0

	MOV	$0, A7
	ECALL

	MOV	A0, ret+0(FP)

	RET

// func Write(trap uint, b []byte, n uint)
TEXT ·Write(SB),$0-40
	MOV	trap+0(FP), A0
	MOV	b+8(FP), A1
	MOV	n+32(FP), A2

	MOV	$0, A7
	ECALL

	RET

// func Read(trap uint, b []byte, n uint) int
TEXT ·Read(SB),$0-48
	MOV	trap+0(FP), A0
	MOV	b+8(FP), A1
	MOV	n+32(FP), A2

	MOV	$0, A7
	ECALL

	MOV	A0, ret+40(FP)

	RET
