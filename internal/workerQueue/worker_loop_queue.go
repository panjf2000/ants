package ants

type LoopQueue struct {
	items     []interface{}
	expiry    []interface{}
	head      int
	tail      int
	remainder int
}

func NewLoopQueue(size int) *LoopQueue {
	if size == 0 {
		return nil
	}

	wq := LoopQueue{
		items:     make([]interface{}, size+1),
		expiry:    make([]interface{}, 0),
		head:      0,
		tail:      0,
		remainder: size + 1,
	}

	return &wq
}

func (wq *LoopQueue) Len() int {
	if wq.remainder == 0 {
		return 0
	}

	return (wq.tail - wq.head + wq.remainder) % wq.remainder
}

func (wq *LoopQueue) Cap() int {
	if wq.remainder == 0 {
		return 0
	}
	return wq.remainder - 1
}

func (wq *LoopQueue) IsEmpty() bool {
	return wq.tail == wq.head
}

func (wq *LoopQueue) Enqueue(worker interface{}) error {
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

func (wq *LoopQueue) Dequeue() interface{} {
	if wq.Len() == 0 {
		return nil
	}

	w := wq.items[wq.head]
	wq.head = (wq.head + 1) % wq.remainder

	return w
}

func (wq *LoopQueue) ReleaseExpiry(isExpiry func(item interface{}) bool) chan interface{} {
	stream := make(chan interface{})

	if wq.Len() == 0 {
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

func (wq *LoopQueue) ReleaseAll(free func(item interface{})) {
	if wq.Len() == 0 {
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
