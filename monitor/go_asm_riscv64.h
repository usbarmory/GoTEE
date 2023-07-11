// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in Go LICENSE file.

#define t0 5
#define t1 6

#define satp     0x180
#define mstatus  0x300
#define mscratch 0x340
#define mepc     0x341
#define mcause   0x342

#define CSRW(RS,CSR) WORD $(0x1073 + RS<<15 + CSR<<20)
#define CSRR(CSR,RD) WORD $(0x2073 + RD<<7 + CSR<<20)
#define CSRS(RS,CSR) WORD $(0x2073 + RS<<15 + CSR<<20)
#define CSRC(RS,CSR) WORD $(0x3073 + RS<<15 + CSR<<20)
#define CSRRW(RS,CSR,RD) WORD $(0x1073 + RD<<7 + RS<<15 + CSR<<20)

#define MRET  WORD $0x30200073
