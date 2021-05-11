// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package tee provides common constants/functions for GoTEE supervisor/applets
// written using the TamaGo framework.
package tee

// TODO: for now we take a lazy approach of using 32MB at the beginning of
// available RAM (USB armory Mk II layout) for the supervisor and the following
// 32MB for the trusted applet.

const (
	// KernelStart defines the privileged RAM start address
	KernelStart = 0x80000000
	// KernelSize defines the privileged RAM size
	KernelSize = 0x2000000

	// AppletStart defines the unprivileged RAM start address
	AppletStart = 0x82000000
	// AppletSize defines the unprivileged RAM size
	AppletSize = 0x2000000
	// AppletStackOffset defines the unprivileged stack offset
	AppletStackOffset = 0x100
)
