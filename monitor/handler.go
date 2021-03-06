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

	"github.com/usbarmory/GoTEE/syscall"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/imx6"
	"github.com/usbarmory/tamago/soc/imx6/imx6ul"
)

// SecureHandler is the default handler for supervisor (SVC) exceptions raised
// by a secure execution context to handle supported GoTEE system calls.
func SecureHandler(ctx *ExecCtx) (err error) {
	switch num := ctx.R0; num {
	case syscall.SYS_EXIT:
		return errors.New("exit")
	case syscall.SYS_WRITE:
		imx6ul.UART2.Tx(byte(ctx.R1))
	case syscall.SYS_NANOTIME:
		t := int64(imx6.ARM.TimerFn() * imx6.ARM.TimerMultiplier)
		ctx.R0 = uint32(t & 0xffffffff)
		ctx.R1 = uint32(t >> 32)
	case syscall.SYS_GETRANDOM:
		off := int(ctx.R1 - ctx.Memory.Start)
		buf := make([]byte, ctx.R2)

		if !(off >= 0 && off < (ctx.Memory.Size-len(buf))) {
			return errors.New("invalid read offset")
		}

		if n, err := rand.Read(buf); err != nil || n != int(ctx.R2) {
			return errors.New("internal error")
		}

		ctx.Memory.Write(ctx.Memory.Start, off, buf)
	case syscall.SYS_RPC_REQ, syscall.SYS_RPC_RES:
		if ctx.Server != nil {
			err = ctx.rpc()
		}
	default:
		err = fmt.Errorf("invalid syscall %d", num)
	}

	return
}

// NonSecureHandler is the default handler for supervisor (SVC) exceptions
// raised by a non-secure execution context to handle supported GoTEE secure
// monitor calls.
func NonSecureHandler(ctx *ExecCtx) (err error) {
	vector := ctx.ExceptionVector

	if vector != arm.SUPERVISOR {
		if ctx.Debug {
			ctx.Print()
		}

		return errors.New(arm.VectorName(vector))
	}

	log.Printf("TODO: Secure <> NonSecure API")

	return
}
