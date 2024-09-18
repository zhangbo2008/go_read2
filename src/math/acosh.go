// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// The original C code, the long comment, and the constants
// below are from FreeBSD's /usr/src/lib/msun/src/e_acosh.c
// and came with this notice. The go code is a simplified
// version of the original C.
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
// //具体的函数定义acosh可以看 https://ww2.mathworks.cn/help/matlab/ref/acosh.html
// __ieee754_acosh(x)
// Method :
//	Based on
//	        acosh(x) = log [ x + sqrt(x*x-1) ]
//	we have
//	        acosh(x) := log(x)+ln2,	if x is large; else
//	        acosh(x) := log(2x-1/(sqrt(x*x-1)+x)) if x>2; else
//	        acosh(x) := log1p(t+sqrt(2.0*t+t*t)); where t=x-1.
// 上述公式的证明是trivial的,简单展开即可. 至于为什么这么计算, 是因为第一个误差足够忽略了,第二个是为了避免整出超界, 原始公式直接算x方会导致整数溢出. 变化后的公式, 平方在分母操作,即使整数溢出了,比如超出2的64次幂了.那么再取倒数,也不会跟真实结果差距过大. 所以是一个更加科学的计算方式, 这个例子就告诉我们,时刻要考虑大整数的计算范围, 至于第三个是因为log1p会有更快的计算优化.
// Special cases:
//	acosh(x) is NaN with signal if x<1.
//	acosh(NaN) is NaN without signal.
//

// Acosh returns the inverse hyperbolic cosine of x.
//
// Special cases are:
//
//	Acosh(+Inf) = +Inf
//	Acosh(x) = NaN if x < 1
//	Acosh(NaN) = NaN
func Acosh(x float64) float64 {
	if haveArchAcosh {
		return archAcosh(x)
	}
	return acosh(x)
}

func acosh(x float64) float64 {
	const Large = 1 << 28 // 2**28
	// first case is special case
	switch {
	case x < 1 || IsNaN(x):
		return NaN()
	case x == 1:
		return 0
	case x >= Large:
		return Log(x) + Ln2 // x > 2**28
	case x > 2:
		return Log(2*x - 1/(x+Sqrt(x*x-1))) // 2**28 > x > 2
	}
	t := x - 1
	return Log1p(t + Sqrt(2*t+t*t)) // 2 >= x > 1
}
