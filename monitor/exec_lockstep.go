// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"errors"
)

// castShadow builds a shadow execution context suitable for lockstep()
func (ctx *ExecCtx) castShadow() *ExecCtx {
	shadow := *ctx
	shadow.Handler = nil

	return &shadow
}

// lockstep runs a shadow execution context for a single scheduling cycle, an
// error is raised when the resulting state differs from the primary execution
// context.
func (shadow *ExecCtx) lockstep(primary *ExecCtx) (err error) {
	shadow.Lockstep(true)
	defer shadow.Lockstep(false)

	if err = shadow.Schedule(); err != nil {
		return
	}

	if shadow.Handler != nil {
		if err = shadow.Handler(primary); err != nil {
			return
		}
	}

	if !Equal(primary, shadow) {
		return errors.New("lockstep failure")
	}

	return
}
