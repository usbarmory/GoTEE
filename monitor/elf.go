// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package monitor

import (
	"bytes"
	"debug/elf"
	_ "embed"
	"fmt"

	"github.com/f-secure-foundry/tamago/dma"
)

func parseELF(mem *dma.Region, buf []byte) (entry uint32) {
	f, err := elf.NewFile(bytes.NewReader(buf))

	if err != nil {
		panic(err)
	}

	for idx, prg := range f.Progs {
		if prg.Type != elf.PT_LOAD {
			continue
		}

		b := make([]byte, prg.Memsz)

		_, err := prg.ReadAt(b[0:prg.Filesz], 0)

		if err != nil {
			panic(fmt.Sprintf("failed to read LOAD section at idx %d, %q", idx, err))
		}

		offset := uint32(prg.Paddr) - mem.Start
		mem.Write(mem.Start, b, int(offset))
	}

	return uint32(f.Entry)
}
