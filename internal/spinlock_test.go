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
	GoroutineNum = 6000
	SleepTime    = 10
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

func BenchmarkMutexWithGoroutineLock(b *testing.B) {
	m := sync.Mutex{}
	var wg sync.WaitGroup
	var f = func() {
		m.Lock()
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		m.Unlock()
		wg.Done()
	}
	wg.Add(GoroutineNum)
	for i := 0; i < GoroutineNum; i++ {
		go f()
	}
	wg.Wait()
}

func BenchmarkOriginSpinLockWithGoroutineLock(b *testing.B) {
	spin := GetOriginSpinLock()
	var wg sync.WaitGroup
	var f = func() {
		spin.Lock()
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		spin.Unlock()
		wg.Done()
	}
	wg.Add(GoroutineNum)
	for i := 0; i < GoroutineNum; i++ {
		go f()
	}
	wg.Wait()
}

func BenchmarkBackOffSpinLockWithGoroutineLock(b *testing.B) {
	spin := NewSpinLock()
	var wg sync.WaitGroup
	var f = func() {
		spin.Lock()
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
		spin.Unlock()
		wg.Done()
	}
	wg.Add(GoroutineNum)
	for i := 0; i < GoroutineNum; i++ {
		go f()
	}
	wg.Wait()
}
