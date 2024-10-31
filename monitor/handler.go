// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/usbarmory/GoTEE/syscall"
)

// SecureHandler is the default handler for exceptions raised by a secure
// execution context to handle supported GoTEE system calls.
func SecureHandler(ctx *ExecCtx) (err error) {
	switch num := ctx.A0(); num {
	case syscall.SYS_EXIT:
		ctx.Stop()
	case syscall.SYS_WRITE:
		print(string(ctx.A1()))
	case syscall.SYS_NANOTIME:
		ctx.Ret(time.Now().UnixNano())
	case syscall.SYS_GETRANDOM:
		off, n, err := ctx.TransferRegion()

		if err != nil {
			return err
		}

		buf := make([]byte, n)

		if _, err := rand.Read(buf); err != nil {
			return errors.New("internal error")
		}

		ctx.Poke(off, buf)
	case syscall.SYS_RPC_REQ, syscall.SYS_RPC_RES:
		if ctx.Server != nil {
			err = ctx.rpc()
		}
	default:
		err = fmt.Errorf("invalid syscall %d", num)
	}

	return
}

// NonSecureHandler is the default handler for exceptions raised by a
// non-secure execution context to handle supported GoTEE secure monitor calls.
func NonSecureHandler(ctx *ExecCtx) (err error) {
	// to be overridden by application
	log.Printf("NonSecureHandler: unimplemented")
	return
}
