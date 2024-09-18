// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package utf16 implements encoding and decoding of UTF-16 sequences.
package utf16

// The conditions replacementChar==unicode.ReplacementChar and
// maxRune==unicode.MaxRune are verified in the tests.
// Defining them locally avoids this package depending on package unicode.

const (
	replacementChar = '\uFFFD'     // Unicode replacement character //作为不合法的表示.
	maxRune         = '\U0010FFFF' // Maximum valid Unicode code point.
)

const ( //解码的边界值.
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000
)

// IsSurrogate reports whether the specified Unicode code point
// can appear in a surrogate pair.
func IsSurrogate(r rune) bool {
	return surr1 <= r && r < surr3
}

// DecodeRune returns the UTF-16 decoding of a surrogate pair.
// If the pair is not a valid UTF-16 surrogate pair, DecodeRune returns
// the Unicode replacement code point U+FFFD.
func DecodeRune(r1, r2 rune) rune { // 一个utf16使用2个rune来表示叫surrogate pair,解析之后是一个rune.
	if surr1 <= r1 && r1 < surr2 && surr2 <= r2 && r2 < surr3 {
		return (r1-surr1)<<10 | (r2 - surr2) + surrSelf //高十位, 低十位, 加上surrself
	}
	return replacementChar
}

// EncodeRune returns the UTF-16 surrogate pair r1, r2 for the given rune.
// If the rune is not a valid Unicode code point or does not need encoding,
// EncodeRune returns U+FFFD, U+FFFD.
func EncodeRune(r rune) (r1, r2 rune) {
	if r < surrSelf || r > maxRune {
		return replacementChar, replacementChar
	}
	r -= surrSelf
	return surr1 + (r>>10)&0x3ff, surr2 + r&0x3ff //这里之所以使用0x3ff是因为她=bin(十个1). 作为两个数位的切分.等于只保留最低的十位.
}

// Encode returns the UTF-16 encoding of the Unicode code point sequence s.
func Encode(s []rune) []uint16 { // uint16是16位. utf16是用一个32位编码的.所以两个uint16组成一个utf16的rune.
	n := len(s)
	for _, v := range s {
		if v >= surrSelf { //大于这个surrsefl的才是一个合法的2个unit16表示,所以空间n++即可.计算出总共空间n了.
			n++
		}
	}

	a := make([]uint16, n)
	n = 0
	for _, v := range s {
		switch {
		case 0 <= v && v < surr1, surr3 <= v && v < surrSelf:
			// normal rune
			a[n] = uint16(v)
			n++
		case surrSelf <= v && v <= maxRune:
			// needs surrogate sequence
			r1, r2 := EncodeRune(v)
			a[n] = uint16(r1)
			a[n+1] = uint16(r2)
			n += 2
		default:
			a[n] = uint16(replacementChar)
			n++
		}
	}
	return a[:n]
}

// AppendRune appends the UTF-16 encoding of the Unicode code point r
// to the end of p and returns the extended buffer. If the rune is not
// a valid Unicode code point, it appends the encoding of U+FFFD.
func AppendRune(a []uint16, r rune) []uint16 { //同上
	// This function is inlineable for fast handling of ASCII.
	switch {
	case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
		// normal rune
		return append(a, uint16(r))
	case surrSelf <= r && r <= maxRune:
		// needs surrogate sequence
		r1, r2 := EncodeRune(r)
		return append(a, uint16(r1), uint16(r2))
	}
	return append(a, replacementChar)
}

// Decode returns the Unicode code point sequence represented
// by the UTF-16 encoding s.
func Decode(s []uint16) []rune {
	// Preallocate capacity to hold up to 64 runes.
	// Decode inlines, so the allocation can live on the stack.
	buf := make([]rune, 0, 64)
	return decode(s, buf)
}

// decode appends to buf the Unicode code point sequence represented
// by the UTF-16 encoding s and return the extended buffer.
func decode(s []uint16, buf []rune) []rune { //同上.
	for i := 0; i < len(s); i++ {
		var ar rune
		switch r := s[i]; {
		case r < surr1, surr3 <= r:
			// normal rune
			ar = rune(r)
		case surr1 <= r && r < surr2 && i+1 < len(s) &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence
			ar = DecodeRune(rune(r), rune(s[i+1]))
			i++
		default:
			// invalid surrogate sequence
			ar = replacementChar
		}
		buf = append(buf, ar)
	}
	return buf
}
