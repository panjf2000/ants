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
	"sync/atomic"
)

// Worker is the actual executor who run the tasks,
// it will start a goroutine that accept tasks and
// perform function calls.
type WorkerWithFunc struct {
	// pool who owns this worker.
	pool *PoolWithFunc

	// args is a job should be done.
	args chan interface{}
}

// run will start a goroutine to repeat the process
// that perform the function calls.
func (w *WorkerWithFunc) run() {
	go func() {
		for args := range w.args {
			if args == nil {
				atomic.AddInt32(&w.pool.running, -1)
				return
			}
			w.pool.poolFunc(args)
			w.pool.putWorker(w)
		}
	}()
}

// stop this worker.
func (w *WorkerWithFunc) stop() {
	w.args <- nil
}

// sendTask send a task to this worker.
func (w *WorkerWithFunc) sendTask(args interface{}) {
	w.args <- args
}
