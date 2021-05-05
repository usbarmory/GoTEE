// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"github.com/f-secure-foundry/tamago/arm"
)

// defined in monitor.s
func monitor()

var MonitorVectorTable = arm.VectorTable{
	Reset:         monitor,
	Undefined:     monitor,
	Supervisor:    monitor,
	PrefetchAbort: monitor,
	DataAbort:     monitor,
	IRQ:           monitor,
	FIQ:           monitor,
}
