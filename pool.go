package ants

import (
	"runtime"
	"sync/atomic"
	"math"
	"sync"
)

type sig struct{}

type f func()

type Pool struct {
	capacity int32
	length   int32
	tasks    chan f
	workers  chan *Worker
	destroy  chan sig
	m        sync.Mutex
}

func NewPool(size int) *Pool {
	p := &Pool{
		capacity: int32(size),
		tasks:    make(chan f, math.MaxInt32),
		//workers:  &sync.Pool{New: func() interface{} { return &Worker{} }},
		workers: make(chan *Worker, size),
		destroy: make(chan sig, runtime.GOMAXPROCS(-1)),
	}
	p.loop()
	return p
}


//-------------------------------------------------------------------------

func (p *Pool) loop() {
	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		go func() {
			for {
				select {
				case task := <-p.tasks:
					p.getWorker().sendTask(task)
				case <-p.destroy:
					return
				}
			}
		}()
	}
}

func (p *Pool) Push(task f) error {
	if len(p.destroy) > 0 {
		return nil
	}
	p.tasks <- task
	return nil
}
func (p *Pool) Size() int32 {
	return atomic.LoadInt32(&p.length)
}

func (p *Pool) Cap() int32 {
	return atomic.LoadInt32(&p.capacity)
}

func (p *Pool) Destroy() error {
	p.m.Lock()
	defer p.m.Unlock()
	for i := 0; i < runtime.GOMAXPROCS(-1) + 1; i++ {
		p.destroy <- sig{}
	}
	return nil
}


//-------------------------------------------------------------------------

func (p *Pool) reachLimit() bool {
	return p.Size() >= p.Cap()
}

func (p *Pool) newWorker() *Worker {
	worker := &Worker{
		pool: p,
		task: make(chan f),
		exit: make(chan sig),
	}
	worker.run()
	atomic.AddInt32(&p.length, 1)
	return worker
}

func (p *Pool) getWorker() *Worker {
	var worker *Worker
	if p.reachLimit() {
		worker = <-p.workers
	}
	worker = p.newWorker()
	return worker
}

