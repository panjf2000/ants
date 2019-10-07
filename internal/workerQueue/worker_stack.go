package ants

type WorkerStack struct {
	items  []interface{}
	expiry []interface{}
}

func NewWorkerStack(size int) *WorkerStack {
	wq := WorkerStack{
		items:  make([]interface{}, 0, size),
		expiry: make([]interface{}, 0),
	}
	return &wq
}

func (wq *WorkerStack) Len() int {
	return len(wq.items)
}

func (wq *WorkerStack) Cap() int {
	return cap(wq.items)
}

func (wq *WorkerStack) IsEmpty() bool {
	return len(wq.items) == 0
}

func (wq *WorkerStack) Enqueue(worker interface{}) error {
	wq.items = append(wq.items, worker)
	return nil
}

func (wq *WorkerStack) Dequeue() interface{} {
	l := wq.Len()
	if l == 0 {
		return nil
	}

	w := wq.items[l-1]
	wq.items = wq.items[:l-1]

	return w
}

func (wq *WorkerStack) ReleaseExpiry(isExpiry func(item interface{}) bool) chan interface{} {
	stream := make(chan interface{})

	n := wq.Len()
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

func (wq *WorkerStack) search(l, r int, isExpiry func(item interface{}) bool) int {
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

func (wq *WorkerStack) ReleaseAll(free func(item interface{})) {
	for i := 0; i < wq.Len(); i++ {
		free(wq.items[i])
	}
	wq.items = wq.items[:0]
}
