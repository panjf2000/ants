package ants

import "time"

type workerStack struct {
	items  []*goWorker
	expiry []*goWorker
}

func newWorkerStack(size int) *workerStack {
	if size < 0 {
		return nil
	}

	wq := workerStack{
		items:  make([]*goWorker, 0, size),
		expiry: make([]*goWorker, 0),
	}
	return &wq
}

func (wq *workerStack) len() int {
	return len(wq.items)
}

func (wq *workerStack) cap() int {
	return cap(wq.items)
}

func (wq *workerStack) isEmpty() bool {
	return len(wq.items) == 0
}

func (wq *workerStack) enqueue(worker *goWorker) error {
	wq.items = append(wq.items, worker)
	return nil
}

func (wq *workerStack) dequeue() *goWorker {
	l := wq.len()
	if l == 0 {
		return nil
	}

	w := wq.items[l-1]
	wq.items = wq.items[:l-1]

	return w
}

func (wq *workerStack) releaseExpiry(duration time.Duration) chan *goWorker {
	stream := make(chan *goWorker)

	n := wq.len()
	if n == 0 {
		close(stream)
		return stream
	}

	expiryTime := time.Now().Add(-duration)
	index := wq.search(0, n-1, expiryTime)

	wq.expiry = wq.expiry[:0]
	if index != -1 {
		wq.expiry = append(wq.expiry, wq.items[:index+1]...)
		m := copy(wq.items, wq.items[index+1:])
		wq.items = wq.items[:m]
	}

	go func() {
		defer close(stream)

		for i := 0; i < len(wq.expiry); i++ {
			stream <- wq.expiry[i]
		}
	}()

	return stream
}

func (wq *workerStack) search(l, r int, expiryTime time.Time) int {
	var mid int
	for l <= r {
		mid = (l + r) / 2
		if expiryTime.Before(wq.items[mid].recycleTime) {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return r
}

func (wq *workerStack) releaseAll() {
	for i := 0; i < wq.len(); i++ {
		wq.items[i].task <- nil
	}
	wq.items = wq.items[:0]
}
