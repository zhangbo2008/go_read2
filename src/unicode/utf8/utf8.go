// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package utf8 implements functions and constants to support text encoded in
// UTF-8. It includes functions to translate between runes and UTF-8 byte sequences.
// See https://en.wikipedia.org/wiki/UTF-8
package utf8

// The conditions RuneError==unicode.ReplacementChar and
// MaxRune==unicode.MaxRune are verified in the tests.
// Defining them locally avoids this package depending on package unicode.

// Numbers fundamental to the encoding.
const (
	RuneError = '\uFFFD'     // the "error" Rune or "Unicode replacement character"
	RuneSelf  = 0x80         // characters below RuneSelf are represented as themselves in a single byte.
	MaxRune   = '\U0010FFFF' // Maximum valid Unicode code point.
	UTFMax    = 4            // maximum number of bytes of a UTF-8 encoded Unicode character.
)

// Code points in the surrogate range are not valid for UTF-8.
const (
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
)

const (
	t1 = 0b00000000
	tx = 0b10000000
	t2 = 0b11000000
	t3 = 0b11100000
	t4 = 0b11110000
	t5 = 0b11111000

	maskx = 0b00111111
	mask2 = 0b00011111
	mask3 = 0b00001111
	mask4 = 0b00000111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1

	// The default lowest and highest continuation byte.
	locb = 0b10000000
	hicb = 0b10111111

	// These names of these constants are chosen to give nice alignment in the
	// table below. The first nibble is an index into acceptRanges or F for
	// special one-byte cases. The second nibble is the Rune length or the
	// Status for the special one-byte case.      acceptRanges 是下面代码定义的范围.
	xx = 0xF1 // invalid: size 1
	as = 0xF0 // ASCII: size 1      表示是一个ascii码
	s1 = 0x02 // accept 0, size 2
	s2 = 0x13 // accept 1, size 3
	s3 = 0x03 // accept 0, size 3
	s4 = 0x23 // accept 2, size 3
	s5 = 0x34 // accept 3, size 4
	s6 = 0x04 // accept 0, size 4
	s7 = 0x44 // accept 4, size 4
)

// first is information about the first byte in a UTF-8 sequence.
var first = [256]uint8{
	//   1   2   3   4   5   6   7   8   9   A   B   C   D   E   F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x00-0x0F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x10-0x1F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x20-0x2F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x30-0x3F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x40-0x4F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x50-0x5F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x60-0x6F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x70-0x7F
	//   1   2   3   4   5   6   7   8   9   A   B   C   D   E   F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0x80-0x8F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0x90-0x9F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xA0-0xAF
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xB0-0xBF
	xx, xx, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 0xC0-0xCF
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 0xD0-0xDF
	s2, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s4, s3, s3, // 0xE0-0xEF
	s5, s6, s6, s6, s7, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xF0-0xFF
}

// acceptRange gives the range of valid values for the second byte in a UTF-8
// sequence.
type acceptRange struct {
	lo uint8 // lowest value for second byte.
	hi uint8 // highest value for second byte.
}

// acceptRanges has size 16 to avoid bounds checks in the code that uses it.
var acceptRanges = [16]acceptRange{
	0: {locb, hicb},
	1: {0xA0, hicb},
	2: {locb, 0x9F},
	3: {0x90, hicb},
	4: {locb, 0x8F},
}

// FullRune reports whether the bytes in p begin with a full UTF-8 encoding of a rune.
// An invalid encoding is considered a full Rune since it will convert as a width-1 error rune.    这个函数判断是否p的头几个字符可以组成一个完整的utf-8的编码of a rune. rune是符文符号的意思.
func FullRune(p []byte) bool {
	n := len(p)
	if n == 0 {
		return false
	}
	x := first[p[0]] //x:第一个byte字符的信息
	if n >= int(x&7) {
		return true // ASCII, invalid or valid. 是合法的.
	}
	// Must be short or invalid.
	accept := acceptRanges[x>>4]                         //x是 两个16进制数, 所以我们>>4来得到第一个位置.得到他的范围.
	if n > 1 && (p[1] < accept.lo || accept.hi < p[1]) { // p长度大于1,那么我们第二个字符比较范围
		return true
	} else if n > 2 && (p[2] < locb || hicb < p[2]) { //p长度大于2,那么我们比第二个字符.
		return true
	}
	return false //其他情况就都不存在了.
}

