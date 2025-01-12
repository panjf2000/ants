// Copyright 2021 Andy Pan & Dietoad. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

/*
Benchmark result for three types of locks:
	goos: darwin
	goarch: arm64
	pkg: github.com/panjf2000/ants/v2/pkg/sync
	BenchmarkMutex-10              	10452573	       111.1 ns/op	       0 B/op	       0 allocs/op
	BenchmarkSpinLock-10           	58953211	        18.01 ns/op	       0 B/op	       0 allocs/op
	BenchmarkBackOffSpinLock-10    	100000000	        10.81 ns/op	       0 B/op	       0 allocs/op
*/

type originSpinLock uint32

func (sl *originSpinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		runtime.Gosched()
	}
}

func (sl *originSpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func NewOriginSpinLock() sync.Locker {
	return new(originSpinLock)
}

func BenchmarkMutex(b *testing.B) {
	m := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			//nolint:staticcheck
			m.Unlock()
		}
	})
}

func BenchmarkSpinLock(b *testing.B) {
	spin := NewOriginSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			//nolint:staticcheck
			spin.Unlock()
		}
	})
}

func BenchmarkBackOffSpinLock(b *testing.B) {
	spin := NewSpinLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spin.Lock()
			//nolint:staticcheck
			spin.Unlock()
		}
	})
}
