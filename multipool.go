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
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

// LoadBalancingStrategy represents the type of load-balancing algorithm.
type LoadBalancingStrategy int

const (
	// RoundRobin distributes task to a list of pools in rotation.
	RoundRobin LoadBalancingStrategy = 1 << (iota + 1)

	// LeastTasks always selects the pool with the least number of pending tasks.
	LeastTasks
)

// MultiPool consists of multiple pools, from which you will benefit the
// performance improvement on basis of the fine-grained locking that reduces
// the lock contention.
// MultiPool is a good fit for the scenario where you have a large number of
// tasks to submit, and you don't want the single pool to be the bottleneck.
type MultiPool struct {
	pools []*Pool
	index uint32
	state int32
	lbs   LoadBalancingStrategy
}

// NewMultiPool instantiates a MultiPool with a size of the pool list and a size
// per pool, and the load-balancing strategy.
func NewMultiPool(size, sizePerPool int, lbs LoadBalancingStrategy, options ...Option) (*MultiPool, error) {
	if lbs != RoundRobin && lbs != LeastTasks {
		return nil, ErrInvalidLoadBalancingStrategy
	}
	pools := make([]*Pool, size)
	for i := 0; i < size; i++ {
		pool, err := NewPool(sizePerPool, options...)
		if err != nil {
			return nil, err
		}
		pools[i] = pool
	}
	return &MultiPool{pools: pools, lbs: lbs}, nil
}

func (mp *MultiPool) next(lbs LoadBalancingStrategy) (idx int) {
	switch lbs {
	case RoundRobin:
		if idx = int((atomic.AddUint32(&mp.index, 1) - 1) % uint32(len(mp.pools))); idx == -1 {
			idx = 0
		}
		return
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

// Submit submits a task to a pool selected by the load-balancing strategy.
func (mp *MultiPool) Submit(task func()) (err error) {
	if mp.IsClosed() {
		return ErrPoolClosed
	}
	if err = mp.pools[mp.next(mp.lbs)].Submit(task); err == nil {
		return
	}
	if err == ErrPoolOverload && mp.lbs == RoundRobin {
		return mp.pools[mp.next(LeastTasks)].Submit(task)
	}
	return
}

// Running returns the number of the currently running workers across all pools.
func (mp *MultiPool) Running() (n int) {
	for _, pool := range mp.pools {
		n += pool.Running()
	}
	return
}

// RunningByIndex returns the number of the currently running workers in the specific pool.
func (mp *MultiPool) RunningByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Running(), nil
}

// Free returns the number of available workers across all pools.
func (mp *MultiPool) Free() (n int) {
	for _, pool := range mp.pools {
		n += pool.Free()
	}
	return
}

// FreeByIndex returns the number of available workers in the specific pool.
func (mp *MultiPool) FreeByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Free(), nil
}

// Waiting returns the number of the currently waiting tasks across all pools.
func (mp *MultiPool) Waiting() (n int) {
	for _, pool := range mp.pools {
		n += pool.Waiting()
	}
	return
}

// WaitingByIndex returns the number of the currently waiting tasks in the specific pool.
func (mp *MultiPool) WaitingByIndex(idx int) (int, error) {
	if idx < 0 || idx >= len(mp.pools) {
		return -1, ErrInvalidPoolIndex
	}
	return mp.pools[idx].Waiting(), nil
}

// Cap returns the capacity of this multi-pool.
func (mp *MultiPool) Cap() (n int) {
	for _, pool := range mp.pools {
		n += pool.Cap()
	}
	return
}

// Tune resizes each pool in multi-pool.
//
// Note that this method doesn't resize the overall
// capacity of multi-pool.
func (mp *MultiPool) Tune(size int) {
	for _, pool := range mp.pools {
		pool.Tune(size)
	}
}

// IsClosed indicates whether the multi-pool is closed.
func (mp *MultiPool) IsClosed() bool {
	return atomic.LoadInt32(&mp.state) == CLOSED
}

// ReleaseTimeout closes the multi-pool with a timeout,
// it waits all pools to be closed before timing out.
func (mp *MultiPool) ReleaseTimeout(timeout time.Duration) error {
	if !atomic.CompareAndSwapInt32(&mp.state, OPENED, CLOSED) {
		return ErrPoolClosed
	}

	errCh := make(chan error, len(mp.pools))
	var wg errgroup.Group
	for i, pool := range mp.pools {
		func(p *Pool, idx int) {
			wg.Go(func() error {
				err := p.ReleaseTimeout(timeout)
				if err != nil {
					err = fmt.Errorf("pool %d: %v", idx, err)
				}
				errCh <- err
				return err
			})
		}(pool, i)
	}

	_ = wg.Wait()

	var errStr strings.Builder
	for i := 0; i < len(mp.pools); i++ {
		if err := <-errCh; err != nil {
			errStr.WriteString(err.Error())
			errStr.WriteString(" | ")
		}
	}

	if errStr.Len() == 0 {
		return nil
	}

	return errors.New(strings.TrimSuffix(errStr.String(), " | "))
}

// Reboot reboots a released multi-pool.
func (mp *MultiPool) Reboot() {
	if atomic.CompareAndSwapInt32(&mp.state, CLOSED, OPENED) {
		atomic.StoreUint32(&mp.index, 0)
		for _, pool := range mp.pools {
			pool.Reboot()
		}
	}
}
