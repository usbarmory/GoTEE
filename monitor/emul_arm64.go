// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"encoding/binary"
	"unsafe"
)

// defined in emul_arm64.s
func getisar0() uint64
func getisar1() uint64
func getpfr0() uint64

var idRegs = map[uint32]func() uint64{
	0xd5380600: getisar0,
	0xd5380620: getisar1,
	0xd5380400: getpfr0,
}

// EmulateIDReg is an helper for handlers which need to emulate EL0 reads of ID
// registers, it returns whether a compatible read was found and handled.
func (ctx *ExecCtx) EmulateIDReg() bool {
	start := ctx.Memory.Start()
	off := uint(ctx.PC) - start

	// flat mapping is assumed
	buf := make([]byte, 4)
	ctx.Memory.Read(start, int(off), buf)
	insn := binary.LittleEndian.Uint32(buf)

	read, ok := idRegs[insn&^0x1f]

	if !ok {
		return false
	}

	rt := insn & 0x1f
	val := read()

	for _, c := range []*ExecCtx{ctx, ctx.Shadow} {
		if c == nil {
			continue
		}

		if rt != 31 {
			(*[31]uint64)(unsafe.Pointer(&c.R0))[rt] = val
		}

		c.PC += 4
	}

	return true
}
