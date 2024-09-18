// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides Go implementations of elementary multi-precision
// arithmetic operations on word vectors. These have the suffix _g.
// These are needed for platforms without assembly implementations of these routines.
// This file also contains elementary operations that can be implemented
// sufficiently efficiently in Go.

package big

import "math/bits"

// A Word represents a single digit of a multi-precision unsigned integer.
type Word uint //一个64位无符号整数, 给这个起名Word, 赋予新的方法. 因为旧的算法不适合大数.精度也不够.

const (
	_S = _W / 8 // word size in bytes

	_W = bits.UintSize // word size in bits
	_B = 1 << _W       // digit base
	_M = _B - 1        // digit mask
)

// Many of the loops in this file are of the form
//   for i := 0; i < len(z) && i < len(x) && i < len(y); i++
// i < len(z) is the real condition.
// However, checking i < len(x) && i < len(y) as well is faster than
// having the compiler do a bounds check in the body of the loop;
// remarkably it is even faster than hoisting the bounds check
// out of the loop, by doing something like
//   _, _ = x[len(z)-1], y[len(z)-1]
// There are other ways to hoist the bounds check out of the loop,
// but the compiler's BCE isn't powerful enough for them (yet?).
// See the discussion in CL 164966.

// ----------------------------------------------------------------------------
// Elementary operations on words
//
// These operations are used by the vector operations below.

// z1<<_W + z0 = x*y
func mulWW(x, y Word) (z1, z0 Word) { // 两个Word相乘.
	hi, lo := bits.Mul(uint(x), uint(y)) // 两个64位数的乘法, 结果是一个128位的, 所以我们用两个64位来表示, 一个表示高64位, 一个表示低64位.
	return Word(hi), Word(lo)
}

// z1<<_W + z0 = x*y + c
func mulAddWWW_g(x, y, c Word) (z1, z0 Word) {
	hi, lo := bits.Mul(uint(x), uint(y))
	var cc uint
	lo, cc = bits.Add(lo, uint(c), 0)
	return Word(hi + cc), Word(lo)
}

// nlz returns the number of leading zeros in x.
// Wraps bits.LeadingZeros call for convenience.
func nlz(x Word) uint { // x bit表示里面前面有多少个0
	return uint(bits.LeadingZeros(uint(x)))
}

// The resulting carry c is either 0 or 1.// 这个函数 z=z+x+y 如果进位了那么c就设置为1,否则设置为0.
func addVV_g(z, x, y []Word) (c Word) { // V是vector的意思, 也就是 []Word,  使用的是小端表示.这点很重要, 使用小端表示能最大程度接近我们的普通四则运算法则!!!!!!!
	//例如:
	// 如果我们将0x1234abcd写入到以0x0000开始的内存中，则结果为；
	// address	big-endian	little-endian
	// 0x0000		0x12				0xcd
	// 0x0001		0x34				0xab
	// 0x0002		0xab				0x34
	// 0x0003		0xcd				0x12

	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x) && i < len(y); i++ {
		zi, cc := bits.Add(uint(x[i]), uint(y[i]), uint(c)) //go里面输出值,都默认初始化为0,
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

// The resulting carry c is either 0 or 1.
func subVV_g(z, x, y []Word) (c Word) { // 向量减法.  方法就是按位减, 从最高位开始减.
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x) && i < len(y); i++ {
		zi, cc := bits.Sub(uint(x[i]), uint(y[i]), uint(c))
		z[i] = Word(zi)
		c = Word(cc) //是否之前借了一位. 借位了,那么借位就参与下轮运算.借位表示最高位再往上借了一个1.  比如 x是   0...0 32个0,  y是 1...1 32个1,  那么x-y结果就是把x借一位更高位, x=10...0 32个0, 用这个减去y. 所以结果是1, 借位是1.
	}
	return
}

