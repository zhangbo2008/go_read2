// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !386 && !amd64 && !s390x && !arm && !arm64 && !loong64 && !ppc64 && !ppc64le && !mips && !mipsle && !wasm && !mips64 && !mips64le && !riscv64

package bytealg

import _ "unsafe" // for go:linkname

func Compare(a, b []byte) int { //返回a,b字典序的大小关系.
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	if l == 0 || &a[0] == &b[0] { //假设a是字符串c的切片[:3] b是字符串c的切片[:4] , 那么&a[0] == &b[0], 这行说明a,b的首地址一样, 说明他俩底层都对应同一片数组, 所以后面进入samebytes判断.只需要比较他俩切片的长度即可.长的大. //写成a[0]==b[0]显然不对,这样判断不了地址底层是同一个. // ps: 这个地方16行做这个判断是为了加速. 当地址一样时候没必要挨个像20行那样逐个比较了.16行可以看做20行的性能优化加速.//这个思路可以用在很多容器的大小比较上.
		goto samebytes
	}
	for i := 0; i < l; i++ { //a,b等长的部分比较完.
		c1, c2 := a[i], b[i]
		if c1 < c2 {
			return -1
		}
		if c1 > c2 {
			return +1
		}
	}
samebytes:
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return +1
	}
	return 0
}

//go:linkname runtime_cmpstring runtime.cmpstring
func runtime_cmpstring(a, b string) int { //这就是上面函数少了16行加速的版本.因为string底层跟byte[]不同.string不存在切片这个操作.
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	for i := 0; i < l; i++ {
		c1, c2 := a[i], b[i]
		if c1 < c2 {
			return -1
		}
		if c1 > c2 {
			return +1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return +1
	}
	return 0
}
