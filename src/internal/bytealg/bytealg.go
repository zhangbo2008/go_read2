// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

import (
	"internal/cpu"
	"unsafe"
)

// Offsets into internal/cpu records for use in assembly.
const ( //获取cpu的配置信息.
	offsetX86HasSSE42  = unsafe.Offsetof(cpu.X86.HasSSE42)
	offsetX86HasAVX2   = unsafe.Offsetof(cpu.X86.HasAVX2)
	offsetX86HasPOPCNT = unsafe.Offsetof(cpu.X86.HasPOPCNT)

	offsetS390xHasVX = unsafe.Offsetof(cpu.S390X.HasVX)

	offsetPPC64HasPOWER9 = unsafe.Offsetof(cpu.PPC64.IsPOWER9)
)

// MaxLen is the maximum length of the string to be searched for (argument b) in Index.
// If MaxLen is not 0, make sure MaxLen >= 4.
var MaxLen int

// PrimeRK is the prime base used in Rabin-Karp algorithm.  Rabin-Karp算法，是由M.O.Rabin和R.A.Karp发明的一种基于散列的字符串查找算法。
const PrimeRK = 16777619

// HashStr returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func HashStr[T string | []byte](sep T) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*PrimeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, PrimeRK
	for i := len(sep); i > 0; i >>= 1 { //字符串长度这个整数作为次幂. primeRK作为底的值.记作pow
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// HashStrRev returns the hash of the reverse of sep and the
// appropriate multiplicative factor for use in Rabin-Karp algorithm.
func HashStrRev[T string | []byte](sep T) (uint32, uint32) { // 这里面T是泛型, 表示sep这个变量, 可以是string, 也可以是[]byte. 然后我们返回他的哈希值, 和pow 两个大整数.
	hash := uint32(0)
	for i := len(sep) - 1; i >= 0; i-- {
		hash = hash*PrimeRK + uint32(sep[i]) //字符串的二进制看做一个整数.然后进行哈希. 哈希算法是每次旧的乘以大素数, 加上新的位置的值.
	}
	var pow, sq uint32 = 1, PrimeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	} //得到pow=PrimeRK**uint(len(sep))
	return hash, pow
}

// IndexRabinKarp uses the Rabin-Karp search algorithm to return the index of the
// first occurrence of sep in s, or -1 if not present.
func IndexRabinKarp[T string | []byte](s, sep T) int { //计算s中sep第一次出现时候的索引.
	// Rabin-Karp search
	hashss, pow := HashStr(sep)
	n := len(sep)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*PrimeRK + uint32(s[i])
	}
	if h == hashss && string(s[:n]) == string(sep) {
		return 0
	}
	for i := n; i < len(s); {
		h *= PrimeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashss && string(s[i-n:i]) == string(sep) {
			return i - n
		}
	}
	return -1
}

// LastIndexRabinKarp uses the Rabin-Karp search algorithm to return the last index of the
// occurrence of sep in s, or -1 if not present.
func LastIndexRabinKarp[T string | []byte](s, sep T) int {
	// Rabin-Karp search from the end of the string
	hashss, pow := HashStrRev(sep)
	n := len(sep)
	last := len(s) - n
	var h uint32
	for i := len(s) - 1; i >= last; i-- {
		h = h*PrimeRK + uint32(s[i])
	}
	if h == hashss && string(s[last:]) == string(sep) {
		return last
	}
	for i := last - 1; i >= 0; i-- {
		h *= PrimeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i+n])
		if h == hashss && string(s[i:i+n]) == string(sep) {
			return i
		}
	}
	return -1
}

// MakeNoZero makes a slice of length and capacity n without zeroing the bytes.
// It is the caller's responsibility to ensure uninitialized bytes
// do not leak to the end user.
func MakeNoZero(n int) []byte //这个函数输入一个n, 返回一个byte的长度为n的切片.不对他进行零化, 至于底层怎么实现的没写.就一个函数名. 可能在其他的汇编文件中.
