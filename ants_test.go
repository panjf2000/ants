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

package ants

import (
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
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
	p, _ := NewPool(AntsSize)
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
	p, _ := NewPool(AntsSize, WithPreAlloc(true))
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
	p, _ := NewPoolWithFunc(AntsSize, func(i interface{}) {
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
	p, _ := NewPoolWithFunc(AntsSize, func(i interface{}) {
		demoPoolFunc(i)
		wg.Done()
	}, WithPreAlloc(true))
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
	p, _ := NewPool(TestSize)
	defer p.Release()

	for i := 0; i < AntsSize; i++ {
		_ = p.Submit(demoFunc)
	}
	time.Sleep(2 * DefaultCleanIntervalTime)
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
	p, _ := NewPoolWithFunc(TestSize, demoPoolFunc)
	defer p.Release()

	for i := 0; i < AntsSize; i++ {
		_ = p.Invoke(dur)
	}
	time.Sleep(2 * DefaultCleanIntervalTime)
	_ = p.Invoke(dur)
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestAntsPoolWithFuncGetWorkerFromCachePreMalloc(t *testing.T) {
	dur := 10
	p, _ := NewPoolWithFunc(TestSize, demoPoolFunc, WithPreAlloc(true))
	defer p.Release()

	for i := 0; i < AntsSize; i++ {
		_ = p.Invoke(dur)
	}
	time.Sleep(2 * DefaultCleanIntervalTime)
	_ = p.Invoke(dur)
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

// Contrast between goroutines without a pool and goroutines with ants pool.

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
	defer Release()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = Submit(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	t.Logf("pool, capacity:%d", Cap())
	t.Logf("pool, running workers number:%d", Running())
	t.Logf("pool, free workers number:%d", Free())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestPanicHandler(t *testing.T) {
	var panicCounter int64
	var wg sync.WaitGroup
	p0, err := NewPool(10, WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
		t.Logf("catch panic with PanicHandler: %v", p)
	}))
	assert.NoErrorf(t, err, "create new pool failed: %v", err)
	defer p0.Release()
	wg.Add(1)
	_ = p0.Submit(func() {
		panic("Oops!")
	})
	wg.Wait()
	c := atomic.LoadInt64(&panicCounter)
	assert.EqualValuesf(t, 1, c, "panic handler didn't work, panicCounter: %d", c)
	assert.EqualValues(t, 0, p0.Running(), "pool should be empty after panic")
	p1, err := NewPoolWithFunc(10, func(p interface{}) { panic(p) }, WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
	}))
	assert.NoErrorf(t, err, "create new pool with func failed: %v", err)
	defer p1.Release()
	wg.Add(1)
	_ = p1.Invoke("Oops!")
	wg.Wait()
	c = atomic.LoadInt64(&panicCounter)
	assert.EqualValuesf(t, 2, c, "panic handler didn't work, panicCounter: %d", c)
	assert.EqualValues(t, 0, p1.Running(), "pool should be empty after panic")
}

func TestPanicHandlerPreMalloc(t *testing.T) {
	var panicCounter int64
	var wg sync.WaitGroup
	p0, err := NewPool(10, WithPreAlloc(true), WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
		t.Logf("catch panic with PanicHandler: %v", p)
	}))
	assert.NoErrorf(t, err, "create new pool failed: %v", err)
	defer p0.Release()
	wg.Add(1)
	_ = p0.Submit(func() {
		panic("Oops!")
	})
	wg.Wait()
	c := atomic.LoadInt64(&panicCounter)
	assert.EqualValuesf(t, 1, c, "panic handler didn't work, panicCounter: %d", c)
	assert.EqualValues(t, 0, p0.Running(), "pool should be empty after panic")
	p1, err := NewPoolWithFunc(10, func(p interface{}) { panic(p) }, WithPanicHandler(func(p interface{}) {
		defer wg.Done()
		atomic.AddInt64(&panicCounter, 1)
	}))
	assert.NoErrorf(t, err, "create new pool with func failed: %v", err)
	defer p1.Release()
	wg.Add(1)
	_ = p1.Invoke("Oops!")
	wg.Wait()
	c = atomic.LoadInt64(&panicCounter)
	assert.EqualValuesf(t, 2, c, "panic handler didn't work, panicCounter: %d", c)
	assert.EqualValues(t, 0, p1.Running(), "pool should be empty after panic")
}

