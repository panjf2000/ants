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

// Package ants implements an efficient and reliable goroutine pool for Go.
//
// With ants, Go applications are able to limit the number of active goroutines,
// recycle goroutines efficiently, and reduce the memory footprint significantly.
// Package ants is extremely useful in the scenarios where a massive number of
// goroutines are created and destroyed frequently, such as highly-concurrent
// batch processing systems, HTTP servers, services of asynchronous tasks, etc.
package ants

import (
	"context"
	"errors"
	"log"
	"math"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	syncx "github.com/panjf2000/ants/v2/pkg/sync"
)

const (
	// DefaultAntsPoolSize is the default capacity for a default goroutine pool.
	DefaultAntsPoolSize = math.MaxInt32

	// DefaultCleanIntervalTime is the interval time to clean up goroutines.
	DefaultCleanIntervalTime = time.Second
)

const (
	// OPENED represents that the pool is opened.
	OPENED = iota

	// CLOSED represents that the pool is closed.
	CLOSED
)

var (
	// ErrLackPoolFunc will be returned when invokers don't provide function for pool.
	ErrLackPoolFunc = errors.New("must provide function for pool")

	// ErrInvalidPoolExpiry will be returned when setting a negative number as the periodic duration to purge goroutines.
	ErrInvalidPoolExpiry = errors.New("invalid expiry for pool")

	// ErrPoolClosed will be returned when submitting task to a closed pool.
	ErrPoolClosed = errors.New("this pool has been closed")

	// ErrPoolOverload will be returned when the pool is full and no workers available.
	ErrPoolOverload = errors.New("too many goroutines blocked on submit or Nonblocking is set")

	// ErrInvalidPreAllocSize will be returned when trying to set up a negative capacity under PreAlloc mode.
	ErrInvalidPreAllocSize = errors.New("can not set up a negative capacity under PreAlloc mode")

	// ErrTimeout will be returned after the operations timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrInvalidPoolIndex will be returned when trying to retrieve a pool with an invalid index.
	ErrInvalidPoolIndex = errors.New("invalid pool index")

	// ErrInvalidLoadBalancingStrategy will be returned when trying to create a MultiPool with an invalid load-balancing strategy.
	ErrInvalidLoadBalancingStrategy = errors.New("invalid load-balancing strategy")

	// ErrInvalidMultiPoolSize  will be returned when trying to create a MultiPool with an invalid size.
	ErrInvalidMultiPoolSize = errors.New("invalid size for multiple pool")

	// workerChanCap determines whether the channel of a worker should be a buffered channel
	// to get the best performance. Inspired by fasthttp at
	// https://github.com/valyala/fasthttp/blob/master/workerpool.go#L139
	workerChanCap = func() int {
		// Use blocking channel if GOMAXPROCS=1.
		// This switches context from sender to receiver immediately,
		// which results in higher performance (under go1.5 at least).
		if runtime.GOMAXPROCS(0) == 1 {
			return 0
		}

		// Use non-blocking workerChan if GOMAXPROCS>1,
		// since otherwise the sender might be dragged down if the receiver is CPU-bound.
		return 1
	}()

	defaultLogger = Logger(log.New(os.Stderr, "[ants]: ", log.LstdFlags|log.Lmsgprefix|log.Lmicroseconds))

	// Init an instance pool when importing ants.
	defaultAntsPool, _ = NewPool(DefaultAntsPoolSize)
)

// Submit submits a task to pool.
func Submit(task func()) error {
	return defaultAntsPool.Submit(task)
}

// Running returns the number of the currently running goroutines.
func Running() int {
	return defaultAntsPool.Running()
}

// Cap returns the capacity of this default pool.
func Cap() int {
	return defaultAntsPool.Cap()
}

// Free returns the available goroutines to work.
func Free() int {
	return defaultAntsPool.Free()
}

// Release Closes the default pool.
func Release() {
	defaultAntsPool.Release()
}

// ReleaseTimeout is like Release but with a timeout, it waits all workers to exit before timing out.
func ReleaseTimeout(timeout time.Duration) error {
	return defaultAntsPool.ReleaseTimeout(timeout)
}

// Reboot reboots the default pool.
func Reboot() {
	defaultAntsPool.Reboot()
}

// Logger is used for logging formatted messages.
type Logger interface {
	// Printf must have the same semantics as log.Printf.
	Printf(format string, args ...any)
}

