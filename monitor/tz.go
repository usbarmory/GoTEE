// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

// defined in tz.s
func write_nsacr(uint32)

// NonSecureAccess sets the NSACR value, which defines the Non-Secure access
// permissions to coprocessors.
func NonSecureAccess(nsacr uint32) {
	write_nsacr(nsacr)
}
