// MIT License

// Copyright (c) 2018 Andy Pan

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ants

import (
	"errors"
	"math"
	"runtime"
)

const (
	// DEFAULT_ANTS_POOL_SIZE is the default capacity for a default goroutine pool.
	DEFAULT_ANTS_POOL_SIZE = math.MaxInt32

	// DEFAULT_CLEAN_INTERVAL_TIME is the interval time to clean up goroutines.
	DEFAULT_CLEAN_INTERVAL_TIME = 1

	// CLOSED represents that the pool is closed.
	CLOSED = 1
)

var (
	// Error types for the Ants API.
	//---------------------------------------------------------------------------
	// ErrInvalidPoolSize will be returned when setting a negative number as pool capacity.
	ErrInvalidPoolSize = errors.New("invalid size for pool")

	// ErrInvalidPoolExpiry will be returned when setting a negative number as the periodic duration to purge goroutines.
	ErrInvalidPoolExpiry = errors.New("invalid expiry for pool")

	// ErrPoolClosed will be returned when submitting task to a closed pool.
	ErrPoolClosed = errors.New("this pool has been closed")
	//---------------------------------------------------------------------------

	// workerChanCap determines whether the channel of a worker should be a buffered channel
	// to get the best performance. Inspired by fasthttp at https://github.com/valyala/fasthttp/blob/master/workerpool.go#L139
	workerChanCap = func() int {
		// Use blocking workerChan if GOMAXPROCS=1.
		// This immediately switches Serve to WorkerFunc, which results
		// in higher performance (under go1.5 at least).
		if runtime.GOMAXPROCS(0) == 1 {
			return 0
		}

		// Use non-blocking workerChan if GOMAXPROCS>1,
		// since otherwise the Serve caller (Acceptor) may lag accepting
		// new connections if WorkerFunc is CPU-bound.
		return 1
	}()

    defaultAntsPool, _ = NewPool(DEFAULT_ANTS_POOL_SIZE)
)

// Init a instance pool when importing ants.

// Submit submits a task to pool.
func Submit(task func()) error {
	return defaultAntsPool.Submit(task)
}

// Running returns the number of the currently running goroutines.
func Running() int {
	return defaultAntsPool.Running()
}

// Cap returns the capacity of this default pool.
func Cap() int {
	return defaultAntsPool.Cap()
}

// Free returns the available goroutines to work.
func Free() int {
	return defaultAntsPool.Free()
}

// Release Closes the default pool.
func Release() {
	_ = defaultAntsPool.Release()
}
