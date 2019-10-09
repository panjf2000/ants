package ants

type loopQueue struct {
	items     []*goWorker
	expiry    []*goWorker
	head      int
	tail      int
	remainder int
}

func newLoopQueue(size int) *loopQueue {
	if size == 0 {
		return nil
	}

	wq := loopQueue{
		items:     make([]*goWorker, size+1),
		expiry:    make([]*goWorker, 0),
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

func (wq *loopQueue) cap() int {
	if wq.remainder == 0 {
		return 0
	}
	return wq.remainder - 1
}

func (wq *loopQueue) isEmpty() bool {
	return wq.tail == wq.head
}

func (wq *loopQueue) enqueue(worker *goWorker) error {
	if wq.remainder == 0 {
		return ErrQueueLengthIsZero
	}
	if (wq.tail+1)%wq.remainder == wq.head {
		return ErrQueueIsFull
	}

	wq.items[wq.tail] = worker
	wq.tail = (wq.tail + 1) % wq.remainder

	return nil
}

func (wq *loopQueue) dequeue() *goWorker {
	if wq.len() == 0 {
		return nil
	}

	w := wq.items[wq.head]
	wq.head = (wq.head + 1) % wq.remainder

	return w
}

func (wq *loopQueue) releaseExpiry(isExpiry func(item *goWorker) bool) chan *goWorker {
	stream := make(chan *goWorker)

	if wq.len() == 0 {
		close(stream)
		return stream
	}

	wq.expiry = wq.expiry[:0]

	for wq.head != wq.tail {
		if isExpiry(wq.items[wq.head]) {
			wq.expiry = append(wq.expiry, wq.items[wq.head])
			wq.head = (wq.head + 1) % wq.remainder
			continue
		}
		break
	}

	go func() {
		defer close(stream)

		for i := 0; i < len(wq.expiry); i++ {
			stream <- wq.expiry[i]
		}
	}()

	return stream
}

//func (wq *LoopQueue)search(compareTime time.Time, l, r int) int {
//	if l == r {
//		if wq.items[l].recycleTime.After(compareTime) {
//			return -1
//		} else {
//			return l
//		}
//	}
//
//	c := cap(wq.items)
//	mid := ((r-l+c)/2 + l) % c
//	if mid == l {
//		return wq.search(compareTime, l, l)
//	} else if wq.items[mid].recycleTime.After(compareTime) {
//		return wq.search(compareTime, l, mid-1)
//	} else {
//		return wq.search(compareTime, mid+1, r)
//	}
//}

func (wq *loopQueue) releaseAll(free func(item *goWorker)) {
	if wq.len() == 0 {
		return
	}

	for wq.head != wq.tail {
		free(wq.items[wq.head])
		wq.head = (wq.head + 1) % wq.remainder
	}
	wq.items = wq.items[:0]
	wq.remainder = 0
	wq.head = 0
	wq.tail = 0
}
