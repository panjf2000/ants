package ants

import "time"

type loopQueue struct {
	items     []*goWorker
	expiry    []*goWorker
	head      int
	tail      int
	remainder int
}

func newWorkerLoopQueue(size int) *loopQueue {
	if size <= 0 {
		return nil
	}

	wq := loopQueue{
		items:     make([]*goWorker, size+1),
		head:      0,
		tail:      0,
		remainder: size + 1,
	}

	return &wq
}

func (wq *loopQueue) len() int {
	if wq.remainder == 0 {
		return 0
	}

	return (wq.tail - wq.head + wq.remainder) % wq.remainder
}

func (wq *loopQueue) isEmpty() bool {
	return wq.tail == wq.head
}

func (wq *loopQueue) insert(worker *goWorker) error {
	if wq.remainder == 0 {
		return ErrQueueLengthIsZero
	}
	next := (wq.tail + 1) % wq.remainder
	if next == wq.head {
		return ErrQueueIsFull
	}

	wq.items[wq.tail] = worker
	wq.tail = next

	return nil
}

func (wq *loopQueue) detach() *goWorker {
	if wq.len() == 0 {
		return nil
	}

	w := wq.items[wq.head]
	wq.head = (wq.head + 1) % wq.remainder

	return w
}

func (wq *loopQueue) findOutExpiry(duration time.Duration) []*goWorker {
	if wq.len() == 0 {
		return nil
	}

	wq.expiry = wq.expiry[:0]
	expiryTime := time.Now().Add(-duration)

	for wq.head != wq.tail {
		if expiryTime.Before(wq.items[wq.head].recycleTime) {
			break
		}
		wq.expiry = append(wq.expiry, wq.items[wq.head])
		wq.head = (wq.head + 1) % wq.remainder
	}
	return wq.expiry
}

func (wq *loopQueue) release() {
	if wq.len() == 0 {
		return
	}

	for wq.head != wq.tail {
		wq.items[wq.head].task <- nil
		wq.head = (wq.head + 1) % wq.remainder
	}
	wq.items = wq.items[:0]
	wq.remainder = 0
	wq.head = 0
	wq.tail = 0
}
