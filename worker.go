package ants

import (
	"sync/atomic"
	"container/list"
	"sync"
)

type Worker struct {
	pool *Pool
	task chan f
	exit chan sig
}

func (w *Worker) run() {
	go func() {
		for {
			select {
			case f := <-w.task:
				f()
				w.pool.workers.push(w)
				//w.pool.wg.Done()
			case <-w.exit:
				atomic.AddInt32(&w.pool.running, -1)
				return
			}
		}
	}()
}

func (w *Worker) stop() {
	w.exit <- sig{}
}

func (w *Worker) sendTask(task f) {
	w.task <- task
}

//--------------------------------------------------------------------------------

type ConcurrentQueue struct {
	queue *list.List
	m     sync.Mutex
}

func NewConcurrentQueue() *ConcurrentQueue {
	q := new(ConcurrentQueue)
	q.queue = list.New()
	return q
}

func (q *ConcurrentQueue) push(v interface{}) {
	defer q.m.Unlock()
	q.m.Lock()
	q.queue.PushFront(v)
}

func (q *ConcurrentQueue) pop() interface{} {
	defer q.m.Unlock()
	q.m.Lock()
	if elem := q.queue.Back(); elem != nil{
		return q.queue.Remove(elem)
	}
	return nil
}