func TestPoolPanicWithoutHandler(t *testing.T) {
	p0, err := NewPool(10)
	assert.NoErrorf(t, err, "create new pool failed: %v", err)
	defer p0.Release()
	_ = p0.Submit(func() {
		panic("Oops!")
	})

	p1, err := NewPoolWithFunc(10, func(p interface{}) {
		panic(p)
	})
	assert.NoErrorf(t, err, "create new pool with func failed: %v", err)
	defer p1.Release()
	_ = p1.Invoke("Oops!")
}

func TestPoolPanicWithoutHandlerPreMalloc(t *testing.T) {
	p0, err := NewPool(10, WithPreAlloc(true))
	assert.NoErrorf(t, err, "create new pool failed: %v", err)
	defer p0.Release()
	_ = p0.Submit(func() {
		panic("Oops!")
	})

	p1, err := NewPoolWithFunc(10, func(p interface{}) {
		panic(p)
	})

	assert.NoErrorf(t, err, "create new pool with func failed: %v", err)

	defer p1.Release()
	_ = p1.Invoke("Oops!")
}

func TestPurgePool(t *testing.T) {
	size := 500
	ch := make(chan struct{})

	p, err := NewPool(size)
	assert.NoErrorf(t, err, "create TimingPool failed: %v", err)
	defer p.Release()

	for i := 0; i < size; i++ {
		j := i + 1
		_ = p.Submit(func() {
			<-ch
			d := j % 100
			time.Sleep(time.Duration(d) * time.Millisecond)
		})
	}
	assert.Equalf(t, size, p.Running(), "pool should be full, expected: %d, but got: %d", size, p.Running())

	close(ch)
	time.Sleep(5 * DefaultCleanIntervalTime)
	assert.Equalf(t, 0, p.Running(), "pool should be empty after purge, but got %d", p.Running())

	ch = make(chan struct{})
	f := func(i interface{}) {
		<-ch
		d := i.(int) % 100
		time.Sleep(time.Duration(d) * time.Millisecond)
	}

	p1, err := NewPoolWithFunc(size, f)
	assert.NoErrorf(t, err, "create TimingPoolWithFunc failed: %v", err)
	defer p1.Release()

	for i := 0; i < size; i++ {
		_ = p1.Invoke(i)
	}
	assert.Equalf(t, size, p1.Running(), "pool should be full, expected: %d, but got: %d", size, p1.Running())

	close(ch)
	time.Sleep(5 * DefaultCleanIntervalTime)
	assert.Equalf(t, 0, p1.Running(), "pool should be empty after purge, but got %d", p1.Running())
}

func TestPurgePreMallocPool(t *testing.T) {
	p, err := NewPool(10, WithPreAlloc(true))
	assert.NoErrorf(t, err, "create TimingPool failed: %v", err)
	defer p.Release()
	_ = p.Submit(demoFunc)
	time.Sleep(3 * DefaultCleanIntervalTime)
	assert.EqualValues(t, 0, p.Running(), "all p should be purged")
	p1, err := NewPoolWithFunc(10, demoPoolFunc)
	assert.NoErrorf(t, err, "create TimingPoolWithFunc failed: %v", err)
	defer p1.Release()
	_ = p1.Invoke(1)
	time.Sleep(3 * DefaultCleanIntervalTime)
	assert.EqualValues(t, 0, p.Running(), "all p should be purged")
}

