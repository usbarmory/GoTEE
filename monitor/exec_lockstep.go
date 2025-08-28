// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"errors"
)

// Clone returns a duplicate execution context suitable for lockstep operation
// (see Shadow field), the original Handler field is not carried over in the
// shadow copy.
func (ctx *ExecCtx) Clone() (shadow *ExecCtx) {
	s := *ctx
	s.Handler = nil

	return &s
}

// lockstep runs a shadow execution context for a single scheduling cycle, an
// error is raised when the resulting state differs from the primary execution
// context.
func (shadow *ExecCtx) lockstep(primary *ExecCtx) (err error) {
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
