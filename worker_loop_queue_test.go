package ants

import (
	"testing"
	"time"
)

func TestNewLoopQueue(t *testing.T) {
	size := 100
	q := newLoopQueue(size)
	if q.len() != 0 {
		t.Fatalf("Len error")
	}

	if q.cap() != size {
		t.Fatalf("Cap error")
	}

	if !q.isEmpty() {
		t.Fatalf("IsEmpty error")
	}

	if q.dequeue() != nil {
		t.Fatalf("Dequeue error")
	}
}

func TestLoopQueue(t *testing.T) {
	size := 10
	q := newLoopQueue(size)

	for i := 0; i < 5; i++ {
		err := q.enqueue(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}

	expired := time.Now()

	if q.len() != 5 {
		t.Fatalf("Len error")
	}

	v := q.dequeue()
	t.Log(v)

	if q.len() != 4 {
		t.Fatalf("Len error")
	}

	for i := 0; i < 6; i++ {
		err := q.enqueue(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}

	if q.len() != 10 {
		t.Fatalf("Len error")
	}

	err := q.enqueue(&goWorker{recycleTime: time.Now()})
	if err == nil {
		t.Fatalf("Enqueue error")
	}

	q.releaseExpiry(func(item *goWorker) bool {
		return item.recycleTime.Before(expired)
	})

	if q.len() != 6 {
		t.Fatalf("Len error")
	}
}