// FullRuneInString is like FullRune but its input is a string. 同上.
func FullRuneInString(s string) bool {
	n := len(s)
	if n == 0 {
		return false
	}
	x := first[s[0]]
	if n >= int(x&7) {
		return true // ASCII, invalid, or valid.
	}
	// Must be short or invalid.
	accept := acceptRanges[x>>4]
	if n > 1 && (s[1] < accept.lo || accept.hi < s[1]) {
		return true
	} else if n > 2 && (s[2] < locb || hicb < s[2]) {
		return true
	}
	return false
}

// DecodeRune unpacks the first UTF-8 encoding in p and returns the rune and
// its width in bytes. If p is empty it returns ([RuneError], 0). Otherwise, if
// the encoding is invalid, it returns (RuneError, 1). Both are impossible
// results for correct, non-empty UTF-8.
//
// An encoding is invalid if it is incorrect UTF-8, encodes a rune that is
// out of range, or is not the shortest possible UTF-8 encoding for the
// value. No other validation is performed.//解析p他第一个utf-8合法的部分.
func DecodeRune(p []byte) (r rune, size int) {
	n := len(p)
	if n < 1 {
		return RuneError, 0
	}
	p0 := p[0]
	x := first[p0]
	if x >= as { // x=0xf0 or x=0xf1
		// The following code simulates an additional check for x == xx and
		// handling the ASCII and invalid cases accordingly. This mask-and-or
		// approach prevents an additional branch.
		mask := rune(x) << 31 >> 31                 // Create 0x0000 or 0xFFFF.  该行测试代码:	x := 0xf1  		println(int32(x) << 31 >> 31) // 11....1 32个 注意右移是补符号位的.
		return rune(p[0])&^mask | RuneError&mask, 1 //如果x是无效的,那么mask是0xffff,所以|前是0,所以返回|后面的 runeerror.  第二个返回恒1.
	}
	sz := int(x & 7)
	accept := acceptRanges[x>>4]
	if n < sz {
		return RuneError, 1
	}
	b1 := p[1]
	if b1 < accept.lo || accept.hi < b1 {
		return RuneError, 1
	}
	if sz <= 2 { // <= instead of == to help the compiler eliminate some bounds checks
		return rune(p0&mask2)<<6 | rune(b1&maskx), 2
	}
	b2 := p[2]
	if b2 < locb || hicb < b2 {
		return RuneError, 1
	}
	if sz <= 3 {
		return rune(p0&mask3)<<12 | rune(b1&maskx)<<6 | rune(b2&maskx), 3
	}
	b3 := p[3]
	if b3 < locb || hicb < b3 {
		return RuneError, 1
	}
	return rune(p0&mask4)<<18 | rune(b1&maskx)<<12 | rune(b2&maskx)<<6 | rune(b3&maskx), 4
}

// DecodeRuneInString is like [DecodeRune] but its input is a string. If s is
// empty it returns ([RuneError], 0). Otherwise, if the encoding is invalid, it
// returns (RuneError, 1). Both are impossible results for correct, non-empty
// UTF-8.
//
// An encoding is invalid if it is incorrect UTF-8, encodes a rune that is
// out of range, or is not the shortest possible UTF-8 encoding for the
// value. No other validation is performed.
func DecodeRuneInString(s string) (r rune, size int) { // 解析s的第一个rune, rune占用byte数记作size
	n := len(s)
	if n < 1 {
		return RuneError, 0
	}
	s0 := s[0]
	x := first[s0]
	if x >= as {
		// The following code simulates an additional check for x == xx and
		// handling the ASCII and invalid cases accordingly. This mask-and-or
		// approach prevents an additional branch.
		mask := rune(x) << 31 >> 31 // Create 0x0000 or 0xFFFF.
		return rune(s[0])&^mask | RuneError&mask, 1
	}
	sz := int(x & 7)
	accept := acceptRanges[x>>4]
	if n < sz {
		return RuneError, 1
	}
	s1 := s[1]
	if s1 < accept.lo || accept.hi < s1 {
		return RuneError, 1
	}
	if sz <= 2 { // <= instead of == to help the compiler eliminate some bounds checks
		return rune(s0&mask2)<<6 | rune(s1&maskx), 2
	}
	s2 := s[2]
	if s2 < locb || hicb < s2 {
		return RuneError, 1
	}
	if sz <= 3 {
		return rune(s0&mask3)<<12 | rune(s1&maskx)<<6 | rune(s2&maskx), 3
	}
	s3 := s[3]
	if s3 < locb || hicb < s3 {
		return RuneError, 1
	}
	return rune(s0&mask4)<<18 | rune(s1&maskx)<<12 | rune(s2&maskx)<<6 | rune(s3&maskx), 4
}

