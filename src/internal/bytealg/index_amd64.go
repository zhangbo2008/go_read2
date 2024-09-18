// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

import "internal/cpu"

const MaxBruteForce = 64

func init() {
	if cpu.X86.HasAVX2 {
		MaxLen = 63
	} else {
		MaxLen = 31
	}
}

// Cutover reports the number of failures of IndexByte we should tolerate
// before switching over to Index.
// n is the number of bytes processed so far.
// See the bytes.Index implementation for details.
func Cutover(n int) int { //在找字符串子串时候. 我们容错率是 8个字符, 错一个字符.再加点偏移量.这是经验值.通过实验获得的大小.
	// 1 error per 8 characters, plus a few slop to start.
	return (n + 16) / 8
}
