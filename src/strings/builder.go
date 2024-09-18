// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"internal/bytealg"
	"unicode/utf8"
	"unsafe"
)

// A Builder is used to efficiently build a string using [Builder.Write] methods.
// It minimizes memory copying. The zero value is ready to use.
// Do not copy a non-zero Builder.
type Builder struct {
	addr *Builder // of receiver, to detect copies by value //用来做一些检测
	buf  []byte   //用来存实际数据.
}

// noescape hides a pointer from escape analysis. It is the identity function
// but escape analysis doesn't think the output depends on the input.
// noescape is inlined and currently compiles down to zero instructions.
// USE CAREFULLY!
// This was copied from the runtime; see issues 23382 and 7921.
//
// 逃逸分析是Go编译器在编译时执行的一个过程，用于确定一个变量是否“逃逸”出了其原始的作用域。简单来说，如果一个变量在函数返回后仍然需要被引用，那么它就发生了逃逸，编译器会将其分配到堆上，而不是栈上。
//
//go:nosplit
//go:nocheckptr        //escape analysis : 什么是逃逸分析？
func noescape(p unsafe.Pointer) unsafe.Pointer { //这个函数可以避免逃逸分析.优化代码速度. 减少不必要的堆分配.
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0) //go里面x^y 是按位异或,  ^x是对x按位取反. 对于任意x, x^0=x. 所以整个函数就是一个恒等函数, 输入p, 输出也是p,但是经过uintptr转化了, 编译器理解为输出不依赖输入了.不会让go进行逃逸分析了.
}

func (b *Builder) copyCheck() {
	if b.addr == nil {
		// This hack works around a failing of Go's escape analysis
		// that was causing b to escape and be heap allocated.
		// See issue 23382.
		// TODO: once issue 7921 is fixed, this should be reverted to
		// just "b.addr = b".
		b.addr = (*Builder)(noescape(unsafe.Pointer(b))) //这个就是上面的函数的使用方法,用来避免逃逸分析带来的对分配. 旧的写法就是b.addr = b. 但是这么写让b的addr指向自己.所以返回值的b依赖b本身了.所以只能放在堆上面了. 现在这么写,逃逸分析就不起作用了. b还是在栈上.节省内存了.这是uintptr的特性.
	} else if b.addr != b {
		panic("strings: illegal use of non-zero Builder copied by value")
	}
}

// String returns the accumulated string.
func (b *Builder) String() string { //把b的数据也就是buf字段进行string化.
	return unsafe.String(unsafe.SliceData(b.buf), len(b.buf))
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *Builder) Len() int { return len(b.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *Builder) Cap() int { return cap(b.buf) } //对于builder来说cap=len

// Reset resets the [Builder] to be empty.
func (b *Builder) Reset() {
	b.addr = nil
	b.buf = nil
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf). //这个函数至少向外拓展b的一倍再加n个比特.
func (b *Builder) grow(n int) {
	buf := bytealg.MakeNoZero(2*cap(b.buf) + n)[:len(b.buf)]
	copy(buf, b.buf)
	b.buf = buf
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative, Grow panics.
func (b *Builder) Grow(n int) {
	b.copyCheck() //先让b 不进行逃逸分析提升效率.栈是用完就回收, 堆是用gc,所以尽量用栈会让程序更快.
	if n < 0 {
		panic("strings.Builder.Grow: negative count")
	}
	if cap(b.buf)-len(b.buf) < n { //cap足够大就不用拓展
		b.grow(n)
	}
}

// Write appends the contents of p to b's buffer.
// Write always returns len(p), nil.
func (b *Builder) Write(p []byte) (int, error) { //p写入b.buff
	b.copyCheck()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to b's buffer.
// The returned error is always nil.
func (b *Builder) WriteByte(c byte) error {
	b.copyCheck()
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
// It returns the length of r and a nil error.
func (b *Builder) WriteRune(r rune) (int, error) { // r进行utf8编码后得到的byte加入b里面
	b.copyCheck()
	n := len(b.buf)
	b.buf = utf8.AppendRune(b.buf, r)
	return len(b.buf) - n, nil
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) { //直接数组添加即可.
	b.copyCheck()
	b.buf = append(b.buf, s...)
	return len(s), nil
}
