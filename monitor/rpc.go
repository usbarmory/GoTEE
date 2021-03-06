// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"errors"
	"fmt"
	"net/rpc/jsonrpc"

	"github.com/usbarmory/GoTEE/syscall"
)

// Read reads up to len(p) bytes into p. The read data is received from the
// execution context memory, after it is being written with syscall.Write().
func (ctx *ExecCtx) Read(p []byte) (n int, err error) {
	off := int(ctx.R1 - ctx.Memory.Start)
	n = int(ctx.R2)

	if len(p) < n {
		return 0, errors.New("invalid read length")
	}

	if !(off >= 0 && off < (ctx.Memory.Size-len(p))) {
		return 0, errors.New("invalid read offset")
	}

	ctx.Memory.Read(ctx.Memory.Start, off, p)

	return
}

// Write writes len(p) bytes from p to the underlying data stream, it never
// returns an error. The written data is buffered within the execution context,
// waiting for its read through syscall.Read().
func (ctx *ExecCtx) Write(p []byte) (int, error) {
	ctx.buf = p
	return len(p), nil
}

// Close has no effect.
func (ctx *ExecCtx) Close() error {
	return nil
}

func (ctx *ExecCtx) rpc() (err error) {
	switch num := ctx.R0; num {
	case syscall.SYS_RPC_REQ:
		err = ctx.Server.ServeRequest(jsonrpc.NewServerCodec(ctx))
	case syscall.SYS_RPC_RES:
		off := int(ctx.R1 - ctx.Memory.Start)
		n := int(ctx.R2)

		if len(ctx.buf) > n {
			return errors.New("invalid buffer size")
		}

		if !(off >= 0 && off < (ctx.Memory.Size-len(ctx.buf))) {
			return errors.New("invalid read offset")
		}

		ctx.Memory.Write(ctx.Memory.Start, off, ctx.buf)
		ctx.R2 = uint32(len(ctx.buf))
		ctx.buf = nil

		return
	default:
		err = fmt.Errorf("invalid syscall %d", num)
	}

	return
}