func TestNonblockingSubmit(t *testing.T) {
	poolSize := 10
	p, err := NewPool(poolSize, WithNonblocking(true))
	assert.NoErrorf(t, err, "create TimingPool failed: %v", err)
	defer p.Release()
	for i := 0; i < poolSize-1; i++ {
		assert.NoError(t, p.Submit(longRunningFunc), "nonblocking submit when pool is not full shouldn't return error")
	}
	ch := make(chan struct{})
	ch1 := make(chan struct{})
	f := func() {
		<-ch
		close(ch1)
	}
	// p is full now.
	assert.NoError(t, p.Submit(f), "nonblocking submit when pool is not full shouldn't return error")
	assert.EqualError(t, p.Submit(demoFunc), ErrPoolOverload.Error(),
		"nonblocking submit when pool is full should get an ErrPoolOverload")
	// interrupt f to get an available worker
	close(ch)
	<-ch1
	assert.NoError(t, p.Submit(demoFunc), "nonblocking submit when pool is not full shouldn't return error")
}

func TestMaxBlockingSubmit(t *testing.T) {
	poolSize := 10
	p, err := NewPool(poolSize, WithMaxBlockingTasks(1))
	assert.NoErrorf(t, err, "create TimingPool failed: %v", err)
	defer p.Release()
	for i := 0; i < poolSize-1; i++ {
		assert.NoError(t, p.Submit(longRunningFunc), "submit when pool is not full shouldn't return error")
	}
	ch := make(chan struct{})
	f := func() {
		<-ch
	}
	// p is full now.
	assert.NoError(t, p.Submit(f), "submit when pool is not full shouldn't return error")
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
	assert.EqualError(t, p.Submit(demoFunc), ErrPoolOverload.Error(),
		"blocking submit when pool reach max blocking submit should return ErrPoolOverload")
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
	var wg sync.WaitGroup
	p, err := NewPoolWithFunc(poolSize, func(i interface{}) {
		longRunningPoolFunc(i)
		wg.Done()
	}, WithNonblocking(true))
	assert.NoError(t, err, "create TimingPool failed: %v", err)
	defer p.Release()
	ch := make(chan struct{})
	wg.Add(poolSize)
	for i := 0; i < poolSize-1; i++ {
		assert.NoError(t, p.Invoke(ch), "nonblocking submit when pool is not full shouldn't return error")
	}
	// p is full now.
	assert.NoError(t, p.Invoke(ch), "nonblocking submit when pool is not full shouldn't return error")
	assert.EqualError(t, p.Invoke(nil), ErrPoolOverload.Error(),
		"nonblocking submit when pool is full should get an ErrPoolOverload")
	// interrupt f to get an available worker
	close(ch)
	wg.Wait()
	assert.NoError(t, p.Invoke(nil), "nonblocking submit when pool is not full shouldn't return error")
}

func TestMaxBlockingSubmitWithFunc(t *testing.T) {
	poolSize := 10
	p, err := NewPoolWithFunc(poolSize, longRunningPoolFunc, WithMaxBlockingTasks(1))
	assert.NoError(t, err, "create TimingPool failed: %v", err)
	defer p.Release()
	for i := 0; i < poolSize-1; i++ {
		assert.NoError(t, p.Invoke(Param), "submit when pool is not full shouldn't return error")
	}
	ch := make(chan struct{})
	// p is full now.
	assert.NoError(t, p.Invoke(ch), "submit when pool is not full shouldn't return error")
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
	assert.EqualErrorf(t, p.Invoke(Param), ErrPoolOverload.Error(),
		"blocking submit when pool reach max blocking submit should return ErrPoolOverload: %v", err)
	// interrupt one func to make blocking submit successful.
	close(ch)
	wg.Wait()
	select {
	case <-errCh:
		t.Fatalf("blocking submit when pool is full should not return error")
	default:
	}
}

