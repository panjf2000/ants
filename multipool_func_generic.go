// MIT License

// Copyright (c) 2025 Andy Pan

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
	"math"
	"sync/atomic"
	"time"
)

// MultiPoolWithFuncGeneric is the generic version of MultiPoolWithFunc.
type MultiPoolWithFuncGeneric[T any] struct {
	pools []*PoolWithFuncGeneric[T]
	index uint32
	state int32
	lbs   LoadBalancingStrategy
}

// NewMultiPoolWithFuncGeneric instantiates a MultiPoolWithFunc with a size of the pool list and a size
// per pool, and the load-balancing strategy.
func NewMultiPoolWithFuncGeneric[T any](size, sizePerPool int, fn func(T), lbs LoadBalancingStrategy, options ...Option) (*MultiPoolWithFuncGeneric[T], error) {
	if size <= 0 {
		return nil, ErrInvalidMultiPoolSize
	}

	if lbs != RoundRobin && lbs != LeastTasks {
		return nil, ErrInvalidLoadBalancingStrategy
	}
	pools := make([]*PoolWithFuncGeneric[T], size)
	for i := 0; i < size; i++ {
		pool, err := NewPoolWithFuncGeneric(sizePerPool, fn, options...)
		if err != nil {
			// Release all previously created pools to avoid resource leak
			for j := 0; j < i; j++ {
				pools[j].Release()
			}
			return nil, err
		}
		pools[i] = pool
	}
	return &MultiPoolWithFuncGeneric[T]{pools: pools, index: math.MaxUint32, lbs: lbs}, nil
}

func (mp *MultiPoolWithFuncGeneric[T]) next(lbs LoadBalancingStrategy) (idx int) {
	switch lbs {
	case RoundRobin:
		return int(atomic.AddUint32(&mp.index, 1) % uint32(len(mp.pools)))
	case LeastTasks:
		leastTasks := 1<<31 - 1
		for i, pool := range mp.pools {
			if n := pool.Running(); n < leastTasks {
				leastTasks = n
				idx = i
			}
		}
		return
	}
	return -1
}

// Invoke submits a task to a pool selected by the load-balancing strategy.
func (mp *MultiPoolWithFuncGeneric[T]) Invoke(args T) (err error) {
	if mp.IsClosed() {
		return ErrPoolClosed
	}

	if err = mp.pools[mp.next(mp.lbs)].Invoke(args); err == nil {
		return
	}
	if err == ErrPoolOverload && mp.lbs == RoundRobin {
		return mp.pools[mp.next(LeastTasks)].Invoke(args)
	}
	return
}

// Running returns the number of the currently running workers across all pools.
func (mp *MultiPoolWithFuncGeneric[T]) Running() (n int) {
	for _, pool := range mp.pools {
		n += pool.Running()
	}
	return
}

// RunningByIndex returns the number of the currently running workers in the specific pool.
func (mp *MultiPoolWithFuncGeneric[T]) RunningByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Running(), nil
}

// Free returns the number of available workers across all pools.
func (mp *MultiPoolWithFuncGeneric[T]) Free() (n int) {
	for _, pool := range mp.pools {
		n += pool.Free()
	}
	return
}

// FreeByIndex returns the number of available workers in the specific pool.
func (mp *MultiPoolWithFuncGeneric[T]) FreeByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Free(), nil
}

// Waiting returns the number of the currently waiting tasks across all pools.
func (mp *MultiPoolWithFuncGeneric[T]) Waiting() (n int) {
	for _, pool := range mp.pools {
		n += pool.Waiting()
	}
	return
}

// WaitingByIndex returns the number of the currently waiting tasks in the specific pool.
func (mp *MultiPoolWithFuncGeneric[T]) WaitingByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Waiting(), nil
}

// Cap returns the capacity of this multi-pool.
func (mp *MultiPoolWithFuncGeneric[T]) Cap() (n int) {
	for _, pool := range mp.pools {
		n += pool.Cap()
	}
	return
}

// Tune resizes each pool in multi-pool.
//
// Note that this method doesn't resize the overall
// capacity of multi-pool.
func (mp *MultiPoolWithFuncGeneric[T]) Tune(size int) {
	for _, pool := range mp.pools {
		pool.Tune(size)
	}
}

// IsClosed indicates whether the multi-pool is closed.
func (mp *MultiPoolWithFuncGeneric[T]) IsClosed() bool {
	return atomic.LoadInt32(&mp.state) == CLOSED
}

// ReleaseTimeout closes the multi-pool with a timeout,
// it waits all pools to be closed before timing out.
func (mp *MultiPoolWithFuncGeneric[T]) ReleaseTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return mp.ReleaseContext(ctx)
}

// ReleaseContext closes the multi-pool with a context,
// it waits all pools to be closed before the context is done.
func (mp *MultiPoolWithFuncGeneric[T]) ReleaseContext(ctx context.Context) error {
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
func (mp *MultiPoolWithFuncGeneric[T]) Reboot() {
	if atomic.CompareAndSwapInt32(&mp.state, CLOSED, OPENED) {
		atomic.StoreUint32(&mp.index, 0)
		for _, pool := range mp.pools {
			pool.Reboot()
		}
	}
}