// DecodeLastRune unpacks the last UTF-8 encoding in p and returns the rune and
// its width in bytes. If p is empty it returns ([RuneError], 0). Otherwise, if
// the encoding is invalid, it returns (RuneError, 1). Both are impossible
// results for correct, non-empty UTF-8.
//
// An encoding is invalid if it is incorrect UTF-8, encodes a rune that is
// out of range, or is not the shortest possible UTF-8 encoding for the
// value. No other validation is performed. //从后往前解析.
func DecodeLastRune(p []byte) (r rune, size int) {
	end := len(p)
	if end == 0 {
		return RuneError, 0
	}
	start := end - 1
	r = rune(p[start]) //开始是最大索引.
	if r < RuneSelf {
		return r, 1
	}
	// guard against O(n^2) behavior when traversing
	// backwards through strings with long sequences of
	// invalid UTF-8.
	lim := end - UTFMax
	if lim < 0 {
		lim = 0
	}
	for start--; start >= lim; start-- {
		if RuneStart(p[start]) {
			break
		}
	}
	if start < 0 {
		start = 0
	}
	r, size = DecodeRune(p[start:end])
	if start+size != end {
		return RuneError, 1
	}
	return r, size
}

// DecodeLastRuneInString is like [DecodeLastRune] but its input is a string. If
// s is empty it returns ([RuneError], 0). Otherwise, if the encoding is invalid,
// it returns (RuneError, 1). Both are impossible results for correct,
// non-empty UTF-8.
//
// An encoding is invalid if it is incorrect UTF-8, encodes a rune that is
// out of range, or is not the shortest possible UTF-8 encoding for the
// value. No other validation is performed.
func DecodeLastRuneInString(s string) (r rune, size int) {
	end := len(s)
	if end == 0 {
		return RuneError, 0
	}
	start := end - 1
	r = rune(s[start])
	if r < RuneSelf {
		return r, 1
	}
	// guard against O(n^2) behavior when traversing
	// backwards through strings with long sequences of
	// invalid UTF-8.
	lim := end - UTFMax
	if lim < 0 {
		lim = 0
	}
	for start--; start >= lim; start-- {
		if RuneStart(s[start]) {
			break
		}
	}
	if start < 0 {
		start = 0
	}
	r, size = DecodeRuneInString(s[start:end])
	if start+size != end {
		return RuneError, 1
	}
	return r, size
}

// RuneLen returns the number of bytes required to encode the rune.
// It returns -1 if the rune is not a valid value to encode in UTF-8.
func RuneLen(r rune) int { //rune是一个unicode编码后的数字,我们返回这个数字解码后表示的字符串的长度.
	switch {
	case r < 0:
		return -1
	case r <= rune1Max:
		return 1
	case r <= rune2Max:
		return 2
	case surrogateMin <= r && r <= surrogateMax:
		return -1
	case r <= rune3Max:
		return 3
	case r <= MaxRune:
		return 4
	}
	return -1
}

// EncodeRune writes into p (which must be large enough) the UTF-8 encoding of the rune.
// If the rune is out of range, it writes the encoding of [RuneError].
// It returns the number of bytes written.
func EncodeRune(p []byte, r rune) int { // 把rune进行utf-8编码,rune本质是int32数字,编码之后是一个字符串. []byte类型.
	// Negative values are erroneous. Making it unsigned addresses the problem.
	switch i := uint32(r); { //因为utf8没有负数编码,所以我们强行转化uint32,负数直接变超大整数.用后续代码判断即可屏蔽掉负数.  addresses the problem 处理问题.
	case i <= rune1Max:
		p[0] = byte(r)
		return 1
	case i <= rune2Max:
		_ = p[1] // eliminate bounds checks// 我们传入p是一个数组, 调用p[1]之前先进行边缘检测. 看看p[1]是否合法. go里面把他赋值给_ 即可检测他是否合法. 不合法go会自动panic. 省略了很多 index<len(str)的代码.
		p[0] = t2 | byte(r>>6)
		p[1] = tx | byte(r)&maskx
		return 2
	case i > MaxRune, surrogateMin <= i && i <= surrogateMax:
		r = RuneError
		fallthrough
	case i <= rune3Max:
		_ = p[2] // eliminate bounds checks
		p[0] = t3 | byte(r>>12)
		p[1] = tx | byte(r>>6)&maskx
		p[2] = tx | byte(r)&maskx
		return 3
	default:
		_ = p[3] // eliminate bounds checks
		p[0] = t4 | byte(r>>18)
		p[1] = tx | byte(r>>12)&maskx
		p[2] = tx | byte(r>>6)&maskx
		p[3] = tx | byte(r)&maskx
		return 4
	}
}

