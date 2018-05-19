package ants

import (
	"sync/atomic"
)

type Worker struct {
	pool *Pool
	task chan f
}

func (w *Worker) run() {
	go func() {
		for f := range w.task {
			if f == nil {
				atomic.AddInt32(&w.pool.running, -1)
				return
			}
			f()
			w.pool.putWorker(w)
			w.pool.wg.Done()
		}
	}()
}

func (w *Worker) stop() {
	w.task <- nil
}

func (w *Worker) sendTask(task f) {
	w.task <- task
}
