package ants

import (
	"runtime"
	"sync/atomic"
	"sync"
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
	m          sync.Mutex
}

func NewPool(size int) *Pool {
	p := &Pool{
		capacity:   int32(size),
		freeSignal: make(chan sig, size),
		destroy:    make(chan sig, runtime.GOMAXPROCS(-1)),
	}

	return p
}

//-------------------------------------------------------------------------
//func (p *Pool) loop() {
//	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
//		go func() {
//			for {
//				select {
//				case <-p.launchSignal:
//					p.getWorker().sendTask(p.tasks.pop().(f))
//				case <-p.destroy:
//					return
//				}
//			}
//		}()
//	}
//}

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
	p.m.Lock()
	defer p.m.Unlock()
	for i := 0; i < runtime.GOMAXPROCS(-1)+1; i++ {
		p.destroy <- sig{}
	}
	return nil
}

//-------------------------------------------------------------------------

func (p *Pool) reachLimit() bool {
	return p.Running() >= p.Cap()
}

func (p *Pool) newWorker() *Worker {
	var w *Worker
	if p.reachLimit() {
		<-p.freeSignal
		return p.getWorker()
	}
	wp := p.workerPool.Get()
	if wp == nil {
		w = &Worker{
			pool: p,
			task: make(chan f),
		}
	} else {
		w = wp.(*Worker)
	}
	w.run()
	atomic.AddInt32(&p.running, 1)
	return w
}

func (p *Pool) getWorker() *Worker {
	var w *Worker
	p.m.Lock()
	workers := p.workers
	n := len(workers) - 1
	if n < 0 {
		p.m.Unlock()
		return p.newWorker()
	} else {
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
		atomic.AddInt32(&p.running, 1)
	}
	p.m.Unlock()
	return w
}

func (p *Pool) putWorker(worker *Worker) {
	p.workerPool.Put(worker)
	p.m.Lock()
	p.workers = append(p.workers, worker)
	if p.reachLimit() {
		p.freeSignal <- sig{}
	}
	p.m.Unlock()
}
