// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"errors"
	"io"
	"unicode/utf8"
)

// A Reader implements the [io.Reader], [io.ReaderAt], [io.ByteReader], [io.ByteScanner],
// [io.RuneReader], [io.RuneScanner], [io.Seeker], and [io.WriterTo] interfaces by reading
// from a string.
// The zero value for Reader operates like a Reader of an empty string.
type Reader struct { //创造一个类,让他来做字符串的读取任务.
	s        string //用来存数据
	i        int64  // current reading index //当前开始读的话, 第一个读入的索引.
	prevRune int    // index of previous rune; or < 0 //上一次读的rune的索引. 在 UnreadRune 函数中有重要使用.
}

// Len returns the number of bytes of the unread portion of the
// string.
func (r *Reader) Len() int { //len函数返回字符串里面未读入的字符长度.
	if r.i >= int64(len(r.s)) { //如果读的位置已经到尾, 那么返回0
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

// Size returns the original length of the underlying string.
// Size is the number of bytes available for reading via [Reader.ReadAt].
// The returned value is always the same and is not affected by calls
// to any other method.
func (r *Reader) Size() int64 { return int64(len(r.s)) } //返回字符串全部长度

// Read implements the [io.Reader] interface.
func (r *Reader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(b, r.s[r.i:]) // 把字符串未读入的数据都放b里面, n是读了多少字符.然后索引i更新.
	r.i += int64(n)
	return
}

// ReadAt implements the [io.ReaderAt] interface.
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error) { //类似同上
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("strings.Reader.ReadAt: negative offset")
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

// ReadByte implements the [io.ByteReader] interface.
func (r *Reader) ReadByte() (byte, error) { //读一个字符
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	b := r.s[r.i]
	r.i++
	return b, nil
}

// UnreadByte implements the [io.ByteScanner] interface.
func (r *Reader) UnreadByte() error { //索引退一个
	if r.i <= 0 {
		return errors.New("strings.Reader.UnreadByte: at beginning of string")
	}
	r.prevRune = -1
	r.i--
	return nil
}

// ReadRune implements the [io.RuneReader] interface.
func (r *Reader) ReadRune() (ch rune, size int, err error) { //从索引位往后读出一个rune记作ch,返回他的占用的byte数记作size
	if r.i >= int64(len(r.s)) {
		r.prevRune = -1
		return 0, 0, io.EOF
	}
	r.prevRune = int(r.i)
	if c := r.s[r.i]; c < utf8.RuneSelf {
		r.i++
		return rune(c), 1, nil
	}
	ch, size = utf8.DecodeRuneInString(r.s[r.i:])
	r.i += int64(size)
	return
}

// UnreadRune implements the [io.RuneScanner] interface.
func (r *Reader) UnreadRune() error {
	if r.i <= 0 {
		return errors.New("strings.Reader.UnreadRune: at beginning of string")
	}
	if r.prevRune < 0 {
		return errors.New("strings.Reader.UnreadRune: previous operation was not ReadRune")
	}
	r.i = int64(r.prevRune)
	r.prevRune = -1
	return nil
}

// Seek implements the [io.Seeker] interface.
func (r *Reader) Seek(offset int64, whence int) (int64, error) { //这个函数用来调整r里面的光标位置.//whence是开始位置, offset是 从whence开始计算的偏移位置, 是一个整数. 返回值是绝对位置. 并且写入r的索引里面.
	r.prevRune = -1
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.i + offset
	case io.SeekEnd:
		abs = int64(len(r.s)) + offset
	default:
		return 0, errors.New("strings.Reader.Seek: invalid whence")
	} //先abs算出来索引位置最后的绝对位置.
	if abs < 0 {
		return 0, errors.New("strings.Reader.Seek: negative position")
	}
	r.i = abs       //绝对位置写入r的索引.
	return abs, nil //返回绝对位置.
}

// WriteTo implements the [io.WriterTo] interface.
func (r *Reader) WriteTo(w io.Writer) (n int64, err error) { //入参是一个 writer, 然后我们把r的信息写入w里面. 写了多少记作n
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, nil
	}
	s := r.s[r.i:]
	m, err := io.WriteString(w, s) //往w里面写s
	if m > len(s) {
		panic("strings.Reader.WriteTo: invalid WriteString count")
	}
	r.i += int64(m)
	n = int64(m)
	if m != len(s) && err == nil {
		err = io.ErrShortWrite
	}
	return
}

// Reset resets the [Reader] to be reading from s.
func (r *Reader) Reset(s string) { *r = Reader{s, 0, -1} } // 最后的rune=-1表示没读

// NewReader returns a new [Reader] reading from s.
// It is similar to [bytes.NewBufferString] but more efficient and non-writable.
func NewReader(s string) *Reader { return &Reader{s, 0, -1} } //新reader, 就是从index0开始, 上一个rune设置为-1表示没有.
