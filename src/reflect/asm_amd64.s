// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

// The frames of each of the two functions below contain two locals, at offsets
// that are known to the runtime.
//
// The first local is a bool called retValid with a whole pointer-word reserved
// for it on the stack. The purpose of this word is so that the runtime knows
// whether the stack-allocated return space contains valid values for stack
// scanning.
//
// The second local is an abi.RegArgs value whose offset is also known to the
// runtime, so that a stack map for it can be constructed, since it contains
// pointers visible to the GC.
#define LOCAL_RETVALID 32
#define LOCAL_REGARGS 40

// makeFuncStub is the code half of the function returned by MakeFunc.
// See the comment on the declaration of makeFuncStub in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
// This frame contains two locals. See the comment above LOCAL_RETVALID.
TEXT ·makeFuncStub(SB),(NOSPLIT|WRAPPER),$312
	NO_LOCAL_POINTERS
	// NO_LOCAL_POINTERS is a lie. The stack map for the two locals in this
	// frame is specially handled in the runtime. See the comment above LOCAL_RETVALID.
	LEAQ	LOCAL_REGARGS(SP), R12 //leaq 计算 LOCAL_REGARGS(SP)的值, 把这个值给R12. 而movq是把这个值当地址来取内容再给R12.
	CALL	runtime·spillArgs(SB)      //src\runtime\asm_amd64.s:626 把寄存器值都写到R12表示的地址位置.一共184个字节.
	MOVQ	DX, 24(SP) // outside of moveMakeFuncArgPtrs's arg area
	MOVQ	DX, 0(SP)
	MOVQ	R12, 8(SP)
	CALL	·moveMakeFuncArgPtrs(SB) //src\reflect\makefunc.go:163 把寄存器整数信息,写入寄存器指针信息里.
	MOVQ	24(SP), DX
	MOVQ	DX, 0(SP)                       //callReflect参数ctxt
	LEAQ	argframe+0(FP), CX
	MOVQ	CX, 8(SP)                   //callReflect参数frame
	MOVB	$0, LOCAL_RETVALID(SP)
	LEAQ	LOCAL_RETVALID(SP), AX
	MOVQ	AX, 16(SP)               //callReflect参数retValid
	LEAQ	LOCAL_REGARGS(SP), AX
	MOVQ	AX, 24(SP)              //callReflect参数regs
	CALL	·callReflect(SB)        //调用上下文信息里面的函数, 一共有4个入参, callreflect在src\reflect\value.go:707 我们能看到他需要4个入参,就是刚才压入栈的.
	LEAQ	LOCAL_REGARGS(SP), R12
	CALL	runtime·unspillArgs(SB) //上面spill的逆操作
	RET

// methodValueCall is the code half of the function returned by makeMethodValue.
// See the comment on the declaration of methodValueCall in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
// This frame contains two locals. See the comment above LOCAL_RETVALID.
TEXT ·methodValueCall(SB),(NOSPLIT|WRAPPER),$312
	NO_LOCAL_POINTERS
	// NO_LOCAL_POINTERS is a lie. The stack map for the two locals in this
	// frame is specially handled in the runtime. See the comment above LOCAL_RETVALID.
	LEAQ	LOCAL_REGARGS(SP), R12
	CALL	runtime·spillArgs(SB)
	MOVQ	DX, 24(SP) // outside of moveMakeFuncArgPtrs's arg area
	MOVQ	DX, 0(SP)
	MOVQ	R12, 8(SP)
	CALL	·moveMakeFuncArgPtrs(SB)
	MOVQ	24(SP), DX
	MOVQ	DX, 0(SP)
	LEAQ	argframe+0(FP), CX
	MOVQ	CX, 8(SP)
	MOVB	$0, LOCAL_RETVALID(SP)
	LEAQ	LOCAL_RETVALID(SP), AX
	MOVQ	AX, 16(SP)
	LEAQ	LOCAL_REGARGS(SP), AX
	MOVQ	AX, 24(SP)
	CALL	·callMethod(SB)       //调用 src\reflect\value.go:964  运行这个函数,得到返回值.
	LEAQ	LOCAL_REGARGS(SP), R12
	CALL	runtime·unspillArgs(SB)
	RET