func TestRebootDefaultPool(t *testing.T) {
	defer Release()
	Reboot()
	var wg sync.WaitGroup
	wg.Add(1)
	_ = Submit(func() {
		demoFunc()
		wg.Done()
	})
	wg.Wait()
	Release()
	assert.EqualError(t, Submit(nil), ErrPoolClosed.Error(), "pool should be closed")
	Reboot()
	wg.Add(1)
	assert.NoError(t, Submit(func() { wg.Done() }), "pool should be rebooted")
	wg.Wait()
}

func TestRebootNewPool(t *testing.T) {
	var wg sync.WaitGroup
	p, err := NewPool(10)
	assert.NoErrorf(t, err, "create Pool failed: %v", err)
	defer p.Release()
	wg.Add(1)
	_ = p.Submit(func() {
		demoFunc()
		wg.Done()
	})
	wg.Wait()
	p.Release()
	assert.EqualError(t, p.Submit(nil), ErrPoolClosed.Error(), "pool should be closed")
	p.Reboot()
	wg.Add(1)
	assert.NoError(t, p.Submit(func() { wg.Done() }), "pool should be rebooted")
	wg.Wait()

	p1, err := NewPoolWithFunc(10, func(i interface{}) {
		demoPoolFunc(i)
		wg.Done()
	})
	assert.NoErrorf(t, err, "create TimingPoolWithFunc failed: %v", err)
	defer p1.Release()
	wg.Add(1)
	_ = p1.Invoke(1)
	wg.Wait()
	p1.Release()
	assert.EqualError(t, p1.Invoke(nil), ErrPoolClosed.Error(), "pool should be closed")
	p1.Reboot()
	wg.Add(1)
	assert.NoError(t, p1.Invoke(1), "pool should be rebooted")
	wg.Wait()
}

func TestInfinitePool(t *testing.T) {
	c := make(chan struct{})
	p, _ := NewPool(-1)
	_ = p.Submit(func() {
		_ = p.Submit(func() {
			<-c
		})
	})
	c <- struct{}{}
	if n := p.Running(); n != 2 {
		t.Errorf("expect 2 workers running, but got %d", n)
	}
	if n := p.Free(); n != -1 {
		t.Errorf("expect -1 of free workers by unlimited pool, but got %d", n)
	}
	p.Tune(10)
	if capacity := p.Cap(); capacity != -1 {
		t.Fatalf("expect capacity: -1 but got %d", capacity)
	}
	var err error
	_, err = NewPool(-1, WithPreAlloc(true))
	assert.EqualErrorf(t, err, ErrInvalidPreAllocSize.Error(), "")
}

func testPoolWithDisablePurge(t *testing.T, p *Pool, numWorker int, waitForPurge time.Duration) {
	sig := make(chan struct{})
	var wg1, wg2 sync.WaitGroup
	wg1.Add(numWorker)
	wg2.Add(numWorker)
	for i := 0; i < numWorker; i++ {
		_ = p.Submit(func() {
			wg1.Done()
			<-sig
			wg2.Done()
		})
	}
	wg1.Wait()

	runningCnt := p.Running()
	assert.EqualValuesf(t, numWorker, runningCnt, "expect %d workers running, but got %d", numWorker, runningCnt)
	freeCnt := p.Free()
	assert.EqualValuesf(t, 0, freeCnt, "expect %d free workers, but got %d", 0, freeCnt)

	// Finish all tasks and sleep for a while to wait for purging, since we've disabled purge mechanism,
	// we should see that all workers are still running after the sleep.
	close(sig)
	wg2.Wait()
	time.Sleep(waitForPurge + waitForPurge/2)

	runningCnt = p.Running()
	assert.EqualValuesf(t, numWorker, runningCnt, "expect %d workers running, but got %d", numWorker, runningCnt)
	freeCnt = p.Free()
	assert.EqualValuesf(t, 0, freeCnt, "expect %d free workers, but got %d", 0, freeCnt)

	err := p.ReleaseTimeout(waitForPurge + waitForPurge/2)
	assert.NoErrorf(t, err, "release pool failed: %v", err)

	runningCnt = p.Running()
	assert.EqualValuesf(t, 0, runningCnt, "expect %d workers running, but got %d", 0, runningCnt)
	freeCnt = p.Free()
	assert.EqualValuesf(t, numWorker, freeCnt, "expect %d free workers, but got %d", numWorker, freeCnt)
}

