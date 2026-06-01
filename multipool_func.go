// MIT License

// Copyright (c) 2023 Andy Pan

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
	"context"
	"sync/atomic"
	"time"
)

// MultiPoolWithFunc consists of multiple pools, from which you will benefit the
// performance improvement on basis of the fine-grained locking that reduces
// the lock contention.
// MultiPoolWithFunc is a good fit for the scenario where you have a large number of
// tasks to submit, and you don't want the single pool to be the bottleneck.
type MultiPoolWithFunc struct {
	pools   []*PoolWithFunc
	metrics []PoolMetrics
	lb      LoadBalancer
	state   int32
}

// NewMultiPoolWithFunc instantiates a MultiPoolWithFunc with a size of the pool list and a size
// per pool, and the load-balancing strategy.
func NewMultiPoolWithFunc(size, sizePerPool int, fn func(any), lbs LoadBalancingStrategy, options ...Option) (*MultiPoolWithFunc, error) {
	if size <= 0 {
		return nil, ErrInvalidMultiPoolSize
	}
	lb, err := newBuiltinLB(lbs)
	if err != nil {
		return nil, err
	}
	return NewMultiPoolWithFuncAndLB(size, sizePerPool, fn, lb, options...)
}

// NewMultiPoolWithFuncAndLB instantiates a MultiPoolWithFunc with a given LoadBalancer.
func NewMultiPoolWithFuncAndLB(size, sizePerPool int, fn func(any), lb LoadBalancer, options ...Option) (*MultiPoolWithFunc, error) {
	if size <= 0 {
		return nil, ErrInvalidMultiPoolSize
	}
	if lb == nil {
		return nil, ErrInvalidLoadBalancingStrategy
	}

	pools := make([]*PoolWithFunc, size)
	metrics := make([]PoolMetrics, size)
	for i := 0; i < size; i++ {
		pool, err := NewPoolWithFunc(sizePerPool, fn, options...)
		if err != nil {
			// Release all previously created pools to avoid resource leak
			for j := 0; j < i; j++ {
				pools[j].Release()
			}
			return nil, err
		}
		pools[i] = pool
		metrics[i] = pool
	}

	return &MultiPoolWithFunc{pools: pools, metrics: metrics, lb: lb}, nil
}

// Invoke submits a task to a pool selected by the load-balancing strategy.
func (mp *MultiPoolWithFunc) Invoke(args any) (err error) {
	if mp.IsClosed() {
		return ErrPoolClosed
	}
	idx := mp.lb.Pick(mp.metrics)
	if !validIdx(idx, len(mp.pools)) {
		return ErrInvalidPoolIndex
	}
	if err = mp.pools[idx].Invoke(args); err == ErrPoolOverload {
		if fb := mp.lb.FallBack(mp.metrics); validIdx(fb, len(mp.pools)) {
			return mp.pools[fb].Invoke(args)
		}
	}
	return
}

// Running returns the number of the currently running workers across all pools.
func (mp *MultiPoolWithFunc) Running() (n int) {
	for _, pool := range mp.pools {
		n += pool.Running()
	}
	return
}

// RunningByIndex returns the number of the currently running workers in the specific pool.
func (mp *MultiPoolWithFunc) RunningByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Running(), nil
}

// Free returns the number of available workers across all pools.
func (mp *MultiPoolWithFunc) Free() (n int) {
	for _, pool := range mp.pools {
		n += pool.Free()
	}
	return
}

// FreeByIndex returns the number of available workers in the specific pool.
func (mp *MultiPoolWithFunc) FreeByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Free(), nil
}

// Waiting returns the number of the currently waiting tasks across all pools.
func (mp *MultiPoolWithFunc) Waiting() (n int) {
	for _, pool := range mp.pools {
		n += pool.Waiting()
	}
	return
}

// WaitingByIndex returns the number of the currently waiting tasks in the specific pool.
func (mp *MultiPoolWithFunc) WaitingByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Waiting(), nil
}

// Cap returns the capacity of this multi-pool.
func (mp *MultiPoolWithFunc) Cap() (n int) {
	for _, pool := range mp.pools {
		n += pool.Cap()
	}
	return
}

// Tune resizes each pool in multi-pool.
//
// Note that this method doesn't resize the overall
// capacity of multi-pool.
func (mp *MultiPoolWithFunc) Tune(size int) {
	for _, pool := range mp.pools {
		pool.Tune(size)
	}
}

// IsClosed indicates whether the multi-pool is closed.
func (mp *MultiPoolWithFunc) IsClosed() bool {
	return atomic.LoadInt32(&mp.state) == CLOSED
}

// ReleaseTimeout closes the multi-pool with a timeout,
// it waits all pools to be closed before timing out.
func (mp *MultiPoolWithFunc) ReleaseTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return mp.ReleaseContext(ctx)
}

// ReleaseContext closes the multi-pool with a context,
// it waits all pools to be closed before the context is done.
func (mp *MultiPoolWithFunc) ReleaseContext(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&mp.state, OPENED, CLOSED) {
		return ErrPoolClosed
	}

	pools := make([]contextReleaser, len(mp.pools))
	for i, p := range mp.pools {
		pools[i] = p
	}
	return releasePools(ctx, pools)
}

// Reboot reboots a released multi-pool.
func (mp *MultiPoolWithFunc) Reboot() {
	if atomic.CompareAndSwapInt32(&mp.state, CLOSED, OPENED) {
		for _, pool := range mp.pools {
			pool.Reboot()
		}
	}
}
