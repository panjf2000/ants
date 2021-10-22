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
	RunTimes   = 1000
	SleepTime  = 15
)

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
		if wait < maxBackoff {
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
			time.Sleep(time.Duration(SleepTime) * time.Millisecond)
			m.Unlock()
		}
	})
}

func BenchmarkOriginSpinLockRunParalleWithFunc(b *testing.B) {
	spin := GetOriginSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			time.Sleep(time.Duration(SleepTime) * time.Millisecond)
			spin.Unlock()
		}
	})
}

func BenchmarkBackOffSpinLockRunParallelWithFunc(b *testing.B) {
	spin := GetBackOffSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			time.Sleep(time.Duration(SleepTime) * time.Millisecond)
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

