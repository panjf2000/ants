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

// PoolWithFunc is like Pool but accepts a unified function for all goroutines to execute.
type PoolWithFunc struct {
	*poolCommon

	// poolFunc is the unified function for processing tasks.
	poolFunc func(any)
}

// Invoke passes arguments to the pool.
//
// Note that you are allowed to call Pool.Invoke() from the current Pool.Invoke(),
// but what calls for special attention is that you will get blocked with the last
// Pool.Invoke() call once the current Pool runs out of its capacity, and to avoid this,
// you should instantiate a PoolWithFunc with ants.WithNonblocking(true).
func (p *PoolWithFunc) Invoke(args any) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}

	w, err := p.retrieveWorker()
	if w != nil {
		w.inputParam(args)
	}
	return err
}

// NewPoolWithFunc instantiates a PoolWithFunc with customized options.
func NewPoolWithFunc(size int, pf func(any), options ...Option) (*PoolWithFunc, error) {
	if pf == nil {
		return nil, ErrLackPoolFunc
	}

	pc, err := newPool(size, options...)
	if err != nil {
		return nil, err
	}

	pool := &PoolWithFunc{
		poolCommon: pc,
		poolFunc:   pf,
	}

	pool.workerCache.New = func() any {
		return &goWorkerWithFunc{
			pool: pool,
			args: make(chan any, workerChanCap),
		}
	}

	return pool, nil
}
