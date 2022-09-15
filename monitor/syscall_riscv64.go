// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

// A0 returns the register treated as first argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A0() uint {
	return uint(ctx.X10)
}

// A1 returns the register treated as second argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A1() uint {
	return uint(ctx.X11)
}

// A2 returns the register treated as third argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A2() uint {
	return uint(ctx.X12)
}

// Ret sets the return value for GoTEE secure monitor calls.
func (ctx *ExecCtx) Ret(val interface{}) {
	switch v := val.(type) {
	case uint64:
		ctx.X10 = v
	case uint:
		ctx.X10 = uint64(v)
	case int64:
		ctx.X10 = uint64(v)
	case int:
		ctx.X10 = uint64(v)
	default:
		panic("invalid return type")
	}
}
