// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !goexperiment.staticlockranking

package sync

import "unsafe"

// Approximation of notifyList in runtime/sema.go. Size and alignment must
// agree. //一个数组列表. 记录的头尾,所以里面多少个也就记录了.
type notifyList struct {
	wait   uint32
	notify uint32
	lock   uintptr // key field of the mutex
	head   unsafe.Pointer
	tail   unsafe.Pointer
}
