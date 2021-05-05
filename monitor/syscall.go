// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"crypto/rand"
	"fmt"

	"github.com/f-secure-foundry/GoTEE/syscall"

	"github.com/f-secure-foundry/tamago/soc/imx6"
)

func Handler(ctx *ExecCtx) (err error) {
	switch num := ctx.R0; num {
	case syscall.SYS_EXIT:
		return fmt.Errorf("exit")
	case syscall.SYS_WRITE:
		imx6.UART2.Tx(byte(ctx.R1))
	case syscall.SYS_UTIME:
		t := int64(imx6.ARM.TimerFn() * imx6.ARM.TimerMultiplier)
		ctx.R0 = uint32(t & 0xffffffff)
		ctx.R1 = uint32(t >> 32)
	case syscall.SYS_GETRANDOM:
		off := int(ctx.R1 - ctx.Memory.Start)
		buf := make([]byte, ctx.R2)

		if n, err := rand.Read(buf); err != nil || n != int(ctx.R2) {
			panic("fatal getrandom error")
		}

		ctx.Memory.Write(ctx.Memory.Start, buf, off)
	default:
		err = fmt.Errorf("invalid syscall %d", num)
	}

	return
}
