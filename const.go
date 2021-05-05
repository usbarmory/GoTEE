// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package tee

// TODO: for now the 1st half of available RAM (USB armory Mk II layout) is
// dedicated to the trusted applet, while the 2nd half belongs to the
// supervisor.

const (
	// KernelStart defines the privileged RAM start address
	KernelStart = 0x90000000
	// KernelSize defines the privileged RAM size
	KernelSize = 0x10000000
	// AppletStart defines the unprivileged RAM start address
	AppletStart = 0x80000000
	// AppletSize defines the unprivileged RAM size
	AppletSize = 0x10000000
	// AppletStackOffset defines the unprivileged stack offset
	AppletStackOffset = 0x100
)
