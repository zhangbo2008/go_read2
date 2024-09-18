// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package poll

import (
	"internal/itoa"
	"runtime"
	"sync"
	"syscall"
)

// asyncIO implements asynchronous cancelable I/O.
// An asyncIO represents a single asynchronous Read or Write
// operation. The result is returned on the result channel.
// The undergoing I/O system call can either complete or be
// interrupted by a note.
type asyncIO struct {
	res chan result

	// mu guards the pid field.
	mu sync.Mutex

	// pid holds the process id of
	// the process running the IO operation.
	pid int
}

// result is the return value of a Read or Write operation.// n是记录读写的字符数量.
type result struct {
	n   int
	err error
}

// newAsyncIO returns a new asyncIO that performs an I/O
// operation by calling fn, which must do one and only one
// interruptible system call. // fn是io操作,b是buffer// newAsyncIO 会新建一个异步io,然后读取通过fn.
func newAsyncIO(fn func([]byte) (int, error), b []byte) *asyncIO {
	aio := &asyncIO{
		res: make(chan result, 0),
	} //res是一个无buffer的channel类型.
	aio.mu.Lock()
	go func() {
		// Lock the current goroutine to its process
		// and store the pid in io so that Cancel can
		// interrupt it. We ignore the "hangup" signal,
		// so the signal does not take down the entire
		// Go runtime.
		runtime.LockOSThread()
		runtime_ignoreHangup()
		aio.pid = syscall.Getpid()
		aio.mu.Unlock() //这个地方释放锁,可以让69行函数来禁止这个函数读取.

		n, err := fn(b)

		aio.mu.Lock() //已经读取完了, 那么就没法停止了,所以这里加锁.
		aio.pid = -1
		runtime_unignoreHangup()
		aio.mu.Unlock()

		aio.res <- result{n, err}
	}()
	return aio
}

// Cancel interrupts the I/O operation, causing
// the Wait function to return.
func (aio *asyncIO) Cancel() { //这个函数可以异步cancel上一个函数的读取任务.
	aio.mu.Lock()
	defer aio.mu.Unlock()
	if aio.pid == -1 {
		return
	} //75行有可能在55行代码运行是触发来cancel55行的代码.
	f, e := syscall.Open("/proc/"+itoa.Itoa(aio.pid)+"/note", syscall.O_WRONLY)
	if e != nil {
		return
	}
	syscall.Write(f, []byte("hangup"))
	syscall.Close(f) //调用系统命令来关闭f这个文件描述符.
}

// Wait for the I/O operation to complete.
func (aio *asyncIO) Wait() (int, error) { //让后续等待io完成. 利用channel即可.
	res := <-aio.res
	return res.n, res.err
}

// The following functions, provided by the runtime, are used to
// ignore and unignore the "hangup" signal received by the process.
func runtime_ignoreHangup()
func runtime_unignoreHangup()