// poolCommon contains all common fields for other sophisticated pools.
type poolCommon struct {
	// capacity of the pool, a negative value means that the capacity of pool is limitless, an infinite pool is used to
	// avoid potential issue of endless blocking caused by nested usage of a pool: submitting a task to pool
	// which submits a new task to the same pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// lock for protecting the worker queue.
	lock sync.Locker

	// workers is a slice that store the available workers.
	workers workerQueue

	// state is used to notice the pool to closed itself.
	state int32

	// cond for waiting to get an idle worker.
	cond *sync.Cond

	// done is used to indicate that all workers are done.
	allDone chan struct{}
	// once is used to make sure the pool is closed just once.
	once *sync.Once

	// workerCache speeds up the obtainment of a usable worker in function:retrieveWorker.
	workerCache sync.Pool

	// waiting is the number of goroutines already been blocked on pool.Submit(), protected by pool.lock
	waiting int32

	purgeDone int32
	purgeCtx  context.Context
	stopPurge context.CancelFunc

	ticktockDone int32
	ticktockCtx  context.Context
	stopTicktock context.CancelFunc

	now atomic.Value

	options *Options
}

func newPool(size int, options ...Option) (*poolCommon, error) {
	if size <= 0 {
		size = -1
	}

	opts := loadOptions(options...)

	if !opts.DisablePurge {
		if expiry := opts.ExpiryDuration; expiry < 0 {
			return nil, ErrInvalidPoolExpiry
		} else if expiry == 0 {
			opts.ExpiryDuration = DefaultCleanIntervalTime
		}
	}

	if opts.Logger == nil {
		opts.Logger = defaultLogger
	}

	p := &poolCommon{
		capacity: int32(size),
		allDone:  make(chan struct{}),
		lock:     syncx.NewSpinLock(),
		once:     &sync.Once{},
		options:  opts,
	}
	if p.options.PreAlloc {
		if size == -1 {
			return nil, ErrInvalidPreAllocSize
		}
		p.workers = newWorkerQueue(queueTypeLoopQueue, size)
	} else {
		p.workers = newWorkerQueue(queueTypeStack, 0)
	}

	p.cond = sync.NewCond(p.lock)

	p.goPurge()
	p.goTicktock()

	return p, nil
}

// purgeStaleWorkers clears stale workers periodically, it runs in an individual goroutine, as a scavenger.
func (p *poolCommon) purgeStaleWorkers() {
	ticker := time.NewTicker(p.options.ExpiryDuration)

	defer func() {
		ticker.Stop()
		atomic.StoreInt32(&p.purgeDone, 1)
	}()

	purgeCtx := p.purgeCtx // copy to the local variable to avoid race from Reboot()
	for {
		select {
		case <-purgeCtx.Done():
			return
		case <-ticker.C:
		}

		if p.IsClosed() {
			break
		}

		var isDormant bool
		p.lock.Lock()
		staleWorkers := p.workers.refresh(p.options.ExpiryDuration)
		n := p.Running()
		isDormant = n == 0 || n == len(staleWorkers)
		p.lock.Unlock()

		// Clean up the stale workers.
		for i := range staleWorkers {
			staleWorkers[i].finish()
			staleWorkers[i] = nil
		}

		// There might be a situation where all workers have been cleaned up (no worker is running),
		// while some invokers still are stuck in p.cond.Wait(), then we need to awake those invokers.
		if isDormant && p.Waiting() > 0 {
			p.cond.Broadcast()
		}
	}
}

const nowTimeUpdateInterval = 500 * time.Millisecond

// ticktock is a goroutine that updates the current time in the pool regularly.
func (p *poolCommon) ticktock() {
	ticker := time.NewTicker(nowTimeUpdateInterval)
	defer func() {
		ticker.Stop()
		atomic.StoreInt32(&p.ticktockDone, 1)
	}()

	ticktockCtx := p.ticktockCtx // copy to the local variable to avoid race from Reboot()
	for {
		select {
		case <-ticktockCtx.Done():
			return
		case <-ticker.C:
		}

		if p.IsClosed() {
			break
		}

		p.now.Store(time.Now())
	}
}

func (p *poolCommon) goPurge() {
	if p.options.DisablePurge {
		return
	}

	// Start a goroutine to clean up expired workers periodically.
	p.purgeCtx, p.stopPurge = context.WithCancel(context.Background())
	go p.purgeStaleWorkers()
}

func (p *poolCommon) goTicktock() {
	p.now.Store(time.Now())
	p.ticktockCtx, p.stopTicktock = context.WithCancel(context.Background())
	go p.ticktock()
}

func (p *poolCommon) nowTime() time.Time {
	return p.now.Load().(time.Time)
}

// Running returns the number of workers currently running.
func (p *poolCommon) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the number of available workers, -1 indicates this pool is unlimited.
func (p *poolCommon) Free() int {
	c := p.Cap()
	if c < 0 {
		return -1
	}
	return c - p.Running()
}

// Waiting returns the number of tasks waiting to be executed.
func (p *poolCommon) Waiting() int {
	return int(atomic.LoadInt32(&p.waiting))
}

