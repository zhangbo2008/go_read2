// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strconv

const fnParseComplex = "ParseComplex"

/*

这里涉及go的类型转化.
1. 数值类型转换 :var a int = 10
               var b float64 = float64(a)
2. 字符串类型转换:
							var str string = "10"
							var num int
							num, _ = strconv.Atoi(str)
3. 接口类型转换:
							package main

							import "fmt"

							func main() {
									var i interface{} = "Hello, World"
									str, ok := i.(string)      //接口时候是数据.(类型)
									if ok {
											fmt.Printf("'%s' is a string\n", str)
									} else {
											fmt.Println("conversion failed")
									}
							}

*/

// convErr splits an error returned by parseFloatPrefix
// into a syntax or range error for ParseComplex.
func convErr(err error, s string) (syntax, range_ error) {
	if x, ok := err.(*NumError); ok { // NumError是err的特殊实现.所以可以尝试转化. 父类转子类: 这里 接口变量.(接口实现类的类型) 的操作可以理解为将父类类型强制转换为子类类型后返回，但是转换的前提是，实现类必须实现了当前接口的所有方法才行，否则go编译会报错
		x.Func = fnParseComplex //写入错误的函数
		x.Num = cloneString(s)  //写入错误函数的input值.
		if x.Err == ErrRange {
			return nil, x
		}
	}
	return err, nil
}

// ParseComplex converts the string s to a complex number
// with the precision specified by bitSize: 64 for complex64, or 128 for complex128.
// When bitSize=64, the result still has type complex128, but it will be
// convertible to complex64 without changing its value.
//
// The number represented by s must be of the form N, Ni, or N±Ni, where N stands
// for a floating-point number as recognized by ParseFloat, and i is the imaginary
// component. If the second N is unsigned, a + sign is required between the two components
// as indicated by the ±. If the second N is NaN, only a + sign is accepted.
// The form may be parenthesized and cannot contain any spaces.
// The resulting complex number consists of the two components converted by ParseFloat.
//
// The errors that ParseComplex returns have concrete type *NumError
// and include err.Num = s.
//
// If s is not syntactically well-formed, ParseComplex returns err.Err = ErrSyntax.
//
// If s is syntactically well-formed but either component is more than 1/2 ULP
// away from the largest floating point number of the given component's size,
// ParseComplex returns err.Err = ErrRange and c = ±Inf for the respective component.
func ParseComplex(s string, bitSize int) (complex128, error) { // 给字符s, s的bit长度. 返回转化后的复数.
	size := 64
	if bitSize == 64 {
		size = 32 // complex64 uses float32 parts
	}

	orig := s

	// Remove parentheses, if any.
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		s = s[1 : len(s)-1]
	}

	var pending error // pending range error, or nil

	// Read real part (possibly imaginary part if followed by 'i').
	re, n, err := parseFloatPrefix(s, size)
	if err != nil {
		err, pending = convErr(err, orig)
		if err != nil {
			return 0, err
		}
	}
	s = s[n:]

	// If we have nothing left, we're done.
	if len(s) == 0 {
		return complex(re, 0), pending
	}

	// Otherwise, look at the next character.
	switch s[0] {
	case '+':
		// Consume the '+' to avoid an error if we have "+NaNi", but
		// do this only if we don't have a "++" (don't hide that error).
		if len(s) > 1 && s[1] != '+' {
			s = s[1:]
		}
	case '-':
		// ok
	case 'i':
		// If 'i' is the last character, we only have an imaginary part.
		if len(s) == 1 {
			return complex(0, re), pending
		}
		fallthrough
	default:
		return 0, syntaxError(fnParseComplex, orig)
	}

	// Read imaginary part.
	im, n, err := parseFloatPrefix(s, size)
	if err != nil {
		err, pending = convErr(err, orig)
		if err != nil {
			return 0, err
		}
	}
	s = s[n:]
	if s != "i" {
		return 0, syntaxError(fnParseComplex, orig)
	}
	return complex(re, im), pending
}
