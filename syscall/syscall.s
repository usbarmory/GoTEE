// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func svc()
TEXT ·svc(SB),$0
	SWI	$0
	RET

// func Exit()
TEXT ·Exit(SB),$0
	MOVW	$const_SYS_EXIT, R0

	SWI	$0

	RET

// func Write(c byte)
TEXT ·Write(SB),$0-4
	MOVW	$const_SYS_WRITE, R0
	MOVW	c+0(FP), R1

	SWI	$0

	RET

// func Utime() int64
TEXT ·Utime(SB),$0-8
	MOVW	$const_SYS_UTIME, R0

	SWI	$0

	MOVW	R0, ret_lo+0(FP)
	MOVW	R1, ret_hi+4(FP)

	RET

// func GetRandom(b []byte, n uint)
TEXT ·GetRandom(SB),$0-8
	MOVW	$const_SYS_GETRANDOM, R0
	MOVW	b+0(FP), R1
	MOVW	n+4(FP), R2

	SWI	$0

	RET