func TestWithDisablePurgePool(t *testing.T) {
	numWorker := 10
	p, _ := NewPool(numWorker, WithDisablePurge(true))
	testPoolWithDisablePurge(t, p, numWorker, DefaultCleanIntervalTime)
}

func TestWithDisablePurgeAndWithExpirationPool(t *testing.T) {
	numWorker := 10
	expiredDuration := time.Millisecond * 100
	p, _ := NewPool(numWorker, WithDisablePurge(true), WithExpiryDuration(expiredDuration))
	testPoolWithDisablePurge(t, p, numWorker, expiredDuration)
}

func testPoolFuncWithDisablePurge(t *testing.T, p *PoolWithFunc, numWorker int, wg1, wg2 *sync.WaitGroup, sig chan struct{}, waitForPurge time.Duration) {
	for i := 0; i < numWorker; i++ {
		_ = p.Invoke(i)
	}
	wg1.Wait()

	runningCnt := p.Running()
	assert.EqualValuesf(t, numWorker, runningCnt, "expect %d workers running, but got %d", numWorker, runningCnt)
	freeCnt := p.Free()
	assert.EqualValuesf(t, 0, freeCnt, "expect %d free workers, but got %d", 0, freeCnt)

	// Finish all tasks and sleep for a while to wait for purging, since we've disabled purge mechanism,
	// we should see that all workers are still running after the sleep.
	close(sig)
	wg2.Wait()
	time.Sleep(waitForPurge + waitForPurge/2)

	runningCnt = p.Running()
	assert.EqualValuesf(t, numWorker, runningCnt, "expect %d workers running, but got %d", numWorker, runningCnt)
	freeCnt = p.Free()
	assert.EqualValuesf(t, 0, freeCnt, "expect %d free workers, but got %d", 0, freeCnt)

	err := p.ReleaseTimeout(waitForPurge + waitForPurge/2)
	assert.NoErrorf(t, err, "release pool failed: %v", err)

	runningCnt = p.Running()
	assert.EqualValuesf(t, 0, runningCnt, "expect %d workers running, but got %d", 0, runningCnt)
	freeCnt = p.Free()
	assert.EqualValuesf(t, numWorker, freeCnt, "expect %d free workers, but got %d", numWorker, freeCnt)
}

func TestWithDisablePurgePoolFunc(t *testing.T) {
	numWorker := 10
	sig := make(chan struct{})
	var wg1, wg2 sync.WaitGroup
	wg1.Add(numWorker)
	wg2.Add(numWorker)
	p, _ := NewPoolWithFunc(numWorker, func(i interface{}) {
		wg1.Done()
		<-sig
		wg2.Done()
	}, WithDisablePurge(true))
	testPoolFuncWithDisablePurge(t, p, numWorker, &wg1, &wg2, sig, DefaultCleanIntervalTime)
}

func TestWithDisablePurgeAndWithExpirationPoolFunc(t *testing.T) {
	numWorker := 2
	sig := make(chan struct{})
	var wg1, wg2 sync.WaitGroup
	wg1.Add(numWorker)
	wg2.Add(numWorker)
	expiredDuration := time.Millisecond * 100
	p, _ := NewPoolWithFunc(numWorker, func(i interface{}) {
		wg1.Done()
		<-sig
		wg2.Done()
	}, WithDisablePurge(true), WithExpiryDuration(expiredDuration))
	testPoolFuncWithDisablePurge(t, p, numWorker, &wg1, &wg2, sig, expiredDuration)
}

