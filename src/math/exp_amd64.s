// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// The method is based on a paper by Naoki Shibata: "Efficient evaluation
// methods of elementary functions suitable for SIMD computation", Proc.
// of International Supercomputing Conference 2010 (ISC'10), pp. 25 -- 32
// (May 2010). The paper is available at
// https://link.springer.com/article/10.1007/s00450-010-0108-2
//
// The original code and the constants below are from the author's
// implementation available at http://freshmeat.net/projects/sleef.
// The README file says, "The software is in public domain.
// You can use the software without any obligation."
//
// This code is a simplified version of the original.

#define LN2 0.6931471805599453094172321214581766 // log_e(2)
#define LOG2E 1.4426950408889634073599246810018920 // 1/LN2
#define LN2U 0.69314718055966295651160180568695068359375 // upper half LN2
#define LN2L 0.28235290563031577122588448175013436025525412068e-12 // lower half LN2
#define PosInf 0x7FF0000000000000  //这是一个64位的值=8bytes
#define NegInf 0xFFF0000000000000
#define Overflow 7.09782712893384e+02   //e的709次幂过大,无法计算.

DATA exprodata<>+0(SB)/8, $0.5           //DATA命令用于往SB里面放全局变量. $后面加数字表示立即数, 加变量表示他的地址.
DATA exprodata<>+8(SB)/8, $1.0
DATA exprodata<>+16(SB)/8, $2.0
DATA exprodata<>+24(SB)/8, $1.6666666666666666667e-1  //这里面float都是用8字节来存.
DATA exprodata<>+32(SB)/8, $4.1666666666666666667e-2
DATA exprodata<>+40(SB)/8, $8.3333333333333333333e-3
DATA exprodata<>+48(SB)/8, $1.3888888888888888889e-3
DATA exprodata<>+56(SB)/8, $1.9841269841269841270e-4
DATA exprodata<>+64(SB)/8, $2.4801587301587301587e-5
GLOBL exprodata<>+0(SB), RODATA, $72       //最后72是长度, 可以看到上面一共有9个8

// func Exp(x float64) float64
TEXT ·archExp(SB),NOSPLIT,$0
	// test bits for not-finite
	MOVQ    x+0(FP), BX   //movb（8位）字节、movw（16位）字、movl（32位）双字、movq（64位）四字 经常用到,要记住. 因为我们这个函数接口原型在src\math\exp_asm.go :  func archExp(x float64) float64  输入64位float, 输出64位float. 所以掉archExp, 数据存到BX里面.
	MOVQ    $~(1<<63), AX // sign bit mask
	MOVQ    BX, DX
	ANDQ    AX, DX          //DX得到x的绝对值
	MOVQ    $PosInf, AX
	CMPQ    AX, DX
	JLE     notFinite      // jump less equal 如果上面结果是小于等于,那么我们就跳转到 notFinite标致.
	// check if argument will overflow
	MOVQ    BX, X0
	MOVSD   $Overflow, X1   //传送string数据的双字节.但是根据这个代码看出来movsd是传输8字节.
	COMISD  X1, X0         // comisd指令: Compares the double precision floating-point values in the low quadwords of operand 1 (first operand) and operand 2 (second operand), and sets the ZF, PF, and CF flags in the EFLAGS register according to the result (unordered, greater than, less than, or equal). The OF, SF, and AF flags in the EFLAGS register are set to 0. The unordered result is returned if either source operand is a NaN (QNaN or SNaN). 比较2个8字节的float数. 这个运算符是反过来算, 比较X0跟X1的关系. 不是cmp里面的比较顺序.
	JA      overflow       //JA   ;无符号大于则跳转
	MOVSD   $LOG2E, X1         //x1=1/ln2
	MULSD   X0, X1                  // x1=x/ln2
	CVTSD2SL X1, BX // BX = exponent    
	CVTSL2SD BX, X1
	CMPB ·useFMA(SB), $1          //是否用fma来运算
	JE   avxfma
	MOVSD   $LN2U, X2
	MULSD   X1, X2              //x1是代码exp.go里面的k了.  BX也是k
	SUBSD   X2, X0              //exp.go:132 //结果存入X2
	MOVSD   $LN2L, X2
	MULSD   X1, X2
	SUBSD   X2, X0
	// reduce argument
	MULSD   $0.0625, X0
	// Taylor series evaluation
	MOVSD   exprodata<>+64(SB), X1
	MULSD   X0, X1
	ADDSD   exprodata<>+56(SB), X1
	MULSD   X0, X1
	ADDSD   exprodata<>+48(SB), X1
	MULSD   X0, X1
	ADDSD   exprodata<>+40(SB), X1
	MULSD   X0, X1
	ADDSD   exprodata<>+32(SB), X1
	MULSD   X0, X1
	ADDSD   exprodata<>+24(SB), X1
	MULSD   X0, X1
	ADDSD   exprodata<>+0(SB), X1
	MULSD   X0, X1
	ADDSD   exprodata<>+8(SB), X1
	MULSD   X1, X0
	MOVSD   exprodata<>+16(SB), X1
	ADDSD   X0, X1
	MULSD   X1, X0
	MOVSD   exprodata<>+16(SB), X1
	ADDSD   X0, X1
	MULSD   X1, X0
	MOVSD   exprodata<>+16(SB), X1
	ADDSD   X0, X1
	MULSD   X1, X0
	MOVSD   exprodata<>+16(SB), X1
	ADDSD   X0, X1
	MULSD   X1, X0
	ADDSD exprodata<>+8(SB), X0
	// return fr * 2**exponent
