// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func write_nsacr(scr uint32)
TEXT ·write_nsacr(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B4.1.111 NSACR, Non-Secure Access Control Register, Security Extensions
	MOVW	scr+0(FP), R0
	MCR	15, 0, R0, C1, C1, 2

	RET
