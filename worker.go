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
				atomic.AddInt32(&w.pool.running, -1)
			case <-w.exit:
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
