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

// PoolWithFunc accepts the tasks from client,
// it limits the total of goroutines to a given number by recycling goroutines.
type PoolWithFunc struct {
	// underlying pool implementation
	pool *Pool
	// poolFunc is the function for processing tasks.
	poolFunc func(interface{})
}

// NewPoolWithFunc generates an instance of ants pool with a specific function.
func NewPoolWithFunc(size int, pf func(interface{}), options ...Option) (*PoolWithFunc, error) {
	if size <= 0 {
		size = -1
	}

	if pf == nil {
		return nil, ErrLackPoolFunc
	}
	pool, err := NewPool(size, options...)
	if err != nil {
		return nil, err
	}

	p := &PoolWithFunc{
		pool:     pool,
		poolFunc: pf,
	}

	return p, nil
}

//---------------------------------------------------------------------------

// Invoke submits a task to pool.
//
// Note that you are allowed to call Pool.Invoke() from the current Pool.Invoke(),
// but what calls for special attention is that you will get blocked with the latest
// Pool.Invoke() call once the current Pool runs out of its capacity, and to avoid this,
// you should instantiate a PoolWithFunc with ants.WithNonblocking(true).
func (p *PoolWithFunc) Invoke(args interface{}) error {
	return p.pool.Submit(func() {
		p.poolFunc(args)
	})
}

// Running returns the amount of the currently running goroutines.
func (p *PoolWithFunc) Running() int {
	return p.pool.Running()
}

// Free returns the amount of available goroutines to work, -1 indicates this pool is unlimited.
func (p *PoolWithFunc) Free() int {
	return p.pool.Free()
}

// Cap returns the capacity of this pool.
func (p *PoolWithFunc) Cap() int {
	return p.pool.Cap()
}

// Tune changes the capacity of this pool, note that it is noneffective to the infinite or pre-allocation pool.
func (p *PoolWithFunc) Tune(size int) {
	p.pool.Tune(size)
}

// IsClosed indicates whether the pool is closed.
func (p *PoolWithFunc) IsClosed() bool {
	return p.pool.IsClosed()
}

// Release closes this pool and releases the worker queue.
func (p *PoolWithFunc) Release() {
	p.pool.Release()
}

// Reboot reboots a closed pool.
func (p *PoolWithFunc) Reboot() {
	p.pool.Reboot()
}
