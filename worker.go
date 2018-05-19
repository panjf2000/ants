package ants

import "sync/atomic"

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
				w.pool.workers <- w
			case <-w.exit:
				atomic.AddInt32(&w.pool.length, -1)
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
