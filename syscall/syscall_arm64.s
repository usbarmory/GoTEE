// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func Supervisor()
TEXT ·Supervisor(SB),$0
	SVC
	RET

// func Exit()
TEXT ·Exit(SB),$0
	MOVW	$const_SYS_EXIT, R0

	SVC

	RET

// func Print(c byte)
TEXT ·Print(SB),$0-1
	MOVD	$const_SYS_WRITE, R0
	MOVBU	c+0(FP), R1

	SVC

	RET

// func Nanotime() int64
TEXT ·Nanotime(SB),$0-8
	MOVD	$const_SYS_NANOTIME, R0

	SVC

	MOVD	R0, ret+0(FP)
	RET

// func Write(trap uint, p []byte, n uint)
TEXT ·Write(SB),$0-40
	MOVD	trap+0(FP), R0
	MOVD	p+8(FP), R1
	MOVD	n+32(FP), R2

	SVC

	RET

// func Read(trap uint, p []byte, n uint) int
TEXT ·Read(SB),$0-48
	MOVD	trap+0(FP), R0
	MOVD	p+8(FP), R1
	MOVD	n+32(FP), R2

	SVC

	MOVD	R0, ret+40(FP)
	RET
