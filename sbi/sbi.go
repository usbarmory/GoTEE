// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package sbi provides basic RISC-V Supervisor Binary Interface Specification
// support for TamaGo unikernels launched in supervised mode through
// monitor.Exec (see monitor package).
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package sbi

import (
	"github.com/usbarmory/GoTEE/monitor"
)

const (
	SBI_MAJOR = 1
	SBI_MINOR = 0
)

// Supported SBI Extension IDs (EID)
const (
	EXT_BASE = 0x10
)

// Base Extension Function IDs (FIDs)
const (
	EXT_BASE_GET_SPEC_VERSION = iota
	EXT_BASE_GET_IMP_ID
	EXT_BASE_GET_IMP_VERSION
	EXT_BASE_PROBE_EXT
	EXT_BASE_GET_MVENDORID
	EXT_BASE_GET_MARCHID
	EXT_BASE_GET_MIMPID
)

// Standard SBI Errors
const (
	SBI_SUCCESS               = 0
	SBI_ERR_FAILED            = -1
	SBI_ERR_NOT_SUPPORTED     = -2
	SBI_ERR_INVALID_PARAM     = -3
	SBI_ERR_DENIED            = -4
	SBI_ERR_INVALID_ADDRESS   = -5
	SBI_ERR_ALREADY_AVAILABLE = -6
	SBI_ERR_ALREADY_STARTED   = -7
	SBI_ERR_ALREADY_STOPPED   = -8
)

type sbiret struct {
	Error int64
	Value int64
}

func baseHandler(ctx *monitor.ExecCtx) (ret sbiret) {
	switch ctx.X16 {
	case EXT_BASE_GET_SPEC_VERSION:
		ret.Value = (SBI_MAJOR << 24) | SBI_MINOR
	case EXT_BASE_GET_IMP_ID, EXT_BASE_GET_IMP_VERSION:
		// report no supported EIDs or implementation details
	case EXT_BASE_PROBE_EXT:
		// report no support for other extensions
	case EXT_BASE_GET_MVENDORID, EXT_BASE_GET_MARCHID, EXT_BASE_GET_MIMPID:
		// zero is always a legal value for these CSRs
	default:
		ret.Error = SBI_ERR_NOT_SUPPORTED
	}

	return
}

// Handler implements basic support for RISC-V SBI calls raised by an execution
// context, it provides minimal support for SBI probing by S-mode kernels. Only
// the Base extension is implemented to report that no other SBI extensions are
// available.
func Handler(ctx *monitor.ExecCtx) (err error) {
	var ret sbiret

	switch ctx.X17 {
	case EXT_BASE:
		ret = baseHandler(ctx)
	default:
		ret.Error = SBI_ERR_NOT_SUPPORTED
	}

	ctx.X10 = uint64(ret.Error)
	ctx.X11 = uint64(ret.Value)

	return
}
