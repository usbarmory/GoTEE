// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// stub for pkg.go.dev coverage
//go:build !tamago

package monitor

// A0 returns the register treated as first argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A0() uint

// A1 returns the register treated as second argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A1() uint

// A2 returns the register treated as third argument for GoTEE secure monitor
// calls.
func (ctx *ExecCtx) A2() uint

// Ret sets the return value for GoTEE secure monitor calls updating the
// relevant execution context registers, including its Shadow if present.
func (ctx *ExecCtx) Ret(val interface{})
