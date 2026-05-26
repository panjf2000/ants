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

package ants_test

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/panjf2000/ants/v2"
)

const (
	RunTimes           = 1e6
	PoolCap            = 5e4
	BenchParam         = 10
	DefaultExpiredTime = 10 * time.Second
)

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func demoPoolFunc(args any) {
	n := args.(int)
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func demoPoolFuncInt(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}

var stopLongRunningFunc int32

func longRunningFunc() {
	for atomic.LoadInt32(&stopLongRunningFunc) == 0 {
		runtime.Gosched()
	}
}

func longRunningPoolFunc(arg any) {
	<-arg.(chan struct{})
}

func longRunningPoolFuncCh(ch chan struct{}) {
	<-ch
}

func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demoFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkChannel(b *testing.B) {
	var wg sync.WaitGroup
	sema := make(chan struct{}, PoolCap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			sema <- struct{}{}
			go func() {
				demoFunc()
				<-sema
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkErrGroup(b *testing.B) {
	var wg sync.WaitGroup
	var pool errgroup.Group
	pool.SetLimit(PoolCap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			pool.Go(func() error {
				demoFunc()
				wg.Done()
				return nil
			})
		}
		wg.Wait()
	}
}

func BenchmarkAntsPool(b *testing.B) {
	var wg sync.WaitGroup
	p, _ := ants.NewPool(PoolCap, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func BenchmarkAntsMultiPool(b *testing.B) {
	var wg sync.WaitGroup
	p, _ := ants.NewMultiPool(10, PoolCap/10, ants.RoundRobin, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.ReleaseTimeout(DefaultExpiredTime) //nolint:errcheck

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
}

func BenchmarkGoroutinesThroughput(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			go demoFunc()
		}
	}
}

func BenchmarkSemaphoreThroughput(b *testing.B) {
	sema := make(chan struct{}, PoolCap)
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			sema <- struct{}{}
			go func() {
				demoFunc()
				<-sema
			}()
		}
	}
}

func BenchmarkAntsPoolThroughput(b *testing.B) {
	p, _ := ants.NewPool(PoolCap, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(demoFunc)
		}
	}
}

func BenchmarkAntsMultiPoolThroughput(b *testing.B) {
	p, _ := ants.NewMultiPool(10, PoolCap/10, ants.RoundRobin, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.ReleaseTimeout(DefaultExpiredTime) //nolint:errcheck

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(demoFunc)
		}
	}
}

func BenchmarkParallelAntsPoolThroughput(b *testing.B) {
	p, _ := ants.NewPool(PoolCap, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.Release()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Submit(demoFunc)
		}
	})
}

func BenchmarkParallelAntsMultiPoolThroughput(b *testing.B) {
	p, _ := ants.NewMultiPool(10, PoolCap/10, ants.RoundRobin, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.ReleaseTimeout(DefaultExpiredTime) //nolint:errcheck

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Submit(demoFunc)
		}
	})
}

// cpuTask simulates a CPU-intensive task.
func cpuTask() {
	n := 0
	for i := 0; i < 1000; i++ {
		n += i * i
	}
	_ = n
}

// ioTask simulates an IO-intensive task with a short sleep.
func ioTask() {
	time.Sleep(time.Millisecond)
}

// mixedTask simulates uneven task durations to stress load-balancing decisions.
func mixedTask() {
	if time.Now().UnixNano()%5 == 0 {
		time.Sleep(10 * time.Millisecond)
	} else {
		time.Sleep(time.Millisecond)
	}
}

func benchmarkMultiPoolLBS(b *testing.B, lb ants.LoadBalancer, task func()) {
	p, _ := ants.NewMultiPoolWithLB(10, PoolCap/10, lb, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.ReleaseTimeout(DefaultExpiredTime) //nolint:errcheck

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Submit(task)
		}
	})
}

// CPU-intensive task benchmarks across LBS strategies.

func BenchmarkMultiPool_RoundRobin_CPUThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewRoundRobinLB(), cpuTask)
}

func BenchmarkMultiPool_LeastTasks_CPUThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewLeastTasksLB(), cpuTask)
}

