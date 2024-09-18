// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi

// These functions are the build-time version of the Go type data structures.

// Their contents must be kept in sync with their definitions.
// Because the host and target type sizes can differ, the compiler and
// linker cannot use the host information that they might get from
// either unsafe.Sizeof and Alignof, nor runtime, reflect, or reflectlite.

// CommonSize returns sizeof(Type) for a compilation target with a given ptrSize.  ptrSize表示当前计算机里面一个指针是多少byte的. 一般64位是8byte
func CommonSize(ptrSize int) int { return 4*ptrSize + 8 + 8 } //对编译的目标计算大小. //参考src\internal\abi\type.go:20 这个结构体. 这个结构体有4个指针. 其他字段加起来是16个byte

// StructFieldSize returns sizeof(StructField) for a compilation target with a given ptrSize
func StructFieldSize(ptrSize int) int { return 3 * ptrSize } //因为结构体StructField是3个指针组成.

// UncommonSize returns sizeof(UncommonType).  This currently does not depend on ptrSize.
// This exported function is in an internal package, so it may change to depend on ptrSize in the future.
func UncommonSize() uint64 { return 4 + 2 + 2 + 4 + 4 } //参考src\internal\abi\type.go:204. 里面各个类型换算成byte大小就是这个4 + 2 + 2 + 4 + 4

// TFlagOff returns the offset of Type.TFlag for a compilation target with a given ptrSize
func TFlagOff(ptrSize int) int { return 2*ptrSize + 4 } // 这个是计算偏移量. 我们看结构体.src\internal\abi\type.go:20 他20到24行代码之间的部分是2个指针加一个u32的大小.所以是2*ptrsize+4
