// Copyright 2021 Andy Pan & Dietoad. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package internal

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	Backoff    = 64
	RunTimes = 1000
	SleepTime = 15
)

func demoFunc(args interface{}) {
	n := args.(int)
	time.Sleep(time.Duration(n) * time.Millisecond)
}

type originSpinLock uint32

func (sl *originSpinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		runtime.Gosched()
	}
}

func (sl *originSpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func GetOriginSpinLock() sync.Locker {
	return new(originSpinLock)
}

type backOffSpinLock uint32

func (sl *backOffSpinLock) Lock() {
	wait := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		for i := 0; i < wait; i++ {
			runtime.Gosched()
		}
		if wait < Backoff {
			wait <<= 1
		}
	}
}

func (sl *backOffSpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func GetBackOffSpinLock() sync.Locker {
	return new(backOffSpinLock)
}

func BenchmarkMutex(b *testing.B) {
	m := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			demoFunc(SleepTime)
			m.Unlock()
		}
	})
}

func BenchmarkOriginSpinLockRunParalleWithFunc(b *testing.B) {
	spin := GetOriginSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			demoFunc(SleepTime)
			spin.Unlock()
		}
	})
}

func BenchmarkBackOffSpinLockRunParallelWithFunc(b *testing.B) {
	spin := GetBackOffSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			demoFunc(SleepTime)
			spin.Unlock()
		}
	})
}

func BenchmarkOriginSpinLockRunParalleWithGoroutine(b *testing.B) {
	spin := GetOriginSpinLock()
	var wg sync.WaitGroup
	var f= func() {
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		wg.Done()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			wg.Add( 1)
			go f()
			wg.Wait()
			spin.Unlock()
		}
	})
}

func BenchmarkBackOffSpinLockRunParallelWithGoroutine(b *testing.B) {
	spin := GetBackOffSpinLock()
	var wg sync.WaitGroup
	var f= func() {
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		wg.Done()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			wg.Add( 1)
			go f()
			wg.Wait()
			spin.Unlock()
		}
	})
}

func BenchmarkMutexWithGoroutineLock(b *testing.B) {
	m := sync.Mutex{}
	var wg sync.WaitGroup
	var f= func() {
		m.Lock()
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		m.Unlock()
		wg.Done()
	}
	wg.Add(RunTimes)
	for i:=0;i<RunTimes;i++ {
		go f()
	}
	wg.Wait()
}

func BenchmarkOriginSpinLockWithGoroutineLock(b *testing.B) {
	spin := GetOriginSpinLock()
	var wg sync.WaitGroup
	var f= func() {
		spin.Lock()
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		spin.Unlock()
		wg.Done()
	}
	wg.Add(RunTimes)
	for i:=0;i<RunTimes;i++ {
		go f()
	}
	wg.Wait()
}

func BenchmarkBackOffSpinLockWithGoroutineLock(b *testing.B) {
	spin := GetBackOffSpinLock()
	var wg sync.WaitGroup
	var f= func() {
		spin.Lock()
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		spin.Unlock()
		wg.Done()
	}
	wg.Add(RunTimes)
	for i:=0;i<RunTimes;i++ {
		go f()
	}
	wg.Wait()
}

func BenchmarkMutexWithGoroutine(b *testing.B) {
	m := sync.Mutex{}
	var wg sync.WaitGroup
	var f= func() {
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		wg.Done()
	}
	wg.Add(RunTimes)
	for i:=0;i<RunTimes;i++ {
		m.Lock()
		go f()
		m.Unlock()
	}
	wg.Wait()
}

func BenchmarkOriginSpinLockWithGoroutine(b *testing.B) {
	spin := GetOriginSpinLock()
	var wg sync.WaitGroup
	var f= func() {
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		wg.Done()
	}
	wg.Add(RunTimes)
	for i:=0;i<RunTimes;i++ {
		spin.Lock()
		go f()
		spin.Unlock()
	}
	wg.Wait()
}

func BenchmarkBackOffSpinLockWithGoroutine(b *testing.B) {
	spin := GetBackOffSpinLock()
	var wg sync.WaitGroup
	var f= func() {
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		wg.Done()
	}
	wg.Add(RunTimes)
	for i:=0;i<RunTimes;i++ {
		spin.Lock()
		go f()
		spin.Unlock()
	}
	wg.Wait()
}
