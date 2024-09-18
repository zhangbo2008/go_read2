// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package poll

import "sync/atomic"

// fdMutex is a specialized synchronization primitive that manages
// lifetime of an fd and serializes access to Read, Write and Close
// methods on FD.
type fdMutex struct {
	state uint64
	rsema uint32
	wsema uint32
}

// fdMutex.state is organized as follows:
// 1 bit - whether FD is closed, if set all subsequent lock operations will fail.
// 1 bit - lock for read operations.
// 1 bit - lock for write operations.
// 20 bits - total number of references (read+write+misc).
// 20 bits - number of outstanding read waiters.
// 20 bits - number of outstanding write waiters.
const (
	mutexClosed  = 1 << 0
	mutexRLock   = 1 << 1
	mutexWLock   = 1 << 2
	mutexRef     = 1 << 3
	mutexRefMask = (1<<20 - 1) << 3
	mutexRWait   = 1 << 23
	mutexRMask   = (1<<20 - 1) << 23
	mutexWWait   = 1 << 43
	mutexWMask   = (1<<20 - 1) << 43
)

const overflowMsg = "too many concurrent operations on a single file or socket (max 1048575)" // 下面给出这个最大值的计算. mutexRMask(的表示的数量,他由23个1组成.)是记录read waiter的数量, 而我们每次添加的大小是mutexRef 也就是8. 所以我们数量是 2**23/8-1=1048575

// Read operations must do rwlock(true)/rwunlock(true).
//
// Write operations must do rwlock(false)/rwunlock(false).
//
// Misc operations must do incref/decref.
// Misc operations include functions like setsockopt and setDeadline.
// They need to use incref/decref to ensure that they operate on the
// correct fd in presence of a concurrent close call (otherwise fd can
// be closed under their feet).
//
// Close operations must do increfAndClose/decref.

// incref adds a reference to mu.
// It reports whether mu is available for reading or writing.
func (mu *fdMutex) incref() bool { //给mu加引用.
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexClosed != 0 { //查询fd是否关闭,已经关闭了那么就直接return false
			return false
		}
		new := old + mutexRef
		if new&mutexRefMask == 0 {
			panic(overflowMsg)
		}
		if atomic.CompareAndSwapUint64(&mu.state, old, new) { //cas 保证了原子操作.
			return true
		}
	}
}

// increfAndClose sets the state of mu to closed.
// It returns false if the file was already closed.
func (mu *fdMutex) increfAndClose() bool {
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexClosed != 0 {
			return false
		}
		// Mark as closed and acquire a reference.
		new := (old | mutexClosed) + mutexRef //标记为关闭,因为我们函数名字就叫添加ref,并且关闭fd.
		if new&mutexRefMask == 0 {
			panic(overflowMsg)
		}
		// Remove all read and write waiters.
		new &^= mutexRMask | mutexWMask
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			// Wake all read and write waiters,
			// they will observe closed flag after wakeup.
			for old&mutexRMask != 0 {
				old -= mutexRWait
				runtime_Semrelease(&mu.rsema) //释放读的信号量
			}
			for old&mutexWMask != 0 {
				old -= mutexWWait
				runtime_Semrelease(&mu.wsema) //释放写信号量.
			}
			return true
		}
	}
}

// decref removes a reference from mu.
// It reports whether there is no remaining reference.
func (mu *fdMutex) decref() bool {
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexRefMask == 0 {
			panic("inconsistent poll.fdMutex")
		}
		new := old - mutexRef
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			return new&(mutexClosed|mutexRefMask) == mutexClosed
		}
	}
}

