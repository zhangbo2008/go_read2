// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sync provides basic synchronization primitives such as mutual
// exclusion locks. Other than the Once and WaitGroup types, most are intended
// for use by low-level library routines. Higher-level synchronization is
// better done via channels and communication.
//
// Values containing the types defined in this package should not be copied.
package sync

import (
	"internal/race"
	"sync/atomic"
	"unsafe"
)

// Provided by runtime via linkname.
func throw(string)
func fatal(string)

// A Mutex is a mutual exclusion lock.
// The zero value for a Mutex is an unlocked mutex.
//
// A Mutex must not be copied after first use.
//
// In the terminology of the Go memory model,
// the n'th call to Unlock “synchronizes before” the m'th call to Lock
// for any n < m.
// A successful call to TryLock is equivalent to a call to Lock.
// A failed call to TryLock does not establish any “synchronizes before”
// relation at all.
type Mutex struct {
	state int32 // 0表示没有锁上, 其他值表示锁上了.
	sema  uint32
}

// A Locker represents an object that can be locked and unlocked.
type Locker interface {
	Lock()
	Unlock()
}

const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken              //用来表示这个锁刚才被从woken中唤醒了.
	mutexStarving
	mutexWaiterShift = iota //这个字段用来检查是否有人在等待这个锁

	// Mutex fairness.
	//
	// Mutex can be in 2 modes of operations: normal and starvation.
	// In normal mode waiters are queued in FIFO order, but a woken up waiter
	// does not own the mutex and competes with new arriving goroutines over
	// the ownership. New arriving goroutines have an advantage -- they are
	// already running on CPU and there can be lots of them, so a woken up
	// waiter has good chances of losing. In such case it is queued at front
	// of the wait queue. If a waiter fails to acquire the mutex for more than 1ms,
	// it switches mutex to the starvation mode.// 一般是normal, 在队列里面排序, 先进先出. 但是这时候新来的进程有更先优先级. 如果等的等不到超过1ms, 就进入饥饿模式.

	//
	// In starvation mode ownership of the mutex is directly handed off from
	// the unlocking goroutine to the waiter at the front of the queue.
	// New arriving goroutines don't try to acquire the mutex even if it appears
	// to be unlocked, and don't try to spin. Instead they queue themselves at
	// the tail of the wait queue. //饥饿模式下拿到锁的进程直接跑到队列头, 新来的进程直接进到队尾. 这个意思就是拿到锁的用于不放开了.他一直用到他自己主动放弃才行.
	//
	// If a waiter receives ownership of the mutex and sees that either
	// (1) it is the last waiter in the queue, or (2) it waited for less than 1 ms,
	// it switches mutex back to normal operation mode. //这两种模式会让锁变成普通模式.
	//
	// Normal mode has considerably better performance as a goroutine can acquire
	// a mutex several times in a row even if there are blocked waiters.
	// Starvation mode is important to prevent pathological cases of tail latency.
	starvationThresholdNs = 1e6
)

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available. //如果已经锁上了,那么这个进程就一直等到互斥锁放开.
func (m *Mutex) Lock() {
	// Fast path: grab unlocked mutex.
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) { //利用元操作, 如果state=0表示没有锁上, 那么就让state=1进行锁上然后函数返回. 如果已经锁上了,那么就走91行.
		if race.Enabled { //这个地方会检测是否代码开始起race检测. race是进程竞争的意思. 比如一个进程改a, 另一个进程也同时可以改a, 那么就会race检测报warning.
			race.Acquire(unsafe.Pointer(m)) // 如果race检测,那么我们就用race来获取这个m变量.//至于race底层代码还是runtime里面, 留在以后研究底层实现.
		}
		return
	}
	// Slow path (outlined so that the fast path can be inlined)
	m.lockSlow() //这里面一直等, 直到拿到锁, 然后锁上.
}

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes. //这里如果拿不到锁, 直接放弃.这是跟上面函数的区别.
func (m *Mutex) TryLock() bool {
	old := m.state
	if old&(mutexLocked|mutexStarving) != 0 { // state 如果是mutexLocked, 那么直接返回失败,  state 如果是mutexStarving, 那么直接返回失败,因为starvign模式下,占有锁的进程不可能被你抢到, 只能等他主动放弃锁的权利才行.
		return false
	}

	// There may be a goroutine waiting for the mutex, but we are
	// running now and can try to grab the mutex before that
	// goroutine wakes up.
	if !atomic.CompareAndSwapInt32(&m.state, old, old|mutexLocked) { //代码运行到这行, 说明锁现在可以尝试去拿到. 那么我们这里用cas操作来变锁的状态.如果我们这行还能读取到state=old, 那么就尝试把状态变为locked. 如果成功了, 那么就115行返回. 如果失败了就109行返回false.
		return false
	}

	if race.Enabled { //这地方还是竞争检测一下, 具体底层race留给以后研究.
		race.Acquire(unsafe.Pointer(m))
	}
	return true
}

