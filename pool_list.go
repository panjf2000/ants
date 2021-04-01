// MIT License

// Copyright (c) 2021 yddeng

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
	"sync"
	"sync/atomic"
)

// PoolList accepts the tasks from client,
// it limits the total of goroutines to a given number by recycling goroutines.
type PoolList struct {
	// inline PoolWithFunc.
	*PoolWithFunc

	// head, rear of the arg list.
	head, rear *argNode

	// the result of processing by the function, client reads the result from output channel.
	output chan interface{}

	// release
	checked int32

	closed bool

	mtx sync.Mutex
}

// argNode is used to store argument, node of the request list.
type argNode struct {
	// argument or result, it depends on the state of done.
	arg interface{}

	// done is used to mark complete.
	done uint32

	// the next node of the current node.
	next *argNode
}

// NewPoolList generates an instance of ants pool with a specific function
// and a output channel, client reads the result from the channel.
func NewPoolList(size int, output chan interface{}, pf func(interface{}) interface{}, options ...Option) (p *PoolList, err error) {
	p = new(PoolList)
	p.output = output

	if p.PoolWithFunc, err = NewPoolWithFunc(size, func(i interface{}) {
		req := i.(*argNode)
		req.arg = pf(req.arg)
		// change the state of done.
		atomic.StoreUint32(&req.done, 1)

		// call checkHead
		if atomic.CompareAndSwapInt32(&p.checked, 0, 1) {
			p.checkHead()
		}
	}, options...); err != nil {
		return
	}

	return
}

// Invoke submits a task to pool. it rewrite the PoolWhitFunc.Invoke
func (p *PoolList) Invoke(args interface{}) error {
	req := &argNode{arg: args}

	p.mtx.Lock()
	if p.closed {
		p.mtx.Unlock()
		return ErrPoolClosed
	}

	if p.head == nil {
		p.head = req
	}
	if p.rear != nil {
		p.rear.next = req
	}
	p.rear = req

	p.mtx.Unlock()

	return p.PoolWithFunc.Invoke(req)
}

func (p *PoolList) Release() {
	p.mtx.Lock()
	p.closed = true
	p.head = nil
	p.rear = nil
	p.output <- nil
	p.mtx.Unlock()
	p.PoolWithFunc.Release()
}

func (p *PoolList) Reboot() {
	p.mtx.Lock()
	p.closed = false
	p.mtx.Unlock()
	p.PoolWithFunc.Reboot()
}

// checkHead check that the header node is complete,
// place the finished result in output channel.
func (p *PoolList) checkHead() {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	defer func() {
		atomic.StoreInt32(&p.checked, 0)
	}()

	for n := p.head; n != nil; n = n.next {
		if atomic.LoadUint32(&n.done) == 0 {
			break
		}
		p.output <- n.arg
		p.head = n.next
	}

	if p.head == nil {
		p.rear = nil
	}

}