// lock adds a reference to mu and locks mu.
// It reports whether mu is available for reading or writing.// mu添加索引, 然后锁上mu
func (mu *fdMutex) rwlock(read bool) bool {
	var mutexBit, mutexWait, mutexMask uint64
	var mutexSema *uint32
	if read {
		mutexBit = mutexRLock
		mutexWait = mutexRWait
		mutexMask = mutexRMask
		mutexSema = &mu.rsema
	} else {
		mutexBit = mutexWLock
		mutexWait = mutexWWait
		mutexMask = mutexWMask
		mutexSema = &mu.wsema
	}
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexClosed != 0 {
			return false
		}
		var new uint64
		if old&mutexBit == 0 { //还没锁上,那么就139行锁上.
			// Lock is free, acquire it.
			new = (old | mutexBit) + mutexRef
			if new&mutexRefMask == 0 {
				panic(overflowMsg)
			}
		} else {
			// Wait for lock.//已经锁上了, 状态加wait
			new = old + mutexWait
			if new&mutexMask == 0 {
				panic(overflowMsg)
			}
		}
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			if old&mutexBit == 0 { //旧的能获得锁, 说明我们成功了.
				return true
			}
			runtime_Semacquire(mutexSema) //没成功就继续获得信号.
			// The signaller has subtracted mutexWait.
		}
	}
}

// unlock removes a reference from mu and unlocks mu.
// It reports whether there is no remaining reference.
func (mu *fdMutex) rwunlock(read bool) bool {
	var mutexBit, mutexWait, mutexMask uint64
	var mutexSema *uint32
	if read {
		mutexBit = mutexRLock
		mutexWait = mutexRWait
		mutexMask = mutexRMask
		mutexSema = &mu.rsema
	} else {
		mutexBit = mutexWLock
		mutexWait = mutexWWait
		mutexMask = mutexWMask
		mutexSema = &mu.wsema
	}
	for {
		old := atomic.LoadUint64(&mu.state)
		if old&mutexBit == 0 || old&mutexRefMask == 0 {
			panic("inconsistent poll.fdMutex")
		}
		// Drop lock, drop reference and wake read waiter if present.
		new := (old &^ mutexBit) - mutexRef
		if old&mutexMask != 0 {
			new -= mutexWait
		}
		if atomic.CompareAndSwapUint64(&mu.state, old, new) {
			if old&mutexMask != 0 {
				runtime_Semrelease(mutexSema)
			}
			return new&(mutexClosed|mutexRefMask) == mutexClosed
		}
	}
}

// Implemented in runtime package.
func runtime_Semacquire(sema *uint32) //src\sync\runtime.go:14 至于更底层的汇编实现这个信号量锁的操作我们留到后面.
func runtime_Semrelease(sema *uint32)

// incref adds a reference to fd.
// It returns an error when fd cannot be used.
func (fd *FD) incref() error {
	if !fd.fdmu.incref() {
		return errClosing(fd.isFile)
	}
	return nil
}

// decref removes a reference from fd.
// It also closes fd when the state of fd is set to closed and there
// is no remaining reference.
func (fd *FD) decref() error {
	if fd.fdmu.decref() {
		return fd.destroy()
	}
	return nil
}

// readLock adds a reference to fd and locks fd for reading.
// It returns an error when fd cannot be used for reading.
func (fd *FD) readLock() error {
	if !fd.fdmu.rwlock(true) {
		return errClosing(fd.isFile)
	}
	return nil
}

// readUnlock removes a reference from fd and unlocks fd for reading.
// It also closes fd when the state of fd is set to closed and there
// is no remaining reference.
func (fd *FD) readUnlock() {
	if fd.fdmu.rwunlock(true) {
		fd.destroy()
	}
}

// writeLock adds a reference to fd and locks fd for writing.
// It returns an error when fd cannot be used for writing.
func (fd *FD) writeLock() error {
	if !fd.fdmu.rwlock(false) {
		return errClosing(fd.isFile)
	}
	return nil
}

// writeUnlock removes a reference from fd and unlocks fd for writing.
// It also closes fd when the state of fd is set to closed and there
// is no remaining reference.
func (fd *FD) writeUnlock() {
	if fd.fdmu.rwunlock(false) {
		fd.destroy()
	}
}
