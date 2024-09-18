// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"sync/atomic"
)

// Once is an object that will perform exactly one action.
//
// A Once must not be copied after first use.
//
// In the terminology of the Go memory model,
// the return from f “synchronizes before”
// the return from any call of once.Do(f).
type Once struct {
	// done indicates whether the action has been performed.
	// It is first in the struct because it is used in the hot path.
	// The hot path is inlined at every call site.
	// Placing done first allows more compact instructions on some architectures (amd64/386),
	// and fewer instructions (to calculate offset) on other architectures.
	done atomic.Uint32 //一个锁加一个flag记录即可.
	m    Mutex
}

// Do calls the function f if and only if Do is being called for the
// first time for this instance of Once. In other words, given
//
//	var once Once
//
// if once.Do(f) is called multiple times, only the first call will invoke f,
// even if f has a different value in each invocation. A new instance of
// Once is required for each function to execute.
//
// Do is intended for initialization that must be run exactly once. Since f
// is niladic, it may be necessary to use a function literal to capture the
// arguments to a function to be invoked by Do:
//
//	config.once.Do(func() { config.init(filename) })
//
// Because no call to Do returns until the one call to f returns, if f causes
// Do to be called, it will deadlock.
//
// If f panics, Do considers it to have returned; future calls of Do return
// without calling f.
func (o *Once) Do(f func()) {
	// Note: Here is an incorrect implementation of Do:
	//
	//	if o.done.CompareAndSwap(0, 1) {
	//		f()
	//	}
	//
	// Do guarantees that when it returns, f has finished.
	// This implementation would not implement that guarantee:
	// given two simultaneous calls, the winner of the cas would
	// call f, and the second would return immediately, without
	// waiting for the first's call to f to complete.
	// This is why the slow path falls back to a mutex, and why
	// the o.done.Store must be delayed until after f returns.

	if o.done.Load() == 0 { // 如果我们flag=0表示还没启用过, 我们才进行f函数调用. 否则啥也不干. 这就保证了单次.
		// Outlined slow-path to allow inlining of the fast-path.
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) { // do f函数,就是加锁即可.
	o.m.Lock()
	defer o.m.Unlock()
	if o.done.Load() == 0 { //注意因为并发关系,我们这个进程跑到这里这段时间消耗了,可能其他进程也修改了o.done变量.所以我们还需要读一次这个o.done变量.再次确认是单例模式.
		defer o.done.Store(1) //这里确认好了我们就在函数f结束之后运行这个o.done变成1即可.
		f()
	}
}