func TestInfinitePoolWithFunc(t *testing.T) {
	c := make(chan struct{})
	p, _ := NewPoolWithFunc(-1, func(i interface{}) {
		demoPoolFunc(i)
		<-c
	})
	_ = p.Invoke(10)
	_ = p.Invoke(10)
	c <- struct{}{}
	c <- struct{}{}
	if n := p.Running(); n != 2 {
		t.Errorf("expect 2 workers running, but got %d", n)
	}
	if n := p.Free(); n != -1 {
		t.Errorf("expect -1 of free workers by unlimited pool, but got %d", n)
	}
	p.Tune(10)
	if capacity := p.Cap(); capacity != -1 {
		t.Fatalf("expect capacity: -1 but got %d", capacity)
	}
	var err error
	_, err = NewPoolWithFunc(-1, demoPoolFunc, WithPreAlloc(true))
	if err != ErrInvalidPreAllocSize {
		t.Errorf("expect ErrInvalidPreAllocSize but got %v", err)
	}
}

func TestReleaseWhenRunningPool(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := NewPool(1)
	wg.Add(2)
	go func() {
		t.Log("start aaa")
		defer func() {
			wg.Done()
			t.Log("stop aaa")
		}()
		for i := 0; i < 30; i++ {
			j := i
			_ = p.Submit(func() {
				t.Log("do task", j)
				time.Sleep(1 * time.Second)
			})
		}
	}()

	go func() {
		t.Log("start bbb")
		defer func() {
			wg.Done()
			t.Log("stop bbb")
		}()
		for i := 100; i < 130; i++ {
			j := i
			_ = p.Submit(func() {
				t.Log("do task", j)
				time.Sleep(1 * time.Second)
			})
		}
	}()

	time.Sleep(3 * time.Second)
	p.Release()
	t.Log("wait for all goroutines to exit...")
	wg.Wait()
}

func TestReleaseWhenRunningPoolWithFunc(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := NewPoolWithFunc(1, func(i interface{}) {
		t.Log("do task", i)
		time.Sleep(1 * time.Second)
	})
	wg.Add(2)
	go func() {
		t.Log("start aaa")
		defer func() {
			wg.Done()
			t.Log("stop aaa")
		}()
		for i := 0; i < 30; i++ {
			_ = p.Invoke(i)
		}
	}()

	go func() {
		t.Log("start bbb")
		defer func() {
			wg.Done()
			t.Log("stop bbb")
		}()
		for i := 100; i < 130; i++ {
			_ = p.Invoke(i)
		}
	}()

	time.Sleep(3 * time.Second)
	p.Release()
	t.Log("wait for all goroutines to exit...")
	wg.Wait()
}

func TestRestCodeCoverage(t *testing.T) {
	_, err := NewPool(-1, WithExpiryDuration(-1))
	t.Log(err)
	_, err = NewPool(1, WithExpiryDuration(-1))
	t.Log(err)
	_, err = NewPoolWithFunc(-1, demoPoolFunc, WithExpiryDuration(-1))
	t.Log(err)
	_, err = NewPoolWithFunc(1, demoPoolFunc, WithExpiryDuration(-1))
	t.Log(err)

	options := Options{}
	options.ExpiryDuration = time.Duration(10) * time.Second
	options.Nonblocking = true
	options.PreAlloc = true
	poolOpts, _ := NewPool(1, WithOptions(options))
	t.Logf("Pool with options, capacity: %d", poolOpts.Cap())

	p0, _ := NewPool(TestSize, WithLogger(log.New(os.Stderr, "", log.LstdFlags)))
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

	pprem, _ := NewPool(TestSize, WithPreAlloc(true))
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

	p, _ := NewPoolWithFunc(TestSize, demoPoolFunc)
	defer func() {
		_ = p.Invoke(Param)
	}()
	defer p.Release()
	for i := 0; i < n; i++ {
		_ = p.Invoke(Param)
	}
	time.Sleep(DefaultCleanIntervalTime)
	t.Logf("pool with func, capacity:%d", p.Cap())
	t.Logf("pool with func, running workers number:%d", p.Running())
	t.Logf("pool with func, free workers number:%d", p.Free())
	p.Tune(TestSize)
	p.Tune(TestSize / 10)
	t.Logf("pool with func, after tuning capacity, capacity:%d, running:%d", p.Cap(), p.Running())

	ppremWithFunc, _ := NewPoolWithFunc(TestSize, demoPoolFunc, WithPreAlloc(true))
	defer func() {
		_ = ppremWithFunc.Invoke(Param)
	}()
	defer ppremWithFunc.Release()
	for i := 0; i < n; i++ {
		_ = ppremWithFunc.Invoke(Param)
	}
	time.Sleep(DefaultCleanIntervalTime)
	t.Logf("pre-malloc pool with func, capacity:%d", ppremWithFunc.Cap())
	t.Logf("pre-malloc pool with func, running workers number:%d", ppremWithFunc.Running())
	t.Logf("pre-malloc pool with func, free workers number:%d", ppremWithFunc.Free())
	ppremWithFunc.Tune(TestSize)
	ppremWithFunc.Tune(TestSize / 10)
	t.Logf("pre-malloc pool with func, after tuning capacity, capacity:%d, running:%d", ppremWithFunc.Cap(),
		ppremWithFunc.Running())
}

