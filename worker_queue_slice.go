package ants

import "time"

type SliceQueue struct {
	items 	[]*goWorker
	expiry 	[]*goWorker
}

func NewSliceQueue(size int) *SliceQueue {
	wq := SliceQueue{
		items:	make([]*goWorker, 0, size),
		expiry: make([]*goWorker, 0),
	}
	return &wq
}

func (wq *SliceQueue)Len() int {
	return len(wq.items)
}

func (wq *SliceQueue)Cap() int {
	return cap(wq.items)
}

func (wq *SliceQueue)IsEmpty() bool {
	return len(wq.items) == 0
}

func (wq *SliceQueue)Enqueue(worker *goWorker) error {
	wq.items = append(wq.items, worker)
	return nil
}

func (wq *SliceQueue)Dequeue() *goWorker {
	l := wq.Len()
	if l == 0 {
		return nil
	}

	//w := wq.items[0]
	//wq.items = wq.items[1:]
	// or
	w := wq.items[l-1]
	wq.items = wq.items[:l-1]

	return w
}

func (wq *SliceQueue)ReleaseExpiry(expiry time.Duration) chan *goWorker {
	compare := time.Now().Add(-expiry)
	stream := make(chan *goWorker)

	n := wq.Len()
	if n == 0 {
		close(stream)
		return stream
	}

	index := wq.search(compare, 0, n-1)

	wq.expiry = wq.expiry[:0]
	if index != -1 { //
		wq.expiry = append(wq.expiry, wq.items[:index+1]...)

		//wq.items = wq.items[index+1:]
		// or copy
		m := copy(wq.items, wq.items[index+1:])
		wq.items = wq.items[:m+1]
	}

	go func() {
		defer close(stream)

		for i := 0; i < len(wq.expiry); i++ {
			stream <- wq.expiry[i]
		}
	}()

	return stream
}

func (wq *SliceQueue)search(compareTime time.Time, l, r int) int {
	if l == r {
		if wq.items[l].recycleTime.After(compareTime) {
			return -1
		} else {
			return l
		}
	}

	mid := (r - l) / 2 + l
	if mid == l {
		return wq.search(compareTime, l, l)
	} else if wq.items[mid].recycleTime.After(compareTime) {
		return wq.search(compareTime, l, mid-1)
	} else {
		return wq.search(compareTime, mid+1, r)
	}
}

func (wq *SliceQueue)ReleaseAll() {
	for i := 0; i < wq.Len(); i++ {
		wq.items[i].task <- nil
	}
	wq.items = wq.items[:0]
	wq.expiry = wq.expiry[:0]
}
