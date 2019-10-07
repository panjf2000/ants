package ants

import (
	"time"
)

type LoopQueue struct {
	items 	[]*goWorker
	expiry 	[]*goWorker
	head	int
	tail	int
}

func NewLoopQueue(size int) *LoopQueue {
	if size == 0 {
		return nil
	}

	wq := LoopQueue{
		items:	make([]*goWorker, size+1),
		expiry: make([]*goWorker, 0),
		head:	0,
		tail:	0,
	}
	return &wq
}

func (wq *LoopQueue)Len() int {
	c := len(wq.items)
	if c == 0 {
		return 0
	}
	return (wq.tail-wq.head+c) % c
}

func (wq *LoopQueue)Cap() int {
	c := len(wq.items)
	if c == 0 {
		return 0
	}

	return c - 1
}

func (wq *LoopQueue)IsEmpty() bool {
	c := len(wq.items)
	if c == 0 {
		return true
	}
	return wq.tail == wq.head
}

func (wq *LoopQueue)Enqueue(worker *goWorker) error {
	c := len(wq.items)
	if c == 0 {
		return ErrPoolOverload
	}
	if (wq.tail+1) % c == wq.head {
		return ErrPoolOverload
	}

	wq.items[wq.tail] = worker
	wq.tail = (wq.tail+1) % c
	return nil
}

func (wq *LoopQueue)Dequeue() *goWorker {
	l := wq.Len()
	if l == 0 {
		return nil
	}

	w := wq.items[wq.head]
	wq.head = (wq.head+1) % len(wq.items)

	return w
}

func (wq *LoopQueue)ReleaseExpiry(expiry time.Duration) chan *goWorker {
	compare := time.Now().Add(-expiry)
	stream := make(chan *goWorker)

	n := wq.Len()
	if n == 0 {
		close(stream)
		return stream
	}

	c := len(wq.items)
	wq.expiry = wq.expiry[:0]

	for wq.head != wq.tail {
		if wq.items[wq.head].recycleTime.Before(compare) {
			wq.expiry = append(wq.expiry, wq.items[wq.head])
			wq.head = (wq.head+1) % c
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

func (wq *LoopQueue)ReleaseAll() {
	c := len(wq.items)
	if wq.Len() == 0 {
		return
	}

	for wq.head != wq.tail {
		wq.items[wq.head].task <- nil
		wq.head = (wq.head+1) % c
	}
	wq.items = wq.items[:0]
	wq.expiry = wq.expiry[:0]
	wq.head = 0
	wq.tail = 0
}
