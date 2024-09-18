// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"sync/atomic"
	"unsafe"
)

// Cond implements a condition variable, a rendezvous point
// for goroutines waiting for or announcing the occurrence
// of an event.
//
// Each Cond has an associated Locker L (often a *Mutex or *RWMutex),
// which must be held when changing the condition and
// when calling the Wait method.
//
// A Cond must not be copied after first use.
//
// In the terminology of the Go memory model, Cond arranges that
// a call to Broadcast or Signal “synchronizes before” any Wait call
// that it unblocks.
//
// For many simple use cases, users will be better off using channels than a
// Cond (Broadcast corresponds to closing a channel, and Signal corresponds to
// sending on a channel).
//
// For more on replacements for sync.Cond, see [Roberto Clapis's series on
// advanced concurrency patterns], as well as [Bryan Mills's talk on concurrency
// patterns].
//
// [Roberto Clapis's series on advanced concurrency patterns]: https://blogtitle.github.io/categories/concurrency/
// [Bryan Mills's talk on concurrency patterns]: https://drive.google.com/file/d/1nPdvhB0PutEJzdCq5ms6UI58dp50fcAN/view
type Cond struct { //cond是一个进程设置一个触发条件.// cond使用比较复杂,可以看https://cloud.tencent.com/developer/article/2296185 讲的比较深入.//主要方法就是 wait来让一个进程等, signal让一个进程通知其他等的进程可以运行一个, Broadcast 让进程通知其他等的进程都可以运行.所以我们核心就是看这3个函数如何实现. 主要靠的是runtime底层,这里不深入其他底层库包.写清楚cond.go里面的设计思路.
	noCopy noCopy

	// L is held while observing or changing the condition
	L Locker

	notify  notifyList  //signal函数时候, 会让notify里面进程随机启动一个.
	checker copyChecker //检查是否发生copy. 跟锁有关的概念都禁止copy.
}

// NewCond returns a new Cond with Locker l.
func NewCond(l Locker) *Cond {
	return &Cond{L: l}
}

// Wait atomically unlocks c.L and suspends execution
// of the calling goroutine. After later resuming execution,
// Wait locks c.L before returning. Unlike in other systems,
// Wait cannot return unless awoken by Broadcast or Signal.
//
// Because c.L is not locked while Wait is waiting, the caller
// typically cannot assume that the condition is true when
// Wait returns. Instead, the caller should Wait in a loop:
//
//	 //这是condition的使用方法.
//		c.L.Lock()
//		for !condition() {
//		    c.Wait()
//		}
//		... make use of condition ...
//		c.L.Unlock()
func (c *Cond) Wait() { //底层设计runtime里面的函数.这里先跳过细节.
	c.checker.check()                     //用来检查是否发生了copy
	t := runtime_notifyListAdd(&c.notify) //wait之前的代码我们获得锁,然后我们告诉运行时,把c能激活的进程都记录在册.等待c.signal时候让他们醒一个.
	c.L.Unlock()                          //这时就可以解锁了.因为下面的wait代码可以并发.不会资源竞争.
	runtime_notifyListWait(&c.notify, t)  //启动等待.
	c.L.Lock()
}

// Signal wakes one goroutine waiting on c, if there is any.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
//
// Signal() does not affect goroutine scheduling priority; if other goroutines
// are attempting to lock c.L, they may be awoken before a "waiting" goroutine.
func (c *Cond) Signal() {
	c.checker.check()
	runtime_notifyListNotifyOne(&c.notify) //之前wait函数注册的那些,可以激活了.
}

// Broadcast wakes all goroutines waiting on c.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
func (c *Cond) Broadcast() {
	c.checker.check()
	runtime_notifyListNotifyAll(&c.notify)
}

// copyChecker holds back pointer to itself to detect object copying.
type copyChecker uintptr

func (c *copyChecker) check() { // 用来防止复制的.
	// Check if c has been copied in three steps:
	// 1. The first comparison is the fast-path. If c has been initialized and not copied, this will return immediately. Otherwise, c is either not initialized, or has been copied.
	// 2. Ensure c is initialized. If the CAS succeeds, we're done. If it fails, c was either initialized concurrently and we simply lost the race, or c has been copied.
	// 3. Do step 1 again. Now that c is definitely initialized, if this fails, c was copied.            //104行, 第一次check函数调用进来c是*0,一个指向0的指针. uintptr(unsafe.Pointer(c)) [这里面的语法不熟悉的请参考unsafe.go:50行的注释,也就是unitptr的第二种使用方式]是这个指针的地址,所以显然不等于, 所以继续走 后续的atomic运算.这个运算当c=*0时候,比如c本身表示数字是0xfff, 那么atomic走完c=*0xfff//第二次check函数调用c=*0xffff,因为c已经指向自己本身地址的数字了.所以uintptr(*c) != uintptr(unsafe.Pointer(c)) 是false.所以就保证了正常不copy模式不会触发panic. //  如果c复制了.那么一个进程会改c的值,一个进程c还是0x0, 那么就会触发panic.//这个代码是一个很好的nocopy的实现.
	if uintptr(*c) != uintptr(unsafe.Pointer(c)) &&
		!atomic.CompareAndSwapUintptr((*uintptr)(c), 0, uintptr(unsafe.Pointer(c))) &&
		uintptr(*c) != uintptr(unsafe.Pointer(c)) {
		panic("sync.Cond is copied")
	}
}

// noCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
