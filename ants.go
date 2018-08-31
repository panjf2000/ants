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
)

const (
	// DefaultAntsPoolSize is the default capacity for a default goroutine pool.
	DefaultAntsPoolSize = math.MaxInt32

	// DefaultCleanIntervalTime is the interval time to clean up goroutines.
	DefaultCleanIntervalTime = 5
)

// Init a instance pool when importing ants.
var defaultAntsPool, _ = NewPool(DefaultAntsPoolSize)

// Submit submits a task to pool.
func Submit(task f) error {
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
	defaultAntsPool.Release()
}

// Errors for the Ants API.
var (
	ErrInvalidPoolSize   = errors.New("invalid size for pool")
	ErrInvalidPoolExpiry = errors.New("invalid expiry for pool")
	ErrPoolClosed        = errors.New("this pool has been closed")
)
