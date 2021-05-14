// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

// example receiver for user mode <--> system RPC over system calls

type Receiver struct{}

func (rcv *Receiver) Echo(in string, out *string) error {
	*out = in
	return nil
}