// Cap returns the capacity of this pool.
func (p *poolCommon) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Tune changes the capacity of this pool, note that it is noneffective to the infinite or pre-allocation pool.
func (p *poolCommon) Tune(size int) {
	capacity := p.Cap()
	if capacity == -1 || size <= 0 || size == capacity || p.options.PreAlloc {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
	if size > capacity {
		if size-capacity == 1 {
			p.cond.Signal()
			return
		}
		p.cond.Broadcast()
	}
}

// IsClosed indicates whether the pool is closed.
func (p *poolCommon) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == CLOSED
}

// Release closes this pool and releases the worker queue.
func (p *poolCommon) Release() {
	if !atomic.CompareAndSwapInt32(&p.state, OPENED, CLOSED) {
		return
	}

	if p.stopPurge != nil {
		p.stopPurge()
		p.stopPurge = nil
	}
	if p.stopTicktock != nil {
		p.stopTicktock()
		p.stopTicktock = nil
	}

	p.lock.Lock()
	p.workers.reset()
	p.lock.Unlock()

	// There might be some callers waiting in retrieveWorker(), so we need to wake them up to prevent
	// those callers blocking infinitely.
	p.cond.Broadcast()
}

// ReleaseTimeout is like Release but with a timeout, it waits all workers to exit before timing out.
func (p *poolCommon) ReleaseTimeout(timeout time.Duration) error {
	if p.IsClosed() || (!p.options.DisablePurge && p.stopPurge == nil) || p.stopTicktock == nil {
		return ErrPoolClosed
	}

	p.Release()

	var purgeCh <-chan struct{}
	if !p.options.DisablePurge {
		purgeCh = p.purgeCtx.Done()
	} else {
		purgeCh = p.allDone
	}

	if p.Running() == 0 {
		p.once.Do(func() {
			close(p.allDone)
		})
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return ErrTimeout
		case <-p.allDone:
			<-purgeCh
			<-p.ticktockCtx.Done()
			if p.Running() == 0 &&
				(p.options.DisablePurge || atomic.LoadInt32(&p.purgeDone) == 1) &&
				atomic.LoadInt32(&p.ticktockDone) == 1 {
				return nil
			}
		}
	}
}

// Reboot reboots a closed pool, it does nothing if the pool is not closed.
// If you intend to reboot a closed pool, use ReleaseTimeout() instead of
// Release() to ensure that all workers are stopped and resource are released
// before rebooting, otherwise you may run into data race.
func (p *poolCommon) Reboot() {
	if atomic.CompareAndSwapInt32(&p.state, CLOSED, OPENED) {
		atomic.StoreInt32(&p.purgeDone, 0)
		p.goPurge()
		atomic.StoreInt32(&p.ticktockDone, 0)
		p.goTicktock()
		p.allDone = make(chan struct{})
		p.once = &sync.Once{}
	}
}

func (p *poolCommon) addRunning(delta int) int {
	return int(atomic.AddInt32(&p.running, int32(delta)))
}

func (p *poolCommon) addWaiting(delta int) {
	atomic.AddInt32(&p.waiting, int32(delta))
}

// retrieveWorker returns an available worker to run the tasks.
func (p *poolCommon) retrieveWorker() (w worker, err error) {
	p.lock.Lock()

retry:
	// First try to fetch the worker from the queue.
	if w = p.workers.detach(); w != nil {
		p.lock.Unlock()
		return
	}

	// If the worker queue is empty, and we don't run out of the pool capacity,
	// then just spawn a new worker goroutine.
	if capacity := p.Cap(); capacity == -1 || capacity > p.Running() {
		p.lock.Unlock()
		w = p.workerCache.Get().(worker)
		w.run()
		return
	}

	// Bail out early if it's in nonblocking mode or the number of pending callers reaches the maximum limit value.
	if p.options.Nonblocking || (p.options.MaxBlockingTasks != 0 && p.Waiting() >= p.options.MaxBlockingTasks) {
		p.lock.Unlock()
		return nil, ErrPoolOverload
	}

	// Otherwise, we'll have to keep them blocked and wait for at least one worker to be put back into pool.
	p.addWaiting(1)
	p.cond.Wait() // block and wait for an available worker
	p.addWaiting(-1)

	if p.IsClosed() {
		p.lock.Unlock()
		return nil, ErrPoolClosed
	}

	goto retry
}

// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *poolCommon) revertWorker(worker worker) bool {
	if capacity := p.Cap(); (capacity > 0 && p.Running() > capacity) || p.IsClosed() {
		p.cond.Broadcast()
		return false
	}

	worker.setLastUsedTime(p.nowTime())

	p.lock.Lock()
	// To avoid memory leaks, add a double check in the lock scope.
	// Issue: https://github.com/panjf2000/ants/issues/113
	if p.IsClosed() {
		p.lock.Unlock()
		return false
	}
	if err := p.workers.insert(worker); err != nil {
		p.lock.Unlock()
		return false
	}
	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()

	return true
}
