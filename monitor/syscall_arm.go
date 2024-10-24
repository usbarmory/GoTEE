// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

// A0 returns the register treated as first argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A0() uint {
	return uint(ctx.R0)
}

// A1 returns the register treated as second argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A1() uint {
	return uint(ctx.R1)
}

// A2 returns the register treated as third argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A2() uint {
	return uint(ctx.R2)
}

// Ret sets the return value for GoTEE secure monitor calls.
func (ctx *ExecCtx) Ret(val interface{}) {
	var r0 uint32
	var r1 uint32

	switch v := val.(type) {
	case uint64:
		r0 = uint32(v & 0xffffffff)
		r1 = uint32(v >> 32)
	case uint:
		r0 = uint32(v)
	case int64:
		r0 = uint32(v & 0xffffffff)
		r1 = uint32(v >> 32)
	case int:
		r0 = uint32(v)
	default:
		panic("invalid return type")
	}

	ctx.R0 = r0
	ctx.R1 = r1

	if ctx.Shadow != nil {
		ctx.Shadow.R0 = r0
		ctx.Shadow.R1 = r1
	}
}