func BenchmarkMultiPool_LeastWaiting_CPUThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewLeastWaitingLB(), cpuTask)
}

// IO-intensive task benchmarks across LBS strategies.

func BenchmarkMultiPool_RoundRobin_IOThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewRoundRobinLB(), ioTask)
}

func BenchmarkMultiPool_LeastTasks_IOThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewLeastTasksLB(), ioTask)
}

func BenchmarkMultiPool_LeastWaiting_IOThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewLeastWaitingLB(), ioTask)
}

// Mixed (uneven duration) task benchmarks across LBS strategies.

func BenchmarkMultiPool_RoundRobin_MixedThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewRoundRobinLB(), mixedTask)
}

func BenchmarkMultiPool_LeastTasks_MixedThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewLeastTasksLB(), mixedTask)
}

func BenchmarkMultiPool_LeastWaiting_MixedThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, ants.NewLeastWaitingLB(), mixedTask)
}

// randomLB is a custom LoadBalancer that picks a pool at random,
// demonstrating how users can plug in their own strategy via NewMultiPoolWithLB.
type randomLB struct{}

func newRandomLB() *randomLB {
	return &randomLB{}
}

func (r *randomLB) Pick(pools []ants.PoolMetrics) int {
	return rand.Intn(len(pools))
}

func (r *randomLB) Fallback(pools []ants.PoolMetrics) int {
	return -1
}

// Custom random LB benchmarks across task types.

func BenchmarkMultiPool_Random_CPUThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, newRandomLB(), cpuTask)
}

func BenchmarkMultiPool_Random_IOThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, newRandomLB(), ioTask)
}

func BenchmarkMultiPool_Random_MixedThroughput(b *testing.B) {
	benchmarkMultiPoolLBS(b, newRandomLB(), mixedTask)
}

func benchmarkMultiPoolWithFuncLBSThroughput(b *testing.B, lb ants.LoadBalancer) {
	p, _ := ants.NewMultiPoolWithFuncAndLB(10, PoolCap/10, demoPoolFunc, lb, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.ReleaseTimeout(DefaultExpiredTime) //nolint:errcheck

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Invoke(BenchParam)
		}
	})
}

func BenchmarkMultiPoolWithFunc_RoundRobin_Throughput(b *testing.B) {
	benchmarkMultiPoolWithFuncLBSThroughput(b, ants.NewRoundRobinLB())
}

func BenchmarkMultiPoolWithFunc_LeastTasks_Throughput(b *testing.B) {
	benchmarkMultiPoolWithFuncLBSThroughput(b, ants.NewLeastTasksLB())
}

func BenchmarkMultiPoolWithFunc_LeastWaiting_Throughput(b *testing.B) {
	benchmarkMultiPoolWithFuncLBSThroughput(b, ants.NewLeastWaitingLB())
}

func benchmarkMultiPoolWithFuncGenericLBSThroughput(b *testing.B, lb ants.LoadBalancer) {
	p, _ := ants.NewMultiPoolWithFuncGenericAndLB(10, PoolCap/10, demoPoolFuncInt, lb, ants.WithExpiryDuration(DefaultExpiredTime))
	defer p.ReleaseTimeout(DefaultExpiredTime) //nolint:errcheck

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Invoke(BenchParam)
		}
	})
}

func BenchmarkMultiPoolWithFuncGeneric_RoundRobin_Throughput(b *testing.B) {
	benchmarkMultiPoolWithFuncGenericLBSThroughput(b, ants.NewRoundRobinLB())
}

func BenchmarkMultiPoolWithFuncGeneric_LeastTasks_Throughput(b *testing.B) {
	benchmarkMultiPoolWithFuncGenericLBSThroughput(b, ants.NewLeastTasksLB())
}

func BenchmarkMultiPoolWithFuncGeneric_LeastWaiting_Throughput(b *testing.B) {
	benchmarkMultiPoolWithFuncGenericLBSThroughput(b, ants.NewLeastWaitingLB())
}
