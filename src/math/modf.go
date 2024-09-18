// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Modf returns integer and fractional floating-point numbers
// that sum to f. Both values have the same sign as f.
//
// Special cases are:
//
//	Modf(±Inf) = ±Inf, NaN
//	Modf(NaN) = NaN, NaN
func Modf(f float64) (int float64, frac float64) {
	if haveArchModf {
		return archModf(f)
	}
	return modf(f)
}

func modf(f float64) (int float64, frac float64) {
	if f < 1 {
		switch {
		case f < 0:
			int, frac = Modf(-f)
			return -int, -frac
		case f == 0:
			return f, f // Return -0, -0 when f == -0
		}
		return 0, f
	}

	x := Float64bits(f)
	e := uint(x>>shift)&mask - bias //shift是IEEE754的float 尾数码. 这个操作算完, 得到真正的f的e的大小. 也就是10的多少次幂.

	// Keep the top 12+e bits, the integer part; clear the rest.
	if e < 64-12 {
		x &^= 1<<(64-12-e) - 1
	}
	int = Float64frombits(x)
	frac = f - int
	return
}