// AppendRune appends the UTF-8 encoding of r to the end of p and
// returns the extended buffer. If the rune is out of range,
// it appends the encoding of [RuneError].
func AppendRune(p []byte, r rune) []byte { //把r append到p后面
	// This function is inlineable for fast handling of ASCII.
	if uint32(r) <= rune1Max {
		return append(p, byte(r))
	}
	return appendRuneNonASCII(p, r)
}

func appendRuneNonASCII(p []byte, r rune) []byte {
	// Negative values are erroneous. Making it unsigned addresses the problem.
	switch i := uint32(r); {
	case i <= rune2Max:
		return append(p, t2|byte(r>>6), tx|byte(r)&maskx)
	case i > MaxRune, surrogateMin <= i && i <= surrogateMax:
		r = RuneError
		fallthrough
	case i <= rune3Max:
		return append(p, t3|byte(r>>12), tx|byte(r>>6)&maskx, tx|byte(r)&maskx)
	default:
		return append(p, t4|byte(r>>18), tx|byte(r>>12)&maskx, tx|byte(r>>6)&maskx, tx|byte(r)&maskx)
	}
}

// RuneCount returns the number of runes in p. Erroneous and short
// encodings are treated as single runes of width 1 byte.
func RuneCount(p []byte) int { // rune本质就是一个32位整数, byte[]本质是8位整数数组. 他们都可以表示一个字符, byte[]表示一个字符或者一个字符串. 这个函数RunecCount就是给一个byte[]字符串.然后计算他能转为rune表示的多少个字符.  r 32位,表示一个 utf8
	np := len(p)
	var n int
	for i := 0; i < np; {
		n++
		c := p[i]
		if c < RuneSelf {
			// ASCII fast path  //如果是单个byte成为一个rune
			i++
			continue
		}
		x := first[c]
		if x == xx {
			i++ // invalid.
			continue
		}
		size := int(x & 7)
		if i+size > np {
			i++ // Short or invalid. //长度不够拼接一个的. 最后加上一个invalid.
			continue
		}
		accept := acceptRanges[x>>4] //根据范围计算合法性.
		if c := p[i+1]; c < accept.lo || accept.hi < c {
			size = 1
		} else if size == 2 {
		} else if c := p[i+2]; c < locb || hicb < c {
			size = 1
		} else if size == 3 {
		} else if c := p[i+3]; c < locb || hicb < c {
			size = 1
		} //计算size
		i += size
	}
	return n
}

// RuneCountInString is like [RuneCount] but its input is a string.
func RuneCountInString(s string) (n int) { //跟上面代码一样.
	ns := len(s)
	for i := 0; i < ns; n++ {
		c := s[i]
		if c < RuneSelf {
			// ASCII fast path
			i++
			continue
		}
		x := first[c]
		if x == xx {
			i++ // invalid.
			continue
		}
		size := int(x & 7)
		if i+size > ns {
			i++ // Short or invalid.
			continue
		}
		accept := acceptRanges[x>>4]
		if c := s[i+1]; c < accept.lo || accept.hi < c {
			size = 1
		} else if size == 2 {
		} else if c := s[i+2]; c < locb || hicb < c {
			size = 1
		} else if size == 3 {
		} else if c := s[i+3]; c < locb || hicb < c {
			size = 1
		}
		i += size
	}
	return n
}

// RuneStart reports whether the byte could be the first byte of an encoded,
// possibly invalid rune. Second and subsequent bytes always have the top two
// bits set to 10.
func RuneStart(b byte) bool { return b&0xC0 != 0x80 } //判断b是否为一个合法rune的第一个byte部分.

// UTF-8编码规则如下：

// 单字节字符（ASCII字符）：0xxxxxxx（最高位为0）
// 双字节字符：110xxxxx 10xxxxxx
// 三字节字符：1110xxxx 10xxxxxx 10xxxxxx
// 四字节字符：11110xxx 10xxxxxx 10xxxxxx 10xxxxxx  所以我们知道一个byte是非ascii非后续字节才是我们要的前缀字节.  10xxxxxx:这种叫后缀字节.

