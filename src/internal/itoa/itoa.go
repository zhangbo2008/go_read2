// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Simple conversions to avoid depending on strconv.

package itoa

// Itoa converts val to a decimal string. //把整数转化为字符串.
func Itoa(val int) string {
	if val < 0 {
		return "-" + Uitoa(uint(-val))
	}
	return Uitoa(uint(val))
}

// Uitoa converts val to a decimal string.
func Uitoa(val uint) string { //把uint转为string
	if val == 0 { // avoid string allocation
		return "0"
	}
	var buf [20]byte // big enough for 64bit value base 10 2的64次幂足够十进制20位来保存.
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + val - q*10) //等号右边得到的是当前数位上的数字的ascii码大小.
		i--
		val = q
	}
	//上面for循环之后val写入了buf中.
	// val < 10
	buf[i] = byte('0' + val) //个位数
	return string(buf[i:])   //返回buf即可.
}

const hex = "0123456789abcdef"

// Uitox converts val (a uint) to a hexadecimal string.
func Uitox(val uint) string { // uint转16进制的字符串
	if val == 0 { // avoid string allocation
		return "0x0"
	}
	var buf [20]byte // big enough for 64bit value base 16 + 0x
	i := len(buf) - 1
	for val >= 16 {
		q := val / 16
		buf[i] = hex[val%16]
		i--
		val = q
	}
	// val < 16
	buf[i] = hex[val%16]
	i--
	buf[i] = 'x'
	i--
	buf[i] = '0'
	return string(buf[i:])
}
