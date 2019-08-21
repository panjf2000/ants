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
	_ = 1 << (10 * iota)
	//KiB // 1024
	MiB // 1048576
	//GiB // 1073741824
	//TiB // 1099511627776             (超过了int32的范围)
	//PiB // 1125899906842624
	//EiB // 1152921504606846976
	//ZiB // 1180591620717411303424    (超过了int64的范围)
	//YiB // 1208925819614629174706176
)

const (
	Param    = 100
	AntsSize = 1000
	TestSize = 10000
	n        = 100000
)

var curMem uint64

// TestAntsPoolWaitToGetWorker is used to test waiting to get worker.
func TestAntsPoolWaitToGetWorker(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := ants.NewPool(AntsSize)
	defer p.Release()

	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Submit(func() {
			demoPoolFunc(Param)
			wg.Done()
		})
	}
	wg.Wait()
	t.Logf("pool, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestAntsPoolWaitToGetWorkerPreMalloc(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := ants.NewPool(AntsSize, ants.WithPreAlloc(true))
	defer p.Release()

	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Submit(func() {
			demoPoolFunc(Param)
			wg.Done()
		})
	}
	wg.Wait()
	t.Logf("pool, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

// TestAntsPoolWithFuncWaitToGetWorker is used to test waiting to get worker.
func TestAntsPoolWithFuncWaitToGetWorker(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(AntsSize, func(i interface{}) {
		demoPoolFunc(i)
		wg.Done()
	})
	defer p.Release()

	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Invoke(Param)
	}
	wg.Wait()
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestAntsPoolWithFuncWaitToGetWorkerPreMalloc(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(AntsSize, func(i interface{}) {
		demoPoolFunc(i)
		wg.Done()
	}, ants.WithPreAlloc(true))
	defer p.Release()

	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Invoke(Param)
	}
	wg.Wait()
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

// TestAntsPoolGetWorkerFromCache is used to test getting worker from sync.Pool.
func TestAntsPoolGetWorkerFromCache(t *testing.T) {
	p, _ := ants.NewPool(TestSize)
	defer p.Release()

	for i := 0; i < AntsSize; i++ {
		_ = p.Submit(demoFunc)
	}
	time.Sleep(2 * ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	_ = p.Submit(demoFunc)
	t.Logf("pool, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

// TestAntsPoolWithFuncGetWorkerFromCache is used to test getting worker from sync.Pool.
func TestAntsPoolWithFuncGetWorkerFromCache(t *testing.T) {
	dur := 10
	p, _ := ants.NewPoolWithFunc(TestSize, demoPoolFunc)
	defer p.Release()

	for i := 0; i < AntsSize; i++ {
		_ = p.Invoke(dur)
	}
	time.Sleep(2 * ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	_ = p.Invoke(dur)
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestAntsPoolWithFuncGetWorkerFromCachePreMalloc(t *testing.T) {
	dur := 10
	p, _ := ants.NewPoolWithFunc(TestSize, demoPoolFunc, ants.WithPreAlloc(true))
	defer p.Release()

	for i := 0; i < AntsSize; i++ {
		_ = p.Invoke(dur)
	}
	time.Sleep(2 * ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	_ = p.Invoke(dur)
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

//-------------------------------------------------------------------------------------------
// Contrast between goroutines without a pool and goroutines with ants pool.
//-------------------------------------------------------------------------------------------
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
		_ = ants.Submit(func() {
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

//-------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------

func TestPanicHandler(t *testing.T) {
	var panicCounter int64
	var wg sync.WaitGroup
	p0, err := ants.NewPool(10, ants.WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
		t.Logf("catch panic with PanicHandler: %v", p)
	}))
	if err != nil {
		t.Fatalf("create new pool failed: %s", err.Error())
	}
	defer p0.Release()
	wg.Add(1)
	_ = p0.Submit(func() {
		panic("Oops!")
	})
	wg.Wait()
	c := atomic.LoadInt64(&panicCounter)
	if c != 1 {
		t.Errorf("panic handler didn't work, panicCounter: %d", c)
	}
	if p0.Running() != 0 {
		t.Errorf("pool should be empty after panic")
	}

	p1, err := ants.NewPoolWithFunc(10, func(p interface{}) { panic(p) }, ants.WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
	}))
	if err != nil {
		t.Fatalf("create new pool with func failed: %s", err.Error())
	}
	defer p1.Release()
	wg.Add(1)
	_ = p1.Invoke("Oops!")
	wg.Wait()
	c = atomic.LoadInt64(&panicCounter)
	if c != 2 {
		t.Errorf("panic handler didn't work, panicCounter: %d", c)
	}
	if p1.Running() != 0 {
		t.Errorf("pool should be empty after panic")
	}
}

func TestPanicHandlerPreMalloc(t *testing.T) {
	var panicCounter int64
	var wg sync.WaitGroup
	p0, err := ants.NewPool(10, ants.WithPreAlloc(true), ants.WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
		t.Logf("catch panic with PanicHandler: %v", p)
	}))
	if err != nil {
		t.Fatalf("create new pool failed: %s", err.Error())
	}
	defer p0.Release()
	wg.Add(1)
	_ = p0.Submit(func() {
		panic("Oops!")
	})
	wg.Wait()
	c := atomic.LoadInt64(&panicCounter)
	if c != 1 {
		t.Errorf("panic handler didn't work, panicCounter: %d", c)
	}
	if p0.Running() != 0 {
		t.Errorf("pool should be empty after panic")
	}

	p1, err := ants.NewPoolWithFunc(10, func(p interface{}) { panic(p) }, ants.WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
	}))
	if err != nil {
		t.Fatalf("create new pool with func failed: %s", err.Error())
	}
	defer p1.Release()
	wg.Add(1)
	_ = p1.Invoke("Oops!")
	wg.Wait()
	c = atomic.LoadInt64(&panicCounter)
	if c != 2 {
		t.Errorf("panic handler didn't work, panicCounter: %d", c)
	}
	if p1.Running() != 0 {
		t.Errorf("pool should be empty after panic")
	}
}

func TestPoolPanicWithoutHandler(t *testing.T) {
	p0, err := ants.NewPool(10)
	if err != nil {
		t.Fatalf("create new pool failed: %s", err.Error())
	}
	defer p0.Release()
	_ = p0.Submit(func() {
		panic("Oops!")
	})

	p1, err := ants.NewPoolWithFunc(10, func(p interface{}) {
		panic(p)
	})
	if err != nil {
		t.Fatalf("create new pool with func failed: %s", err.Error())
	}
	defer p1.Release()
	_ = p1.Invoke("Oops!")
}

func TestPoolPanicWithoutHandlerPreMalloc(t *testing.T) {
	p0, err := ants.NewPool(10, ants.WithPreAlloc(true))
	if err != nil {
		t.Fatalf("create new pool failed: %s", err.Error())
	}
	defer p0.Release()
	_ = p0.Submit(func() {
		panic("Oops!")
	})

	p1, err := ants.NewPoolWithFunc(10, func(p interface{}) {
		panic(p)
	})
	if err != nil {
		t.Fatalf("create new pool with func failed: %s", err.Error())
	}
	defer p1.Release()
	_ = p1.Invoke("Oops!")
}

func TestPurge(t *testing.T) {
	p, err := ants.NewPool(10)
	if err != nil {
		t.Fatalf("create TimingPool failed: %s", err.Error())
	}
	defer p.Release()
	_ = p.Submit(demoFunc)
	time.Sleep(3 * ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	if p.Running() != 0 {
		t.Error("all p should be purged")
	}
	p1, err := ants.NewPoolWithFunc(10, demoPoolFunc)
	if err != nil {
		t.Fatalf("create TimingPoolWithFunc failed: %s", err.Error())
	}
	defer p1.Release()
	_ = p1.Invoke(1)
	time.Sleep(3 * ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	if p.Running() != 0 {
		t.Error("all p should be purged")
	}
}

func TestPurgePreMalloc(t *testing.T) {
	p, err := ants.NewPool(10, ants.WithPreAlloc(true))
	if err != nil {
		t.Fatalf("create TimingPool failed: %s", err.Error())
	}
	defer p.Release()
	_ = p.Submit(demoFunc)
	time.Sleep(3 * ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	if p.Running() != 0 {
		t.Error("all p should be purged")
	}
	p1, err := ants.NewPoolWithFunc(10, demoPoolFunc)
	if err != nil {
		t.Fatalf("create TimingPoolWithFunc failed: %s", err.Error())
	}
	defer p1.Release()
	_ = p1.Invoke(1)
	time.Sleep(3 * ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	if p.Running() != 0 {
		t.Error("all p should be purged")
	}
}

func TestNonblockingSubmit(t *testing.T) {
	poolSize := 10
	p, err := ants.NewPool(poolSize, ants.WithNonblocking(true))
	if err != nil {
		t.Fatalf("create TimingPool failed: %s", err.Error())
	}
	defer p.Release()
	for i := 0; i < poolSize-1; i++ {
		if err := p.Submit(longRunningFunc); err != nil {
			t.Fatalf("nonblocking submit when pool is not full shouldn't return error")
		}
	}
	ch := make(chan struct{})
	f := func() {
		<-ch
	}
	// p is full now.
	if err := p.Submit(f); err != nil {
		t.Fatalf("nonblocking submit when pool is not full shouldn't return error")
	}
	if err := p.Submit(demoFunc); err == nil || err != ants.ErrPoolOverload {
		t.Fatalf("nonblocking submit when pool is full should get an ErrPoolOverload")
	}
	// interrupt f to get an available worker
	close(ch)
	time.Sleep(1 * time.Second)
	if err := p.Submit(demoFunc); err != nil {
		t.Fatalf("nonblocking submit when pool is not full shouldn't return error")
	}
}

func TestMaxBlockingSubmit(t *testing.T) {
	poolSize := 10
	p, err := ants.NewPool(poolSize, ants.WithMaxBlockingTasks(1))
	if err != nil {
		t.Fatalf("create TimingPool failed: %s", err.Error())
	}
	defer p.Release()
	for i := 0; i < poolSize-1; i++ {
		if err := p.Submit(longRunningFunc); err != nil {
			t.Fatalf("submit when pool is not full shouldn't return error")
		}
	}
	ch := make(chan struct{})
	f := func() {
		<-ch
	}
	// p is full now.
	if err := p.Submit(f); err != nil {
		t.Fatalf("submit when pool is not full shouldn't return error")
	}
	var wg sync.WaitGroup
	wg.Add(1)
	errCh := make(chan error, 1)
	go func() {
		// should be blocked. blocking num == 1
		if err := p.Submit(demoFunc); err != nil {
			errCh <- err
		}
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	// already reached max blocking limit
	if err := p.Submit(demoFunc); err != ants.ErrPoolOverload {
		t.Fatalf("blocking submit when pool reach max blocking submit should return ErrPoolOverload")
	}
	// interrupt f to make blocking submit successful.
	close(ch)
	wg.Wait()
	select {
	case <-errCh:
		t.Fatalf("blocking submit when pool is full should not return error")
	default:
	}
}

func TestNonblockingSubmitWithFunc(t *testing.T) {
	poolSize := 10
	p, err := ants.NewPoolWithFunc(poolSize, longRunningPoolFunc, ants.WithNonblocking(true))
	if err != nil {
		t.Fatalf("create TimingPool failed: %s", err.Error())
	}
	defer p.Release()
	for i := 0; i < poolSize-1; i++ {
		if err := p.Invoke(nil); err != nil {
			t.Fatalf("nonblocking submit when pool is not full shouldn't return error")
		}
	}
	ch := make(chan struct{})
	// p is full now.
	if err := p.Invoke(ch); err != nil {
		t.Fatalf("nonblocking submit when pool is not full shouldn't return error")
	}
	if err := p.Invoke(nil); err == nil || err != ants.ErrPoolOverload {
		t.Fatalf("nonblocking submit when pool is full should get an ErrPoolOverload")
	}
	// interrupt f to get an available worker
	close(ch)
	time.Sleep(1 * time.Second)
	if err := p.Invoke(nil); err != nil {
		t.Fatalf("nonblocking submit when pool is not full shouldn't return error")
	}
}

func TestMaxBlockingSubmitWithFunc(t *testing.T) {
	poolSize := 10
	p, err := ants.NewPoolWithFunc(poolSize, longRunningPoolFunc, ants.WithMaxBlockingTasks(1))
	if err != nil {
		t.Fatalf("create TimingPool failed: %s", err.Error())
	}
	defer p.Release()
	for i := 0; i < poolSize-1; i++ {
		if err := p.Invoke(Param); err != nil {
			t.Fatalf("submit when pool is not full shouldn't return error")
		}
	}
	ch := make(chan struct{})
	// p is full now.
	if err := p.Invoke(ch); err != nil {
		t.Fatalf("submit when pool is not full shouldn't return error")
	}
	var wg sync.WaitGroup
	wg.Add(1)
	errCh := make(chan error, 1)
	go func() {
		// should be blocked. blocking num == 1
		if err := p.Invoke(Param); err != nil {
			errCh <- err
		}
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	// already reached max blocking limit
	if err := p.Invoke(Param); err != ants.ErrPoolOverload {
		t.Fatalf("blocking submit when pool reach max blocking submit should return ErrPoolOverload: %v", err)
	}
	// interrupt one func to make blocking submit successful.
	close(ch)
	wg.Wait()
	select {
	case <-errCh:
		t.Fatalf("blocking submit when pool is full should not return error")
	default:
	}
}
func TestRestCodeCoverage(t *testing.T) {
	_, err := ants.NewPool(-1, ants.WithExpiryDuration(-1))
	t.Log(err)
	_, err = ants.NewPool(1, ants.WithExpiryDuration(-1))
	t.Log(err)
	_, err = ants.NewPoolWithFunc(-1, demoPoolFunc, ants.WithExpiryDuration(-1))
	t.Log(err)
	_, err = ants.NewPoolWithFunc(1, demoPoolFunc, ants.WithExpiryDuration(-1))
	t.Log(err)

	options := ants.Options{}
	options.ExpiryDuration = time.Duration(10) * time.Second
	options.Nonblocking = true
	options.PreAlloc = true
	poolOpts, _ := ants.NewPool(1, ants.WithOptions(options))
	t.Logf("Pool with options, capacity: %d", poolOpts.Cap())

	p0, _ := ants.NewPool(TestSize)
	defer func() {
		_ = p0.Submit(demoFunc)
	}()
	defer p0.Release()
	for i := 0; i < n; i++ {
		_ = p0.Submit(demoFunc)
	}
	t.Logf("pool, capacity:%d", p0.Cap())
	t.Logf("pool, running workers number:%d", p0.Running())
	t.Logf("pool, free workers number:%d", p0.Free())
	p0.Tune(TestSize)
	p0.Tune(TestSize / 10)
	t.Logf("pool, after tuning capacity, capacity:%d, running:%d", p0.Cap(), p0.Running())

	pprem, _ := ants.NewPool(TestSize, ants.WithPreAlloc(true))
	defer func() {
		_ = pprem.Submit(demoFunc)
	}()
	defer pprem.Release()
	for i := 0; i < n; i++ {
		_ = pprem.Submit(demoFunc)
	}
	t.Logf("pre-malloc pool, capacity:%d", pprem.Cap())
	t.Logf("pre-malloc pool, running workers number:%d", pprem.Running())
	t.Logf("pre-malloc pool, free workers number:%d", pprem.Free())
	pprem.Tune(TestSize)
	pprem.Tune(TestSize / 10)
	t.Logf("pre-malloc pool, after tuning capacity, capacity:%d, running:%d", pprem.Cap(), pprem.Running())

	p, _ := ants.NewPoolWithFunc(TestSize, demoPoolFunc)
	defer func() {
		_ = p.Invoke(Param)
	}()
	defer p.Release()
	for i := 0; i < n; i++ {
		_ = p.Invoke(Param)
	}
	time.Sleep(ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	t.Logf("pool with func, capacity:%d", p.Cap())
	t.Logf("pool with func, running workers number:%d", p.Running())
	t.Logf("pool with func, free workers number:%d", p.Free())
	p.Tune(TestSize)
	p.Tune(TestSize / 10)
	t.Logf("pool with func, after tuning capacity, capacity:%d, running:%d", p.Cap(), p.Running())

	ppremWithFunc, _ := ants.NewPoolWithFunc(TestSize, demoPoolFunc, ants.WithPreAlloc(true))
	defer func() {
		_ = ppremWithFunc.Invoke(Param)
	}()
	defer ppremWithFunc.Release()
	for i := 0; i < n; i++ {
		_ = ppremWithFunc.Invoke(Param)
	}
	time.Sleep(ants.DEFAULT_CLEAN_INTERVAL_TIME * time.Second)
	t.Logf("pre-malloc pool with func, capacity:%d", ppremWithFunc.Cap())
	t.Logf("pre-malloc pool with func, running workers number:%d", ppremWithFunc.Running())
	t.Logf("pre-malloc pool with func, free workers number:%d", ppremWithFunc.Free())
	ppremWithFunc.Tune(TestSize)
	ppremWithFunc.Tune(TestSize / 10)
	t.Logf("pre-malloc pool with func, after tuning capacity, capacity:%d, running:%d", ppremWithFunc.Cap(),
		ppremWithFunc.Running())
}