func (m *Mutex) lockSlow() { //整个思路比较复杂, 核心就是我们最上面注释里面写的, 普通模式和饥饿模式的转化, 里面的判定.
	var waitStartTime int64
	starving := false
	awoke := false
	iter := 0
	old := m.state
	for {
		// Don't spin in starvation mode, ownership is handed off to waiters
		// so we won't be able to acquire the mutex anyway.
		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			// Active spinning makes sense.
			// Try to set mutexWoken flag to inform Unlock
			// to not wake other blocked goroutines.
			if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
				atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
				awoke = true
			}
			runtime_doSpin()
			iter++
			old = m.state
			continue
		}
		new := old
		// Don't try to acquire starving mutex, new arriving goroutines must queue.
		if old&mutexStarving == 0 {
			new |= mutexLocked
		}
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}
		// The current goroutine switches mutex to starvation mode.
		// But if the mutex is currently unlocked, don't do the switch.
		// Unlock expects that starving mutex has waiters, which will not
		// be true in this case.
		if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}
		if awoke {
			// The goroutine has been woken from sleep,
			// so we need to reset the flag in either case.
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			new &^= mutexWoken
		}
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			if old&(mutexLocked|mutexStarving) == 0 {
				break // locked the mutex with CAS
			}
			// If we were already waiting before, queue at the front of the queue.
			queueLifo := waitStartTime != 0
			if waitStartTime == 0 {
				waitStartTime = runtime_nanotime()
			}
			runtime_SemacquireMutex(&m.sema, queueLifo, 1)
			starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs //starvationThresholdNs表示1ms
			old = m.state
			if old&mutexStarving != 0 {
				// If this goroutine was woken and mutex is in starvation mode,
				// ownership was handed off to us but mutex is in somewhat
				// inconsistent state: mutexLocked is not set and we are still
				// accounted as waiter. Fix that.
				if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
					throw("sync: inconsistent mutex state")
				}
				delta := int32(mutexLocked - 1<<mutexWaiterShift)
				if !starving || old>>mutexWaiterShift == 1 {
					// Exit starvation mode.
					// Critical to do it here and consider wait time.
					// Starvation mode is so inefficient, that two goroutines
					// can go lock-step infinitely once they switch mutex
					// to starvation mode.
					delta -= mutexStarving
				}
				atomic.AddInt32(&m.state, delta)
				break
			}
			awoke = true
			iter = 0
		} else {
			old = m.state
		}
	}

	if race.Enabled {
		race.Acquire(unsafe.Pointer(m))
	}
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
//
// A locked Mutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *Mutex) Unlock() {
	if race.Enabled {
		_ = m.state
		race.Release(unsafe.Pointer(m))
	}

	// Fast path: drop lock bit.
	new := atomic.AddInt32(&m.state, -mutexLocked) //状态1 减去1,如果结果是0, 那么就是解锁了.
	if new != 0 {                                  //如果解锁失败, 那么就进入慢解锁.
		// Outlined slow path to allow inlining the fast path.
		// To hide unlockSlow during tracing we skip one extra frame when tracing GoUnblock.
		m.unlockSlow(new)
	}
}

func (m *Mutex) unlockSlow(new int32) { //new是当前的状态.
	if (new+mutexLocked)&mutexLocked == 0 {
		fatal("sync: unlock of unlocked mutex")
	}
	if new&mutexStarving == 0 { //非饥饿模式.
		old := new
		for {
			// If there are no waiters or a goroutine has already
			// been woken or grabbed the lock, no need to wake anyone.
			// In starvation mode ownership is directly handed off from unlocking
			// goroutine to the next waiter. We are not part of this chain,
			// since we did not observe mutexStarving when we unlocked the mutex above.
			// So get off the way.
			if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
				return //It enters a loop where it checks if there are no waiters (goroutines waiting for the mutex) or if the mutex has already been grabbed by another goroutine since it was unlocked.
				// If either condition is true, the function returns, as there’s nothing to do.
			}
			// Grab the right to wake someone.
			new = (old - 1<<mutexWaiterShift) | mutexWoken //If there are waiters, it tries to claim the right to wake one up by setting the mutexWoken bit in the mutex state.
			if atomic.CompareAndSwapInt32(&m.state, old, new) {
				runtime_Semrelease(&m.sema, false, 1)
				return
			}
			old = m.state
		}
	} else {
		// Starving mode: handoff mutex ownership to the next waiter, and yield
		// our time slice so that the next waiter can start to run immediately.
		// Note: mutexLocked is not set, the waiter will set it after wakeup.
		// But mutex is still considered locked if mutexStarving is set,
		// so new coming goroutines won't acquire it.
		runtime_Semrelease(&m.sema, true, 1)
	}
}