// Valid reports whether p consists entirely of valid UTF-8-encoded runes.
func Valid(p []byte) bool { //这个函数返回p是否是一个合法的utf8编码组成的byte数组.
	// This optimization avoids the need to recompute the capacity
	// when generating code for p[8:], bringing it to parity with
	// ValidString, which was 20% faster on long ASCII strings.
	p = p[:len(p):len(p)] // 两个冒号切片: a[x:y:z] 切片内容是： [x:y] 切片长度，[x:z] 切片容量。长度计算：y-x，容量计算：z-x。

	// Fast path. Check for and skip 8 bytes of ASCII characters per iteration.
	for len(p) >= 8 {
		// Combining two 32 bit loads allows the same code to be used
		// for 32 and 64 bit platforms.
		// The compiler can generate a 32bit load for first32 and second32
		// on many platforms. See test/codegen/memcombine.go.
		first32 := uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16 | uint32(p[3])<<24
		second32 := uint32(p[4]) | uint32(p[5])<<8 | uint32(p[6])<<16 | uint32(p[7])<<24
		//拼出来2个32位的整数
		if (first32|second32)&0x80808080 != 0 {
			// Found a non ASCII byte (>= RuneSelf). //一个ascii码是256长度,是8位. first32|second32 是4个ascii码, 每一个ascii码跟0x80做&运算. 大于0x80的跟0x80做&肯定非0,从而找到了非ascii吗,找到非ascii码我们就停下来. 我们通过这个循环过滤掉ascii码, 因为ascii码比较简单0-127都是合法的.直接跳过即可.
			break
		}
		p = p[8:]
	}
	n := len(p)
	for i := 0; i < n; { //这段跟之前代码很类似了.解析utf8编码,如果走到最后说明合法.返回true
		pi := p[i]
		if pi < RuneSelf {
			i++
			continue
		}
		x := first[pi]
		if x == xx {
			return false // Illegal starter byte.
		}
		size := int(x & 7)
		if i+size > n {
			return false // Short or invalid.
		}
		accept := acceptRanges[x>>4]
		if c := p[i+1]; c < accept.lo || accept.hi < c {
			return false
		} else if size == 2 {
		} else if c := p[i+2]; c < locb || hicb < c {
			return false
		} else if size == 3 {
		} else if c := p[i+3]; c < locb || hicb < c {
			return false
		}
		i += size
	}
	return true
}

// ValidString reports whether s consists entirely of valid UTF-8-encoded runes.
func ValidString(s string) bool { //同上, 跳过.
	// Fast path. Check for and skip 8 bytes of ASCII characters per iteration.
	for len(s) >= 8 {
		// Combining two 32 bit loads allows the same code to be used
		// for 32 and 64 bit platforms.
		// The compiler can generate a 32bit load for first32 and second32
		// on many platforms. See test/codegen/memcombine.go.
		first32 := uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24
		second32 := uint32(s[4]) | uint32(s[5])<<8 | uint32(s[6])<<16 | uint32(s[7])<<24
		if (first32|second32)&0x80808080 != 0 {
			// Found a non ASCII byte (>= RuneSelf).
			break
		}
		s = s[8:]
	}
	n := len(s)
	for i := 0; i < n; {
		si := s[i]
		if si < RuneSelf {
			i++
			continue
		}
		x := first[si]
		if x == xx {
			return false // Illegal starter byte.
		}
		size := int(x & 7)
		if i+size > n {
			return false // Short or invalid.
		}
		accept := acceptRanges[x>>4]
		if c := s[i+1]; c < accept.lo || accept.hi < c {
			return false
		} else if size == 2 {
		} else if c := s[i+2]; c < locb || hicb < c {
			return false
		} else if size == 3 {
		} else if c := s[i+3]; c < locb || hicb < c {
			return false
		}
		i += size
	}
	return true
}

// ValidRune reports whether r can be legally encoded as UTF-8.
// Code points that are out of range or a surrogate half are illegal.
func ValidRune(r rune) bool { //r 是否是一个合法utf8编码,根据整数判断范围即可.
	switch {
	case 0 <= r && r < surrogateMin:
		return true
	case surrogateMax < r && r <= MaxRune:
		return true
	}
	return false
}
