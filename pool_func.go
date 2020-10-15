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
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2/internal"
)

// PoolWithFunc accepts the tasks from client,
// it limits the total of goroutines to a given number by recycling goroutines.
type PoolWithFunc struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// workers is a slice that store the available workers.
	workers []*goWorkerWithFunc

	// state is used to notice the pool to closed itself.
	state int32

	// lock for synchronous operation.
	lock sync.Locker

	// cond for waiting to get a idle worker.
	cond *sync.Cond

	// poolFunc is the function for processing tasks.
	poolFunc func(interface{})

	// workerCache speeds up the obtainment of the an usable worker in function:retrieveWorker.
	workerCache sync.Pool

	// blockingNum is the number of the goroutines already been blocked on pool.Submit, protected by pool.lock
	blockingNum int

	options *Options
}

// purgePeriodically clears expired workers periodically which runs in an individual goroutine, as a scavenger.
func (p *PoolWithFunc) purgePeriodically() {
	heartbeat := time.NewTicker(p.options.ExpiryDuration)
	defer heartbeat.Stop()

	var expiredWorkers []*goWorkerWithFunc
	for range heartbeat.C {
		if atomic.LoadInt32(&p.state) == CLOSED {
			break
		}
		currentTime := time.Now()
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers)
		var i int
		for i = 0; i < n && currentTime.Sub(idleWorkers[i].recycleTime) > p.options.ExpiryDuration; i++ {
		}
		expiredWorkers = append(expiredWorkers[:0], idleWorkers[:i]...)
		if i > 0 {
			m := copy(idleWorkers, idleWorkers[i:])
			for i = m; i < n; i++ {
				idleWorkers[i] = nil
			}
			p.workers = idleWorkers[:m]
		}
		p.lock.Unlock()

		// Notify obsolete workers to stop.
		// This notification must be outside the p.lock, since w.task
		// may be blocking and may consume a lot of time if many workers
		// are located on non-local CPUs.
		for i, w := range expiredWorkers {
			w.args <- nil
			expiredWorkers[i] = nil
		}

		// There might be a situation that all workers have been cleaned up(no any worker is running)
		// while some invokers still get stuck in "p.cond.Wait()",
		// then it ought to wakes all those invokers.
		if p.Running() == 0 {
			p.cond.Broadcast()
		}
	}
}

// NewPoolWithFunc generates an instance of ants pool with a specific function.
func NewPoolWithFunc(size int, pf func(interface{}), options ...Option) (*PoolWithFunc, error) {
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}

	if pf == nil {
		return nil, ErrLackPoolFunc
	}

	opts := loadOptions(options...)

	if expiry := opts.ExpiryDuration; expiry < 0 {
		return nil, ErrInvalidPoolExpiry
	} else if expiry == 0 {
		opts.ExpiryDuration = DefaultCleanIntervalTime
	}

	if opts.Logger == nil {
		opts.Logger = defaultLogger
	}

	p := &PoolWithFunc{
		capacity: int32(size),
		poolFunc: pf,
		lock:     internal.NewSpinLock(),
		options:  opts,
	}
	p.workerCache.New = func() interface{} {
		return &goWorkerWithFunc{
			pool: p,
			args: make(chan interface{}, workerChanCap),
		}
	}
	if p.options.PreAlloc {
		p.workers = make([]*goWorkerWithFunc, 0, size)
	}
	p.cond = sync.NewCond(p.lock)

	// Start a goroutine to clean up expired workers periodically.
	go p.purgePeriodically()

	return p, nil
}

//---------------------------------------------------------------------------

// Invoke submits a task to pool.
func (p *PoolWithFunc) Invoke(args interface{}) error {
	if atomic.LoadInt32(&p.state) == CLOSED {
		return ErrPoolClosed
	}
	var w *goWorkerWithFunc
	if w = p.retrieveWorker(); w == nil {
		return ErrPoolOverload
	}
	w.args <- args
	return nil
}

// Running returns the number of the currently running goroutines.
func (p *PoolWithFunc) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns a available goroutines to work.
func (p *PoolWithFunc) Free() int {
	return p.Cap() - p.Running()
}

// Cap returns the capacity of this pool.
func (p *PoolWithFunc) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Tune changes the capacity of this pool.
func (p *PoolWithFunc) Tune(size int) {
	if size <= 0 || size == p.Cap() || p.options.PreAlloc {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
}

// Release Closes this pool.
func (p *PoolWithFunc) Release() {
	atomic.StoreInt32(&p.state, CLOSED)
	p.lock.Lock()
	idleWorkers := p.workers
	for _, w := range idleWorkers {
		w.args <- nil
	}
	p.workers = nil
	p.lock.Unlock()
}

// Reboot reboots a released pool.
func (p *PoolWithFunc) Reboot() {
	if atomic.CompareAndSwapInt32(&p.state, CLOSED, OPENED) {
		go p.purgePeriodically()
	}
}

//---------------------------------------------------------------------------

// incRunning increases the number of the currently running goroutines.
func (p *PoolWithFunc) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *PoolWithFunc) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// retrieveWorker returns a available worker to run the tasks.
func (p *PoolWithFunc) retrieveWorker() (w *goWorkerWithFunc) {
	spawnWorker := func() {
		w = p.workerCache.Get().(*goWorkerWithFunc)
		w.run()
	}

	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
	} else if p.Running() < p.Cap() {
		p.lock.Unlock()
		spawnWorker()
	} else {
		if p.options.Nonblocking {
			p.lock.Unlock()
			return
		}
	Reentry:
		if p.options.MaxBlockingTasks != 0 && p.blockingNum >= p.options.MaxBlockingTasks {
			p.lock.Unlock()
			return
		}
		p.blockingNum++
		p.cond.Wait()
		p.blockingNum--
		if p.Running() == 0 {
			p.lock.Unlock()
			spawnWorker()
			return
		}
		l := len(p.workers) - 1
		if l < 0 {
			goto Reentry
		}
		w = p.workers[l]
		p.workers[l] = nil
		p.workers = p.workers[:l]
		p.lock.Unlock()
	}
	return
}

// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *PoolWithFunc) revertWorker(worker *goWorkerWithFunc) bool {
	if atomic.LoadInt32(&p.state) == CLOSED || p.Running() > p.Cap() {
		return false
	}
	worker.recycleTime = time.Now()
	p.lock.Lock()

	// To avoid memory leaks, add a double check in the lock scope.
	// Issue: https://github.com/panjf2000/ants/issues/113
	if atomic.LoadInt32(&p.state) == CLOSED {
		p.lock.Unlock()
		return false
	}

	p.workers = append(p.workers, worker)

	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()
	return true
}
