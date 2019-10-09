package ants

import (
	"testing"
	"time"
)

func TestNewWorkerStack(t *testing.T) {
	size := 100
	q := newWorkerStack(size)
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

func TestWorkerStack(t *testing.T) {
	q := newWorkerStack(0)

	for i := 0; i < 5; i++ {
		err := q.enqueue(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}
	if q.len() != 5 {
		t.Fatalf("Len error")
	}

	expired := time.Now()

	err := q.enqueue(&goWorker{recycleTime: expired})
	if err != nil {
		t.Fatalf("Enqueue error")
	}

	for i := 0; i < 6; i++ {
		err := q.enqueue(&goWorker{recycleTime: time.Now()})
		if err != nil {
			t.Fatalf("Enqueue error")
		}
	}

	if q.len() != 12 {
		t.Fatalf("Len error")
	}

	q.releaseExpiry(func(item *goWorker) bool {
		// return item.(time.Time).Before(expired)
		return !item.recycleTime.After(expired)
	})

	if q.len() != 6 {
		t.Fatalf("Len error")
	}
}
