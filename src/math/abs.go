// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Abs returns the absolute value of x.
//
// Special cases are:
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func Abs(x float64) float64 { //目的就是去掉符号位. 1<<63就是int64的最高位也就是符号位,  取反之后就是 011...1111 然后跟x的二进制做交集. 就等于把二进制的x最高位变成0了.
	return Float64frombits(Float64bits(x) &^ (1 << 63))
}
