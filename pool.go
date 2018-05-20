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
	p.lock.Lock()
	defer p.lock.Unlock()
	for i := 0; i < runtime.GOMAXPROCS(-1)+1; i++ {
		p.destroy <- sig{}
	}
	return nil
}

//-------------------------------------------------------------------------

func (p *Pool) reachLimit() bool {
	return p.Running() >= p.Cap()
}

//func (p *Pool) newWorker() *Worker {
//	var w *Worker
//	if p.reachLimit() {
//		<-p.freeSignal
//		return p.getWorker()
//	}
//	wp := p.workerPool.Get()
//	if wp == nil {
//		w = &Worker{
//			pool: p,
//			task: make(chan f),
//		}
//	} else {
//		w = wp.(*Worker)
//	}
//	w.run()
//	atomic.AddInt32(&p.running, 1)
//	return w
//}
//
//func (p *Pool) getWorker() *Worker {
//	var w *Worker
//	p.lock.Lock()
//	workers := p.workers
//	n := len(workers) - 1
//	if n < 0 {
//		p.lock.Unlock()
//		return p.newWorker()
//	} else {
//		w = workers[n]
//		workers[n] = nil
//		p.workers = workers[:n]
//		//atomic.AddInt32(&p.running, 1)
//	}
//	p.lock.Unlock()
//	return w
//}

//func (p *Pool) newWorker() *Worker {
//	var w *Worker
//	if p.reachLimit() {
//		<-p.freeSignal
//		return p.getWorker()
//	}
//	wp := p.workerPool.Get()
//	if wp == nil {
//		w = &Worker{
//			pool: p,
//			task: make(chan f),
//		}
//	} else {
//		w = wp.(*Worker)
//	}
//	w.run()
//	atomic.AddInt32(&p.running, 1)
//	return w
//}

func (p *Pool) getWorker() *Worker {
	//fmt.Printf("init running workers number:%d\n", p.running)
	var w *Worker
	waiting := false

	p.lock.Lock()
	workers := p.workers
	n := len(workers) - 1
	if n < 0 {
		//fmt.Printf("running workers number:%d\n", p.running)
		if p.running >= p.capacity {
			waiting = true
		}
	} else {
		w = workers[n]
		workers[n] = nil
		p.workers = workers[:n]
		//atomic.AddInt32(&p.running, 1)
	}
	p.lock.Unlock()

	if waiting {
		<-p.freeSignal
		//p.lock.Lock()
		//fmt.Println("wait for a worker")
		//fmt.Println("get for a worker")
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
		//p.lock.Unlock()
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
	//fmt.Printf("put a worker, running worker number:%d\n", p.Running())
}
