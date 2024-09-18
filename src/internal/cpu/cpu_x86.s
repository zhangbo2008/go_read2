// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || amd64

#include "textflag.h"

// func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)  //输入加出参一共6个4字节的.所以占用24. 不需要额外的栈空间,  所以是$0-24
TEXT ·cpuid(SB), NOSPLIT, $0-24
	MOVL eaxArg+0(FP), AX
	MOVL ecxArg+4(FP), CX
	CPUID            // 这个数cpuid函数调用. 这个函数说明可以参考https://www.felixcloutier.com/x86/cpuid// Returns processor identification and feature information to the EAX, EBX, ECX, and EDX registers, as determined by input entered in EAX (in some cases, ECX as well).

	MOVL AX, eax+8(FP)
	MOVL BX, ebx+12(FP)
	MOVL CX, ecx+16(FP)
	MOVL DX, edx+20(FP)   //这里面eax,ebx,ecx,edx都是命名.
	RET

// func xgetbv() (eax, edx uint32)
TEXT ·xgetbv(SB),NOSPLIT,$0-8
	MOVL $0, CX
	XGETBV   // https://www.felixcloutier.com/x86/xgetbv  Reads an XCR specified by ECX into EDX:EAX.
	MOVL AX, eax+0(FP)
	MOVL DX, edx+4(FP)
	RET

// func getGOAMD64level() int32
TEXT ·getGOAMD64level(SB),NOSPLIT,$0-4 //就是简单的if else函数.
#ifdef GOAMD64_v4
	MOVL $4, ret+0(FP)
#else
#ifdef GOAMD64_v3
	MOVL $3, ret+0(FP)
#else
#ifdef GOAMD64_v2
	MOVL $2, ret+0(FP)
#else
	MOVL $1, ret+0(FP)
#endif
#endif
#endif
	RET
