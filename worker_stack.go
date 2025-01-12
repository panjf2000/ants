/*
 * Copyright (c) 2019. Ants Authors. All rights reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package ants

import "time"

type workerStack struct {
	items  []worker
	expiry []worker
}

func newWorkerStack(size int) *workerStack {
	return &workerStack{
		items: make([]worker, 0, size),
	}
}

func (ws *workerStack) len() int {
	return len(ws.items)
}

func (ws *workerStack) isEmpty() bool {
	return len(ws.items) == 0
}

func (ws *workerStack) insert(w worker) error {
	ws.items = append(ws.items, w)
	return nil
}

func (ws *workerStack) detach() worker {
	l := ws.len()
	if l == 0 {
		return nil
	}

	w := ws.items[l-1]
	ws.items[l-1] = nil // avoid memory leaks
	ws.items = ws.items[:l-1]

	return w
}

func (ws *workerStack) refresh(duration time.Duration) []worker {
	n := ws.len()
	if n == 0 {
		return nil
	}

	expiryTime := time.Now().Add(-duration)
	index := ws.binarySearch(0, n-1, expiryTime)

	ws.expiry = ws.expiry[:0]
	if index != -1 {
		ws.expiry = append(ws.expiry, ws.items[:index+1]...)
		m := copy(ws.items, ws.items[index+1:])
		for i := m; i < n; i++ {
			ws.items[i] = nil
		}
		ws.items = ws.items[:m]
	}
	return ws.expiry
}

func (ws *workerStack) binarySearch(l, r int, expiryTime time.Time) int {
	for l <= r {
		mid := l + ((r - l) >> 1) // avoid overflow when computing mid
		if expiryTime.Before(ws.items[mid].lastUsedTime()) {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return r
}

func (ws *workerStack) reset() {
	for i := 0; i < ws.len(); i++ {
		ws.items[i].finish()
		ws.items[i] = nil
	}
	ws.items = ws.items[:0]
}
