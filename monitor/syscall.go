// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import "errors"

// TransferRegion validates the registers used in memory transfer request for
// GoTEE secure monitor calls (syscall.Read(), syscall.Write()) and returns the
// computed memory offset and transfer size.
func (ctx *ExecCtx) TransferRegion() (off int, n int, err error) {
	off = int(ctx.A1()) - int(ctx.Memory.Start())
	n = int(ctx.A2())
	s := int(ctx.Memory.Size())

	if valid := (off >= 0) && (n <= s) && (off < s-n); !valid {
		err = errors.New("invalid offset")
	}

	return
}