// The resulting carry c is either 0 or 1.
func addVW_g(z, x []Word, y Word) (c Word) { //
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		zi, cc := bits.Add(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

// addVWlarge is addVW, but intended for large z.
// The only difference is that we check on every iteration
// whether we are done with carries,
// and if so, switch to a much faster copy instead.
// This is only a good idea for large z,
// because the overhead of the check and the function call
// outweigh the benefits when z is small. //结果很大时候这个算法更快.
func addVWlarge(z, x []Word, y Word) (c Word) {
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		if c == 0 { //如果进位是0.那么我们直接copyx即可.
			copy(z[i:], x[i:])
			return
		}
		zi, cc := bits.Add(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

func subVW_g(z, x []Word, y Word) (c Word) {
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		zi, cc := bits.Sub(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

// subVWlarge is to subVW as addVWlarge is to addVW.
func subVWlarge(z, x []Word, y Word) (c Word) {
	c = y
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		if c == 0 {
			copy(z[i:], x[i:])
			return
		}
		zi, cc := bits.Sub(uint(x[i]), uint(c), 0)
		z[i] = Word(zi)
		c = Word(cc)
	}
	return
}

func shlVU_g(z, x []Word, s uint) (c Word) { // shl左移, V是 []word, U是uint缩写._g表示实现使用go语言.
	if s == 0 {
		copy(z, x)
		return
	}
	if len(z) == 0 {
		return
	}
	s &= _W - 1 // hint to the compiler that shifts by s don't need guard code
	ŝ := _W - s
	ŝ &= _W - 1                       // ditto   //s_hat是左移操作之后保留的位数
	c = x[len(z)-1] >> ŝ              // x的最后一个往右便宜. 举例子: x是64位的数, 我们要左移10位,那么我们先把这个x右移54位, 就得到了进位的值.c// 所以真实shl结果是z+c
	for i := len(z) - 1; i > 0; i-- { //因为我们使用的是小端表示,所以我们要从最大的索引开始.
		z[i] = x[i]<<s | x[i-1]>>ŝ //这个最好举例子来理解: 比如我们数据uint4[],  表示起来是0x1234abcd,  那么我们小端表示是 0xabcd, 0x1234,  左移结果是0x234abcd0, 第一步我们需要 1234左移1,然后跟0xabcd右移3的做or运算即可.举例子是理解bit运算的好方法.
	}
	z[0] = x[0] << s
	return
}

func shrVU_g(z, x []Word, s uint) (c Word) { //类似
	if s == 0 {
		copy(z, x)
		return
	}
	if len(z) == 0 {
		return
	}
	if len(x) != len(z) {
		// This is an invariant guaranteed by the caller.
		panic("len(x) != len(z)")
	}
	s &= _W - 1 // hint to the compiler that shifts by s don't need guard code
	ŝ := _W - s
	ŝ &= _W - 1 // ditto
	c = x[0] << ŝ
	for i := 1; i < len(z); i++ {
		z[i-1] = x[i-1]>>s | x[i]<<ŝ
	}
	z[len(z)-1] = x[len(z)-1] >> s
	return
}

func mulAddVWW_g(z, x []Word, y, r Word) (c Word) {
	c = r
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		c, z[i] = mulAddWWW_g(x[i], y, c)
	}
	return
}

func addMulVVW_g(z, x []Word, y Word) (c Word) {
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		z1, z0 := mulAddWWW_g(x[i], y, z[i])     //得到结果的高低位, 高位作为进位继续加
		lo, cc := bits.Add(uint(z0), uint(c), 0) //进位加上之前的进位c即可.
		c, z[i] = Word(cc), Word(lo)
		c += z1
	}
	return
}

// q = ( x1 << _W + x0 - r)/y. m = floor(( _B^2 - 1 ) / d - _B). Requiring x1<y.
// An approximate reciprocal with a reference to "Improved Division by Invariant Integers
// (IEEE Transactions on Computers, 11 Jun. 2010)"
func divWW(x1, x0, y, m Word) (q, r Word) { //带余除法
	s := nlz(y)
	if s != 0 {
		x1 = x1<<s | x0>>(_W-s)
		x0 <<= s
		y <<= s
	}
	d := uint(y)
	// We know that
	//   m = ⎣(B^2-1)/d⎦-B
	//   ⎣(B^2-1)/d⎦ = m+B
	//   (B^2-1)/d = m+B+delta1    0 <= delta1 <= (d-1)/d
	//   B^2/d = m+B+delta2        0 <= delta2 <= 1
	// The quotient we're trying to compute is
	//   quotient = ⎣(x1*B+x0)/d⎦
	//            = ⎣(x1*B*(B^2/d)+x0*(B^2/d))/B^2⎦
	//            = ⎣(x1*B*(m+B+delta2)+x0*(m+B+delta2))/B^2⎦
	//            = ⎣(x1*m+x1*B+x0)/B + x0*m/B^2 + delta2*(x1*B+x0)/B^2⎦
	// The latter two terms of this three-term sum are between 0 and 1.
	// So we can compute just the first term, and we will be low by at most 2.
	t1, t0 := bits.Mul(uint(m), uint(x1))
	_, c := bits.Add(t0, uint(x0), 0)
	t1, _ = bits.Add(t1, uint(x1), c)
	// The quotient is either t1, t1+1, or t1+2.
	// We'll try t1 and adjust if needed.
	qq := t1
	// compute remainder r=x-d*q.
	dq1, dq0 := bits.Mul(d, qq)
	r0, b := bits.Sub(uint(x0), dq0, 0)
	r1, _ := bits.Sub(uint(x1), dq1, b)
	// The remainder we just computed is bounded above by B+d:
	// r = x1*B + x0 - d*q.
	//   = x1*B + x0 - d*⎣(x1*m+x1*B+x0)/B⎦
	//   = x1*B + x0 - d*((x1*m+x1*B+x0)/B-alpha)                                   0 <= alpha < 1
	//   = x1*B + x0 - x1*d/B*m                         - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*⎣(B^2-1)/d-B⎦             - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*⎣(B^2-1)/d-B⎦             - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*((B^2-1)/d-B-beta)        - x1*d - x0*d/B + d*alpha   0 <= beta < 1
	//   = x1*B + x0 - x1*B + x1/B + x1*d + x1*d/B*beta - x1*d - x0*d/B + d*alpha
	//   =        x0        + x1/B        + x1*d/B*beta        - x0*d/B + d*alpha
	//   = x0*(1-d/B) + x1*(1+d*beta)/B + d*alpha
	//   <  B*(1-d/B) +  d*B/B          + d          because x0<B (and 1-d/B>0), x1<d, 1+d*beta<=B, alpha<1
	//   =  B - d     +  d              + d
	//   = B+d
	// So r1 can only be 0 or 1. If r1 is 1, then we know q was too small.
	// Add 1 to q and subtract d from r. That guarantees that r is <B, so
	// we no longer need to keep track of r1.
	if r1 != 0 {
		qq++
		r0 -= d
	}
	// If the remainder is still too large, increment q one more time.
	if r0 >= d {
		qq++
		r0 -= d
	}
	return Word(qq), Word(r0 >> s)
}

// reciprocalWord return the reciprocal of the divisor. rec = floor(( _B^2 - 1 ) / u - _B). u = d1 << nlz(d1).
func reciprocalWord(d1 Word) Word { // 计算d1的倒数
	u := uint(d1 << nlz(d1))
	x1 := ^u
	x0 := uint(_M)
	rec, _ := bits.Div(x1, x0, u) // (_B^2-1)/U-_B = (_B*(_M-C)+_M)/U
	return Word(rec)
}
