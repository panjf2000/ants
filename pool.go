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
	capacity     int32
	running      int32
	tasks        *ConcurrentQueue
	workers      *ConcurrentQueue
	freeSignal   chan sig
	launchSignal chan sig
	destroy      chan sig
	m            *sync.Mutex
	wg           *sync.WaitGroup
}

func NewPool(size int) *Pool {
	p := &Pool{
		capacity:     int32(size),
		tasks:        NewConcurrentQueue(),
		workers:      NewConcurrentQueue(),
		freeSignal:   make(chan sig, math.MaxInt32),
		launchSignal: make(chan sig, math.MaxInt32),
		destroy:      make(chan sig, runtime.GOMAXPROCS(-1)),
		wg:           &sync.WaitGroup{},
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
				case <-p.launchSignal:
					p.getWorker().sendTask(p.tasks.pop().(f))
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
	p.tasks.push(task)
	p.wg.Add(1)
	p.launchSignal <- sig{}
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

func (p *Pool) Wait() {
	p.wg.Wait()
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
	if p.reachLimit() {
		<-p.freeSignal
		return p.getWorker()
	}
	worker := &Worker{
		pool: p,
		task: make(chan f),
		exit: make(chan sig),
	}
	worker.run()
	atomic.AddInt32(&p.running, 1)
	return worker
}

func (p *Pool) getWorker() *Worker {
	if w := p.workers.pop(); w != nil {
		return w.(*Worker)
	}
	return p.newWorker()
}

func (p *Pool) putWorker(worker *Worker) {
	p.workers.push(worker)
	if p.reachLimit() {
		p.freeSignal <- sig{}
	}
}
