package ants

import "time"

type loopQueue struct {
	items  []worker
	expiry []worker
	head   int
	tail   int
	size   int
	isFull bool
}

func newWorkerLoopQueue(size int) *loopQueue {
	return &loopQueue{
		items: make([]worker, size),
		size:  size,
	}
}

func (wq *loopQueue) len() int {
	if wq.size == 0 || wq.isEmpty() {
		return 0
	}

	if wq.head == wq.tail && wq.isFull {
		return wq.size
	}

	if wq.tail > wq.head {
		return wq.tail - wq.head
	}

	return wq.size - wq.head + wq.tail
}

func (wq *loopQueue) isEmpty() bool {
	return wq.head == wq.tail && !wq.isFull
}

func (wq *loopQueue) insert(w worker) error {
	if wq.size == 0 {
		return errQueueIsReleased
	}

	if wq.isFull {
		return errQueueIsFull
	}
	wq.items[wq.tail] = w
	wq.tail = (wq.tail + 1) % wq.size

	if wq.tail == wq.head {
		wq.isFull = true
	}

	return nil
}

func (wq *loopQueue) detach() worker {
	if wq.isEmpty() {
		return nil
	}

	w := wq.items[wq.head]
	wq.items[wq.head] = nil
	wq.head = (wq.head + 1) % wq.size

	wq.isFull = false

	return w
}

func (wq *loopQueue) refresh(duration time.Duration) []worker {
	expiryTime := time.Now().Add(-duration)
	index := wq.binarySearch(expiryTime)
	if index == -1 {
		return nil
	}
	wq.expiry = wq.expiry[:0]

	if wq.head <= index {
		wq.expiry = append(wq.expiry, wq.items[wq.head:index+1]...)
		for i := wq.head; i < index+1; i++ {
			wq.items[i] = nil
		}
	} else {
		wq.expiry = append(wq.expiry, wq.items[0:index+1]...)
		wq.expiry = append(wq.expiry, wq.items[wq.head:]...)
		for i := 0; i < index+1; i++ {
			wq.items[i] = nil
		}
		for i := wq.head; i < wq.size; i++ {
			wq.items[i] = nil
		}
	}
	head := (index + 1) % wq.size
	wq.head = head
	if len(wq.expiry) > 0 {
		wq.isFull = false
	}

	return wq.expiry
}

func (wq *loopQueue) binarySearch(expiryTime time.Time) int {
	var mid, nlen, basel, tmid int
	nlen = len(wq.items)

	// if no need to remove work, return -1
	if wq.isEmpty() || expiryTime.Before(wq.items[wq.head].lastUsedTime()) {
		return -1
	}

	// example
	// size = 8, head = 7, tail = 4
	// [ 2, 3, 4, 5, nil, nil, nil,  1]  true position
	//   0  1  2  3    4   5     6   7
	//              tail          head
	//
	//   1  2  3  4  nil nil   nil   0   mapped position
	//            r                  l

	// base algorithm is a copy from worker_stack
	// map head and tail to effective left and right
	r := (wq.tail - 1 - wq.head + nlen) % nlen
	basel = wq.head
	l := 0
	for l <= r {
		mid = l + ((r - l) >> 1) // avoid overflow when computing mid
		// calculate true mid position from mapped mid position
		tmid = (mid + basel + nlen) % nlen
		if expiryTime.Before(wq.items[tmid].lastUsedTime()) {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	// return true position from mapped position
	return (r + basel + nlen) % nlen
}

func (wq *loopQueue) reset() {
	if wq.isEmpty() {
		return
	}

retry:
	if w := wq.detach(); w != nil {
		w.finish()
		goto retry
	}
	wq.items = wq.items[:0]
	wq.size = 0
	wq.head = 0
	wq.tail = 0
}
