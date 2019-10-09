package ants

type workerStack struct {
	items  []*goWorker
	expiry []*goWorker
}

func newWorkerStack(size int) *workerStack {
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

func (wq *workerStack) releaseExpiry(isExpiry func(item *goWorker) bool) chan *goWorker {
	stream := make(chan *goWorker)

	n := wq.len()
	if n == 0 {
		close(stream)
		return stream
	}

	index := wq.search(0, n-1, isExpiry)

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

func (wq *workerStack) search(l, r int, isExpiry func(item *goWorker) bool) int {
	if l == r {
		if isExpiry(wq.items[l]) {
			return l
		} else {
			return -1
		}
	}

	mid := (r-l)/2 + l
	if mid == l {
		return wq.search(l, l, isExpiry)
	} else if isExpiry(wq.items[mid]) {
		return wq.search(mid, r, isExpiry)
	} else {
		return wq.search(l, mid, isExpiry)
	}
}

func (wq *workerStack) releaseAll(free func(item *goWorker)) {
	for i := 0; i < wq.len(); i++ {
		free(wq.items[i])
	}
	wq.items = wq.items[:0]
}
