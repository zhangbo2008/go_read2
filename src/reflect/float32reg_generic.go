// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !ppc64 && !ppc64le && !riscv64

package reflect

import "unsafe"

// This file implements a straightforward conversion of a float32
// value into its representation in a register. This conversion
// applies for amd64 and arm64. It is also chosen for the case of
// zero argument registers, but is not used.

func archFloat32FromReg(reg uint64) float32 { //reg表示寄存器中保存的一个整数.我们把他转成u32,然后返回他的指针.
	i := uint32(reg)
	return *(*float32)(unsafe.Pointer(&i))
}

func archFloat32ToReg(val float32) uint64 { //寄存器里面都是用u64的,所以我们把这个值转64即可.
	return uint64(*(*uint32)(unsafe.Pointer(&val)))
}
