// Copyright (c) 2018 Andy Pan
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package ants

import (
	"runtime"
	"sync/atomic"
	"sync"
	"math"
)

type sig struct{}

type f func()

type Pool struct {
	capacity   int32
	running    int32
	freeSignal chan sig
	workers    []*Worker
	workerPool sync.Pool
	destroy    chan sig
	lock       sync.Mutex
}

func NewPool(size int) *Pool {
	p := &Pool{
		capacity:   int32(size),
		freeSignal: make(chan sig, math.MaxInt32),
		destroy:    make(chan sig, runtime.GOMAXPROCS(-1)),
	}

	return p
}

//-------------------------------------------------------------------------

func (p *Pool) Push(task f) error {
	if len(p.destroy) > 0 {
		return nil
	}
	w := p.getWorker()
	w.sendTask(task)
	return nil
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *Pool) Destroy() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	for i := 0; i < runtime.GOMAXPROCS(-1)+1; i++ {
		p.destroy <- sig{}
	}
	return nil
}

//-------------------------------------------------------------------------

func (p *Pool) getWorker() *Worker {
	var w *Worker
	waiting := false

	p.lock.Lock()
	workers := p.workers
	n := len(workers) - 1
	if n < 0 {
		if p.running >= p.capacity {
			waiting = true
		}
	} else {
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
	}
	p.lock.Unlock()

	if waiting {
		<-p.freeSignal
		for {
			p.lock.Lock()
			workers = p.workers
			l := len(workers) - 1
			if l < 0 {
				p.lock.Unlock()
				continue
			}
			w = workers[l]
			workers[l] = nil
			p.workers = workers[:l]
			p.lock.Unlock()
			break
		}
	} else {
		wp := p.workerPool.Get()
		if wp == nil {
			w = &Worker{
				pool: p,
				task: make(chan f),
			}
			w.run()
			atomic.AddInt32(&p.running, 1)
		} else {
			w = wp.(*Worker)
		}
	}
	return w
}

func (p *Pool) putWorker(worker *Worker) {
	p.workerPool.Put(worker)
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	p.lock.Unlock()
	p.freeSignal <- sig{}
}
