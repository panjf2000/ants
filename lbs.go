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
	"math"
	"sync/atomic"
)

// LoadBalancingStrategy represents the type of load-balancing algorithm.
type LoadBalancingStrategy int

const (
	// RoundRobin distributes task to a list of pools in rotation.
	RoundRobin LoadBalancingStrategy = 1 << (iota + 1)

	// LeastTasks always selects the pool with the least number of pending tasks.
	LeastTasks
)

// PoolMetrics exposes the read-only stats a LoadBalancer needs to make a pick decision.
type PoolMetrics interface {
	Running() int
	Waiting() int
	Free() int
	Cap() int
}

// LoadBalancer picks a pool index from a slice of PoolMetrics.
type LoadBalancer interface {
	Pick(pools []PoolMetrics) int
	// Fallback is called when the pool chosen by Pick is overloaded.
	// Return -1 to indicate no fallback is supported.
	Fallback(pools []PoolMetrics) int
}

// roundRobinLB distributes tasks across pools in rotation.
type roundRobinLB struct {
	index uint32
}

// NewRoundRobinLB returns a RoundRobin load balancer.
func NewRoundRobinLB() LoadBalancer {
	return &roundRobinLB{index: math.MaxUint32}
}

func (r *roundRobinLB) Pick(pools []PoolMetrics) int {
	return int(atomic.AddUint32(&r.index, 1) % uint32(len(pools)))
}

func (r *roundRobinLB) Fallback(pools []PoolMetrics) int {
	return leastTasksPick(pools)
}

func leastTasksPick(pools []PoolMetrics) int {
	idx, least := 0, math.MaxInt32
	for i, p := range pools {
		if n := p.Running(); n < least {
			least = n
			idx = i
		}
	}
	return idx
}

// leastTasksLB picks the pool with the fewest running tasks.
type leastTasksLB struct{}

// NewLeastTasksLB returns a LeastTasks load balancer.
func NewLeastTasksLB() LoadBalancer {
	return &leastTasksLB{}
}

func (l *leastTasksLB) Pick(pools []PoolMetrics) int {
	return leastTasksPick(pools)
}

func (l *leastTasksLB) Fallback(pools []PoolMetrics) int {
	return -1
}

// leastWaiting picks the pool with the fewest waiting tasks.
type leastWaiting struct{}

// NewLeastWaitingLB returns a LeastWaiting load balancer.
func NewLeastWaitingLB() LoadBalancer {
	return &leastWaiting{}
}

func (l *leastWaiting) Pick(pools []PoolMetrics) int {
	idx, least := 0, math.MaxInt32
	for i, p := range pools {
		if n := p.Waiting(); n < least {
			least = n
			idx = i
		}
	}
	return idx
}

func (l *leastWaiting) Fallback(pools []PoolMetrics) int {
	return -1
}

// validIdx checks that idx returned by a LoadBalancer is within bounds.
func validIdx(idx, n int) bool {
	return idx >= 0 && idx < n
}

func newBuiltinLB(lbs LoadBalancingStrategy) (LoadBalancer, error) {
	switch lbs {
	case RoundRobin:
		return NewRoundRobinLB(), nil
	case LeastTasks:
		return NewLeastTasksLB(), nil
	}
	return nil, ErrInvalidLoadBalancingStrategy
}
