// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi

const (
	// See abi_generic.go.

	// RAX, RBX, RCX, RDI, RSI, R8, R9, R10, R11.  //这9个用来存整数.
	IntArgRegs = 9

	// X0 -> X14.
	FloatArgRegs = 15 //这15个用来存float

	// We use SSE2 registers which support 64-bit float operations.  The 8 registers are named xmm0 through xmm7.
	EffectiveFloatRegSize = 8
)