ldexp:
	ADDL    $0x3FF, BX // add bias   $0x3ff是源, BX是目标. 所以BX+0x3ff
	JLE     denormal  //The JLE (Jump if Less Than or Equal) instruction checks if the result of the addition is less than or equal to zero.
	CMPL    BX, $0x7FF
	JGE     overflow
lastStep:
	SHLQ    $52, BX
	MOVQ    BX, X1
	MULSD   X1, X0
	MOVSD   X0, ret+8(FP)
	RET
notFinite:
	// test bits for -Inf   再判断他是不是负无穷.
	MOVQ    $NegInf, AX    //$表示取变量的值.  
	CMPQ    AX, BX
	JNE     notNegInf         //跳到非负无穷
	// -Inf, return 0
underflow: // return 0
	MOVQ    $0, ret+8(FP)
	RET
overflow: // return +Inf
	MOVQ    $PosInf, BX   //这行运行完就运行121行.这种设置也就是导致这种flag跳转很容易bug, 跟c语言goto一样.
notNegInf: // NaN or +Inf, return x
	MOVQ    BX, ret+8(FP)  //这个偏移量,可以看汇编的内存图.   从上往下的地址是 返回值, 入参, 局部变量. 之所以这么排列是因为函数最后返回时候, 要栈pop,要pop到返回值的地址,所以返回值一定在最上面, 才保证把下面已经无用的信息都pop干净了. 之后压入的是入参,  因为入参是父函数给的,所以要提前压入, 局部变量是子函数启动时候才有的,所以最后进入.这里面偏移量是8, 是因为我们函数压入的是float64. 所以是8bytes的. 这样下行ret就把bx里面的值返回了.
	RET
denormal:
	CMPL    BX, $-52
	JL      underflow
	ADDL    $0x3FE, BX // add bias - 1
	SHLQ    $52, BX
	MOVQ    BX, X1
	MULSD   X1, X0
	MOVQ    $1, BX
	JMP     lastStep

avxfma:
	MOVSD   $LN2U, X2
	VFNMADD231SD X2, X1, X0
	MOVSD   $LN2L, X2
	VFNMADD231SD X2, X1, X0
	// reduce argument
	MULSD   $0.0625, X0
	// Taylor series evaluation
	MOVSD   exprodata<>+64(SB), X1
	VFMADD213SD exprodata<>+56(SB), X0, X1
	VFMADD213SD exprodata<>+48(SB), X0, X1
	VFMADD213SD exprodata<>+40(SB), X0, X1
	VFMADD213SD exprodata<>+32(SB), X0, X1
	VFMADD213SD exprodata<>+24(SB), X0, X1
	VFMADD213SD exprodata<>+0(SB), X0, X1
	VFMADD213SD exprodata<>+8(SB), X0, X1
	MULSD   X1, X0
	VADDSD exprodata<>+16(SB), X0, X1
	MULSD   X1, X0
	VADDSD exprodata<>+16(SB), X0, X1
	MULSD   X1, X0
	VADDSD exprodata<>+16(SB), X0, X1
	MULSD   X1, X0
	VADDSD exprodata<>+16(SB), X0, X1
	VFMADD213SD   exprodata<>+8(SB), X1, X0
	JMP ldexp
