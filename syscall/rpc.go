// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package syscall

import (
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// Stream implements a data stream interface to exchange data buffers between
// user and supervisor mode over syscalls.
//
// It is used by NewClient() to stream JSON-RPC calls from user mode and
// receive responses from the supervisor, over syscalls.
//
// The implementation is not safe against concurrent reads and writes, which
// should be avoided.
type Stream struct {
	// ReadSyscall is the syscall number associated to Read()
	ReadSyscall uint
	// ReadSyscall is the syscall number associated to Write()
	WriteSyscall uint
}

// Read reads up to len(p) bytes into p, it never returns an error. The read is
// requested, through the Stream ReadSyscall, to the supervisor.
func (s *Stream) Read(p []byte) (int, error) {
	// Go net/rpc client constantly polls for responses, even when no
	// outstanding requests are present (see *Client.input() in
	// net/rpc/client.go)
	//
	// To avoid a background loop of SYS_READ system calls we return io.EOF
	// to shut down the rpc.Client after each request.

	n := Read(s.ReadSyscall, p, uint(len(p)))
	return int(n), io.EOF
}

// Write writes len(p) bytes from p to the underlying data stream, it never
// returns an error. The write is issued, through the Stream WriteSyscall, to
// the supervisor.
func (s *Stream) Write(p []byte) (n int, err error) {
	n = len(p)
	Write(s.WriteSyscall, p, uint(n))
	return
}

// Close has no effect.
func (s *Stream) Close() error {
	return nil
}

// NewClient returns a new client suitable for RPC calls to the supervisor. The
// client automatically closes after Call() is invoked on it the first time,
// therefore a new instance is needed for each call (also see Call()).
func NewClient() *rpc.Client {
	return jsonrpc.NewClient(&Stream{
		ReadSyscall:  SYS_RPC_RES,
		WriteSyscall: SYS_RPC_REQ,
	})
}

// Call is a convenience method that issues an RPC call on a disposable client
// created with NewClient().
func Call(serviceMethod string, args interface{}, reply interface{}) error {
	return NewClient().Call(serviceMethod, args, reply)
}