func TestPoolTuneScaleUp(t *testing.T) {
	c := make(chan struct{})
	p, _ := NewPool(2)
	for i := 0; i < 2; i++ {
		_ = p.Submit(func() {
			<-c
		})
	}
	if n := p.Running(); n != 2 {
		t.Errorf("expect 2 workers running, but got %d", n)
	}
	// test pool tune scale up one
	p.Tune(3)
	_ = p.Submit(func() {
		<-c
	})
	if n := p.Running(); n != 3 {
		t.Errorf("expect 3 workers running, but got %d", n)
	}
	// test pool tune scale up multiple
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.Submit(func() {
				<-c
			})
		}()
	}
	p.Tune(8)
	wg.Wait()
	if n := p.Running(); n != 8 {
		t.Errorf("expect 8 workers running, but got %d", n)
	}
	for i := 0; i < 8; i++ {
		c <- struct{}{}
	}
	p.Release()

	// test PoolWithFunc
	pf, _ := NewPoolWithFunc(2, func(i interface{}) {
		<-c
	})
	for i := 0; i < 2; i++ {
		_ = pf.Invoke(1)
	}
	if n := pf.Running(); n != 2 {
		t.Errorf("expect 2 workers running, but got %d", n)
	}
	// test pool tune scale up one
	pf.Tune(3)
	_ = pf.Invoke(1)
	if n := pf.Running(); n != 3 {
		t.Errorf("expect 3 workers running, but got %d", n)
	}
	// test pool tune scale up multiple
	var pfwg sync.WaitGroup
	for i := 0; i < 5; i++ {
		pfwg.Add(1)
		go func() {
			defer pfwg.Done()
			_ = pf.Invoke(1)
		}()
	}
	pf.Tune(8)
	pfwg.Wait()
	if n := pf.Running(); n != 8 {
		t.Errorf("expect 8 workers running, but got %d", n)
	}
	for i := 0; i < 8; i++ {
		c <- struct{}{}
	}
	close(c)
	pf.Release()
}

func TestReleaseTimeout(t *testing.T) {
	p, _ := NewPool(10)
	for i := 0; i < 5; i++ {
		_ = p.Submit(func() {
			time.Sleep(time.Second)
		})
	}
	assert.NotZero(t, p.Running())
	err := p.ReleaseTimeout(2 * time.Second)
	assert.NoError(t, err)

	var pf *PoolWithFunc
	pf, _ = NewPoolWithFunc(10, func(i interface{}) {
		dur := i.(time.Duration)
		time.Sleep(dur)
	})
	for i := 0; i < 5; i++ {
		_ = pf.Invoke(time.Second)
	}
	assert.NotZero(t, pf.Running())
	err = pf.ReleaseTimeout(2 * time.Second)
	assert.NoError(t, err)
}
