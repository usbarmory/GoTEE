// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"errors"
)

func (ctx *ExecCtx) castShadow() *ExecCtx {
	shadow := *ctx
	shadow.Handler = nil
	shadow.Shadow = ctx

	return &shadow
}

// lockstep runs a shadow execution context for a single scheduling cycle, an
// error is raised when the resulting state differs from the primary execution
// context.
func (ctx *ExecCtx) lockstep() (err error) {
	ctx.Lockstep(true)
	defer ctx.Lockstep(false)

	if err = ctx.schedule(); err != nil {
		return
	}

	if ctx.Handler != nil {
		if err = ctx.Handler(ctx); err != nil {
			return
		}
	}

	if !Equal(ctx, ctx.Shadow) {
		return errors.New("lockstep failure")
	}

	return
}
