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
func (ctx *ExecCtx) Read(p []byte) (int, error) {
	off := ctx.A1() - ctx.Memory.Start() + ctx.off
	n := ctx.A2() - ctx.off
	s := uint(len(p))

	switch {
	case n <= 0:
		ctx.off = 0
		return -1, nil
	case n <= s:
		ctx.off = 0
		s = n
	case n > s:
		ctx.off += s
	}

	if !(off >= 0 && off < (ctx.Memory.Size()-s)) {
		return 0, errors.New("invalid offset")
	}

	ctx.Memory.Read(ctx.Memory.Start(), int(off), p[0:s])

	return int(s), nil
}

// Write writes len(p) bytes from p to the underlying data stream, it never
// returns an error. The written data is buffered within the execution context,
// waiting for its read through syscall.Read().
func (ctx *ExecCtx) Write(p []byte) (int, error) {
	ctx.buf = p
	return len(p), nil
}

// Flush writes data previously buffered by Write to the execution context
// memory.
func (ctx *ExecCtx) Flush() error {
	var last bool

	off := ctx.A1() - ctx.Memory.Start()
	n := ctx.A2()
	s := uint(len(ctx.buf))

	if s > n {
		s = n
	} else {
		last = true
	}

	if !(off >= 0 && off < (ctx.Memory.Size()-s)) {
		return errors.New("invalid offset")
	}

	ctx.Memory.Write(ctx.Memory.Start(), int(off), ctx.buf[0:s])
	ctx.Ret(s)

	if last {
		ctx.buf = nil
	} else {
		ctx.buf = ctx.buf[s:]
	}

	return nil
}

// Close has no effect.
func (ctx *ExecCtx) Close() error {
	return nil
}

func (ctx *ExecCtx) rpc() (err error) {
	switch num := ctx.A0(); num {
	case syscall.SYS_RPC_REQ:
		err = ctx.Server.ServeRequest(jsonrpc.NewServerCodec(ctx))
	case syscall.SYS_RPC_RES:
		err = ctx.Flush()
	default:
		return fmt.Errorf("invalid syscall %d", num)
	}

	return
}
