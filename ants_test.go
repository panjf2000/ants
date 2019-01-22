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
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/panjf2000/ants"
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
	GiB // 1073741824
	TiB // 1099511627776             (超过了int32的范围)
	PiB // 1125899906842624
	EiB // 1152921504606846976
	ZiB // 1180591620717411303424    (超过了int64的范围)
	YiB // 1208925819614629174706176
)

const (
	Param    = 100
	AntsSize = 1000
	TestSize = 10000
	n        = 100000
)

var curMem uint64

func TestAntsPoolWithFunc(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(AntsSize, func(i interface{}) {
		demoPoolFunc(i)
		wg.Done()
	})
	defer p.Release()

	for i := 0; i < n; i++ {
		wg.Add(1)
		p.Serve(Param)
	}
	wg.Wait()
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestAntsPool(t *testing.T) {
	defer ants.Release()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		ants.Submit(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	t.Logf("pool, capacity:%d", ants.Cap())
	t.Logf("pool, running workers number:%d", ants.Running())
	t.Logf("pool, free workers number:%d", ants.Free())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestPanicHandler(t *testing.T) {
	p0, err := ants.NewPool(10)
	if err != nil {
		t.Fatalf("create new pool failed: %s", err.Error())
	}
	defer p0.Release()
	var panicCounter int64
	var wg sync.WaitGroup
	p0.PanicHandler = func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
	}
	wg.Add(1)
	p0.Submit(func() {
		panic("test")
	})
	wg.Wait()
	c := atomic.LoadInt64(&panicCounter)
	if c != 1 {
		t.Errorf("panic handler didn't work, panicCounter: %d", c)
	}
	if p0.Running() != 0 {
		t.Errorf("pool should be empty after panic")
	}
	p1, err := ants.NewPoolWithFunc(10, func (p interface{}) {
		panic(p)
	})
	if err != nil {
		t.Fatalf("create new pool with func failed: %s", err.Error())
	}
	defer p1.Release()
	p1.PanicHandler = func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
	}
	wg.Add(1)
	p1.Serve("test")
	wg.Wait()
	c = atomic.LoadInt64(&panicCounter)
	if c != 2 {
		t.Errorf("panic handler didn't work, panicCounter: %d", c)
	}
	if p1.Running() != 0 {
		t.Errorf("pool should be empty after panic")
	}
}

func TestCodeCov(t *testing.T) {
	_, err := ants.NewTimingPool(-1, -1)
	t.Log(err)
	_, err = ants.NewTimingPool(1, -1)
	t.Log(err)
	_, err = ants.NewTimingPoolWithFunc(-1, -1, demoPoolFunc)
	t.Log(err)
	_, err = ants.NewTimingPoolWithFunc(1, -1, demoPoolFunc)
	t.Log(err)

	p0, _ := ants.NewPool(AntsSize)
	defer p0.Submit(demoFunc)
	defer p0.Release()
	for i := 0; i < n; i++ {
		p0.Submit(demoFunc)
	}
	t.Logf("pool, capacity:%d", p0.Cap())
	t.Logf("pool, running workers number:%d", p0.Running())
	t.Logf("pool, free workers number:%d", p0.Free())
	p0.ReSize(AntsSize)
	p0.ReSize(AntsSize / 2)
	t.Logf("pool, after resize, capacity:%d, running:%d", p0.Cap(), p0.Running())

	p, _ := ants.NewPoolWithFunc(TestSize, demoPoolFunc)
	defer p.Serve(Param)
	defer p.Release()
	for i := 0; i < n; i++ {
		p.Serve(Param)
	}
	time.Sleep(ants.DefaultCleanIntervalTime * time.Second)
	t.Logf("pool with func, capacity:%d", p.Cap())
	t.Logf("pool with func, running workers number:%d", p.Running())
	t.Logf("pool with func, free workers number:%d", p.Free())
	p.ReSize(TestSize)
	p.ReSize(AntsSize)
	t.Logf("pool with func, after resize, capacity:%d, running:%d", p.Cap(), p.Running())
}

func TestPurge(t *testing.T) {
	p, err := ants.NewTimingPool(10, 1)
	defer p.Release()
	if err != nil {
		t.Fatalf("create TimingPool failed: %s", err.Error())
	}
	p.Submit(demoFunc)
	time.Sleep(5 * time.Second)
	if p.Running() != 0 {
		t.Error("all p should be purged")
	}
	p1, err := ants.NewTimingPoolWithFunc(10, 1, demoPoolFunc)
	defer p1.Release()
	if err != nil {
		t.Fatalf("create TimingPoolWithFunc failed: %s", err.Error())
	}
	p1.Serve(1)
	time.Sleep(5 * time.Second)
	if p.Running() != 0 {
		t.Error("all p should be purged")
	}
}
