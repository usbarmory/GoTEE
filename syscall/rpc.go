// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
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
	ReadSyscall  uint
	// ReadSyscall is the syscall number associated to Write()
	WriteSyscall uint
}

// Read reads up to len(p) bytes into p, it never returns an error. The read is
// requested, through the Stream ReadSyscall, to the supervisor.
func (s *Stream) Read(p []byte) (int, error) {
	n := Read(s.ReadSyscall, p, uint(len(p)))
	return int(n), io.EOF
	// FIXME: on second try this leads to RPC error: connection is shut down
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

// NewClient returns a new client suitable for RPC calls to the supervisor.
func NewClient() *rpc.Client {
	return jsonrpc.NewClient(&Stream{
		ReadSyscall:  SYS_RPC_RES,
		WriteSyscall: SYS_RPC_REQ,
	})
}
