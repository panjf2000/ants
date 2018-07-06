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
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type pf func(interface{}) error

// PoolWithFunc accept the tasks from client,it limits the total
// of goroutines to a given number by recycling goroutines.
type PoolWithFunc struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// expiryDuration set the expired time (second) of every worker.
	expiryDuration time.Duration

	// freeSignal is used to notice pool there are available
	// workers which can be sent to work.
	freeSignal chan sig

	// workers is a slice that store the available workers.
	workers []*WorkerWithFunc

	// release is used to notice the pool to closed itself.
	release chan sig

	// lock for synchronous operation
	lock sync.Mutex

	// pf is the function for processing tasks
	poolFunc pf

	once sync.Once
}

func (p *PoolWithFunc) monitorAndClear() {
	go func() {
		for {
			time.Sleep(p.expiryDuration)
			currentTime := time.Now()
			p.lock.Lock()
			idleWorkers := p.workers
			n := 0
			for i, w := range idleWorkers {
				if currentTime.Sub(w.recycleTime) <= p.expiryDuration {
					break
				}
				n = i
				w.stop()
				idleWorkers[i] = nil
				p.running--
			}
			if n > 0 {
				n += 1
				p.workers = idleWorkers[n:]
			}
			p.lock.Unlock()
		}
	}()
}

// NewPoolWithFunc generates a instance of ants pool with a specific function.
func NewPoolWithFunc(size, expiry int, f pf) (*PoolWithFunc, error) {
	if size <= 0 {
		return nil, ErrPoolSizeInvalid
	}
	p := &PoolWithFunc{
		capacity:       int32(size),
		freeSignal:     make(chan sig, math.MaxInt32),
		release:        make(chan sig, 1),
		expiryDuration: time.Duration(expiry) * time.Second,
		poolFunc:       f,
	}
	p.MonitorAndClear()
	return p, nil
}

//-------------------------------------------------------------------------

// Serve submit a task to pool
func (p *PoolWithFunc) Serve(args interface{}) error {
	//if atomic.LoadInt32(&p.closed) == 1 {
	//	return ErrPoolClosed
	//}
	if len(p.release) > 0 {
		return ErrPoolClosed
	}
	w := p.getWorker()
	w.sendTask(args)
	return nil
}

// Running returns the number of the currently running goroutines
func (p *PoolWithFunc) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the available goroutines to work
func (p *PoolWithFunc) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

// Cap returns the capacity of this pool
func (p *PoolWithFunc) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Release Closed this pool
func (p *PoolWithFunc) Release() error {
	p.once.Do(func() {
		p.release <- sig{}
		running := p.Running()
		for i := 0; i < running; i++ {
			p.getWorker().stop()
		}
		for i := range p.workers {
			p.workers[i] = nil
		}
	})
	return nil
}

// ReSize change the capacity of this pool
func (p *PoolWithFunc) ReSize(size int) {
	if size < p.Cap() {
		diff := p.Cap() - size
		for i := 0; i < diff; i++ {
			p.getWorker().stop()
		}
	} else if size == p.Cap() {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
}

//-------------------------------------------------------------------------

// getWorker returns a available worker to run the tasks.
func (p *PoolWithFunc) getWorker() *WorkerWithFunc {
	var w *WorkerWithFunc
	waiting := false

	p.lock.Lock()
	workers := p.workers
	n := len(workers) - 1
	if n < 0 {
		if p.running >= p.capacity {
			waiting = true
		} else {
			p.running++
		}
	} else {
		<-p.freeSignal
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
	}
	p.lock.Unlock()

	if waiting {
		<-p.freeSignal
		p.lock.Lock()
		workers = p.workers
		l := len(workers) - 1
		w = workers[l]
		workers[l] = nil
		p.workers = workers[:l]
		p.lock.Unlock()
	} else if w == nil {
		w = &WorkerWithFunc{
			pool: p,
			args: make(chan interface{}),
		}
		w.run()
	}
	return w
}

// putWorker puts a worker back into free pool, recycling the goroutines.
func (p *PoolWithFunc) putWorker(worker *WorkerWithFunc) {
	worker.recycleTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	p.lock.Unlock()
	p.freeSignal <- sig{}
}
