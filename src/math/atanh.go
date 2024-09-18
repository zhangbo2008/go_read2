// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// The original C code, the long comment, and the constants
// below are from FreeBSD's /usr/src/lib/msun/src/e_atanh.c
// and came with this notice. The go code is a simplified
// version of the original C.  具体函数公式 atanh看这个https://ww2.mathworks.cn/help/matlab/ref/atanh.html
//
// ====================================================
// Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.
//
// Developed at SunPro, a Sun Microsystems, Inc. business.
// Permission to use, copy, modify, and distribute this
// software is freely granted, provided that this notice
// is preserved.
// ====================================================
//
// #下面我们给出计算公式!!!!!!!!!!!!!!!!!!!!!
// __ieee754_atanh(x)
// Method :
//	1. Reduce x to positive by atanh(-x) = -atanh(x) , 所以我们2里面专注于x大于0, 小宇1的情况.
//	2. For x>=0.5
//	            1              2x                          x
//	atanh(x) = --- * log(1 + -------) = 0.5 * log1p(2 * --------)-------(1)
//	            2             1 - x                      1 - x
//
//	For x<0.5
//	atanh(x) = 0.5*log1p(2x+2x*x/(1-x))----------------(2)
//  下面我们分析这2个情况.  首先tanh取值范围 -1,1 所以我们这里x 取值范围-1,1,只需要考虑正数,所以0到1之间, 这里面之所以这么分,  公式(1) 里面如果x无线接近1, 那么结果是一个x 的一阶/ x的一阶,  (2)里面是一个x的二阶/x的一阶. 相对来说肯定(1)的误差小. 注意float乘法有误差. 再看x<0.5, 假设x接近0, 那么1里面有一个除法,有误差.而2里面已经把2x这个大数项提出来了,所以计算得到的误差会比(1)小.
// Special cases:
//	atanh(x) is NaN if |x| > 1 with signal;
//	atanh(NaN) is that NaN with no signal;
//	atanh(+-1) is +-INF with signal.
//

// Atanh returns the inverse hyperbolic tangent of x.
//
// Special cases are:
//
//	Atanh(1) = +Inf
//	Atanh(±0) = ±0
//	Atanh(-1) = -Inf
//	Atanh(x) = NaN if x < -1 or x > 1
//	Atanh(NaN) = NaN
func Atanh(x float64) float64 {
	if haveArchAtanh {
		return archAtanh(x)
	}
	return atanh(x)
}

func atanh(x float64) float64 {
	const NearZero = 1.0 / (1 << 28) // 2**-28
	// special cases
	switch {
	case x < -1 || x > 1 || IsNaN(x):
		return NaN()
	case x == 1:
		return Inf(1)
	case x == -1:
		return Inf(-1)
	}
	sign := false
	if x < 0 {
		x = -x
		sign = true
	}
	var temp float64
	switch {
	case x < NearZero:
		temp = x
	case x < 0.5:
		temp = x + x
		temp = 0.5 * Log1p(temp+temp*x/(1-x))
	default:
		temp = 0.5 * Log1p((x+x)/(1-x))
	}
	if sign {
		temp = -temp
	}
	return temp
}
