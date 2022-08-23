// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func Supervisor()
TEXT ·Supervisor(SB),$0
	SWI	$0
	RET

// func Exit()
TEXT ·Exit(SB),$0
	MOVW	$const_SYS_EXIT, R0

	SWI	$0

	RET

// func Print(c byte)
TEXT ·Print(SB),$0-1
	MOVW	$const_SYS_WRITE, R0
	MOVB	c+0(FP), R1

	SWI	$0

	RET

// func Nanotime() int64
TEXT ·Nanotime(SB),$0-8
	MOVW	$const_SYS_NANOTIME, R0

	SWI	$0

	MOVW	R0, ret_lo+0(FP)
	MOVW	R1, ret_hi+4(FP)

	RET

// func Write(trap uint, b []byte, n uint)
TEXT ·Write(SB),$0-20
	MOVW	trap+0(FP), R0
	MOVW	b+4(FP), R1
	MOVW	n+16(FP), R2

	SWI	$0

	RET

// func Read(trap uint, b []byte, n uint) uint
TEXT ·Read(SB),$0-24
	MOVW	trap+0(FP), R0
	MOVW	b+4(FP), R1
	MOVW	n+16(FP), R2

	SWI	$0

	MOVW	R2, ret+20(FP)

	RET
