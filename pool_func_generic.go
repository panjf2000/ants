// MIT License

// Copyright (c) 2025 Andy Pan

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

// PoolWithFuncGeneric is the generic version of PoolWithFunc.
type PoolWithFuncGeneric[T any] struct {
	*poolCommon

	// fn is the unified function for processing tasks.
	fn func(T)
}

// Invoke passes the argument to the pool to start a new task.
func (p *PoolWithFuncGeneric[T]) Invoke(arg T) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}

	w, err := p.retrieveWorker()
	if w != nil {
		w.(*goWorkerWithFuncGeneric[T]).arg <- arg
	}
	return err
}

// NewPoolWithFuncGeneric instantiates a PoolWithFuncGeneric[T] with customized options.
func NewPoolWithFuncGeneric[T any](size int, pf func(T), options ...Option) (*PoolWithFuncGeneric[T], error) {
	if pf == nil {
		return nil, ErrLackPoolFunc
	}

	pc, err := newPool(size, options...)
	if err != nil {
		return nil, err
	}

	pool := &PoolWithFuncGeneric[T]{
		poolCommon: pc,
		fn:         pf,
	}

	pool.workerCache.New = func() any {
		return &goWorkerWithFuncGeneric[T]{
			pool: pool,
			arg:  make(chan T, workerChanCap),
			exit: make(chan struct{}, 1),
		}
	}

	return pool, nil
}
