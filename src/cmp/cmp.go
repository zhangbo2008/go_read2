// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cmp provides types and functions related to comparing
// ordered values.
package cmp

// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
//
// Note that floating-point types may contain NaN ("not-a-number") values.
// An operator such as == or < will always report false when
// comparing a NaN value with any other value, NaN or not.
// See the [Compare] function for a consistent way to compare NaN values.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
} // ~类型表示底层是这个类型就行的.范围更大. //这里面表示这些类型都可以.用来表示可以排序的类. //下面就用泛型.

// Less reports whether x is less than y.
// For floating-point types, a NaN is considered less than any non-NaN,
// and -0.0 is not less than (is equal to) 0.0.//  Nan比非Noan要小. NaN是not a number的缩写.
func Less[T Ordered](x, y T) bool {
	return (isNaN(x) && !isNaN(y)) || x < y
}

// Compare returns
//
//	-1 if x is less than y,
//	 0 if x equals y,
//	+1 if x is greater than y.
//
// For floating-point types, a NaN is considered less than any non-NaN,
// a NaN is considered equal to a NaN, and -0.0 is equal to 0.0.
func Compare[T Ordered](x, y T) int {
	xNaN := isNaN(x)
	yNaN := isNaN(y)
	if xNaN && yNaN {
		return 0
	}
	if xNaN || x < y {
		return -1
	}
	if yNaN || x > y {
		return +1
	}
	return 0
}

// isNaN reports whether x is a NaN without requiring the math package.
// This will always return false if T is not floating-point.
func isNaN[T Ordered](x T) bool { ////这是浮点数硬件决定的.  只有nan 符合 自己不等于自己这个性质.
	return x != x
}

// Or returns the first of its arguments that is not equal to the zero value.
// If no argument is non-zero, it returns the zero value.
func Or[T comparable](vals ...T) T { // 返回T每个元素做or运算的结果.
	var zero T
	for _, val := range vals {
		if val != zero {
			return val
		}
	}
	return zero
}
