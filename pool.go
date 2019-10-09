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

// Pool accept the tasks from client, it limits the total of goroutines to a given number by recycling goroutines.
type Pool struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// expiryDuration set the expired time (second) of every worker.
	expiryDuration time.Duration

	// workers is a slice that store the available workers.
	workers workerQueue

	// release is used to notice the pool to closed itself.
	release int32

	// lock for synchronous operation.
	lock sync.Locker

	// cond for waiting to get a idle worker.
	cond *sync.Cond

	// once makes sure releasing this pool will just be done for one time.
	once sync.Once

	// workerCache speeds up the obtainment of the an usable worker in function:retrieveWorker.
	workerCache sync.Pool

	// panicHandler is used to handle panics from each worker goroutine.
	// if nil, panics will be thrown out again from worker goroutines.
	panicHandler func(interface{})

	// Max number of goroutine blocking on pool.Submit.
	// 0 (default value) means no such limit.
	maxBlockingTasks int32

	// goroutine already been blocked on pool.Submit
	// protected by pool.lock
	blockingNum int32

	// When nonblocking is true, Pool.Submit will never be blocked.
	// ErrPoolOverload will be returned when Pool.Submit cannot be done at once.
	// When nonblocking is true, MaxBlockingTasks is inoperative.
	nonblocking bool
}

// Clear expired workers periodically.
func (p *Pool) periodicallyPurge() {
	heartbeat := time.NewTicker(p.expiryDuration)
	defer heartbeat.Stop()

	for range heartbeat.C {
		if atomic.LoadInt32(&p.release) == CLOSED {
			break
		}

		p.lock.Lock()
		stream := p.workers.releaseExpiry(p.expiryDuration)
		p.lock.Unlock()

		// Notify obsolete workers to stop.
		// This notification must be outside the p.lock, since w.task
		// may be blocking and may consume a lot of time if many workers
		// are located on non-local CPUs.
		for w := range stream {
			w.task <- nil
		}

		// There might be a situation that all workers have been cleaned up(no any worker is running)
		// while some invokers still get stuck in "p.cond.Wait()",
		// then it ought to wakes all those invokers.
		if p.Running() == 0 {
			p.cond.Broadcast()
		}
	}
}

// NewPool generates an instance of ants pool.
func NewPool(size int, options ...Option) (*Pool, error) {
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}

	opts := new(Options)
	for _, option := range options {
		option(opts)
	}

	if expiry := opts.ExpiryDuration; expiry < 0 {
		return nil, ErrInvalidPoolExpiry
	} else if expiry == 0 {
		opts.ExpiryDuration = DefaultCleanIntervalTime
	}

	p := &Pool{
		capacity:         int32(size),
		expiryDuration:   opts.ExpiryDuration,
		nonblocking:      opts.Nonblocking,
		maxBlockingTasks: int32(opts.MaxBlockingTasks),
		panicHandler:     opts.PanicHandler,
		lock:             internal.NewSpinLock(),
	}
	p.workerCache = sync.Pool{
		New: func() interface{} {
			return &goWorker{
				pool: p,
				task: make(chan func(), workerChanCap),
			}
		},
	}
	if opts.PreAlloc {
		p.workers = newQueue(loopQueueType, size)
	} else {
		p.workers = newQueue(stackType, 0)
	}

	p.cond = sync.NewCond(p.lock)

	// Start a goroutine to clean up expired workers periodically.
	go p.periodicallyPurge()

	return p, nil
}

// ---------------------------------------------------------------------------

// Submit submits a task to this pool.
func (p *Pool) Submit(task func()) error {
	if atomic.LoadInt32(&p.release) == CLOSED {
		return ErrPoolClosed
	}
	var w *goWorker
	if w = p.retrieveWorker(); w == nil {
		return ErrPoolOverload
	}
	w.task <- task
	return nil
}

// Running returns the number of the currently running goroutines.
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the available goroutines to work.
func (p *Pool) Free() int {
	return p.Cap() - p.Running()
}

// Cap returns the capacity of this pool.
func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Tune changes the capacity of this pool.
func (p *Pool) Tune(size int) {
	if p.Cap() == size {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
}

// Release Closes this pool.
func (p *Pool) Release() {
	p.once.Do(func() {
		atomic.StoreInt32(&p.release, 1)
		p.lock.Lock()
		p.workers.releaseAll()
		p.lock.Unlock()
	})
}

// ---------------------------------------------------------------------------

// incRunning increases the number of the currently running goroutines.
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// retrieveWorker returns a available worker to run the tasks.
func (p *Pool) retrieveWorker() *goWorker {
	var w *goWorker
	spawnWorker := func() {
		w = p.workerCache.Get().(*goWorker)
		w.run()
	}

	p.lock.Lock()

	w = p.workers.dequeue()
	if w != nil {
		p.lock.Unlock()
	} else if p.Running() < p.Cap() {
		p.lock.Unlock()
		spawnWorker()
	} else {
		if p.nonblocking {
			p.lock.Unlock()
			return nil
		}
	Reentry:
		if p.maxBlockingTasks != 0 && p.blockingNum >= p.maxBlockingTasks {
			p.lock.Unlock()
			return nil
		}
		p.blockingNum++
		p.cond.Wait()
		p.blockingNum--
		if p.Running() == 0 {
			p.lock.Unlock()
			spawnWorker()
			return w
		}

		w = p.workers.dequeue()
		if w == nil {
			goto Reentry
		}

		p.lock.Unlock()
	}
	return w
}

// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *Pool) revertWorker(worker *goWorker) bool {
	if atomic.LoadInt32(&p.release) == CLOSED || p.Running() > p.Cap() {
		return false
	}
	worker.recycleTime = time.Now()
	p.lock.Lock()

	err := p.workers.enqueue(worker)
	if err != nil {
		return false
	}

	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()
	return true
}
