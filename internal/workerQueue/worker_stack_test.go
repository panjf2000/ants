package ants

import (
	"testing"
	"time"
)

func TestNewWorkerStack(t *testing.T) {
	size := 100
	q := NewWorkerStack(size)
	if q.Len() != 0 {
		t.Fatalf("Len error")
	}

	if q.Cap() != size {
		t.Fatalf("Cap error")
	}

	if !q.IsEmpty() {
		t.Fatalf("IsEmpty error")
	}

	if q.Dequeue() != nil {
		t.Fatalf("Dequeue error")
	}
}

func TestWorkerStack(t *testing.T) {
	q := NewWorkerStack(0)

	for i := 0; i < 5; i++ {
		err := q.Enqueue(time.Now())
		if err != nil {
			break
		}
	}
	if q.Len() != 5 {
		t.Fatalf("Len error")
	}

	expired := time.Now()

	err := q.Enqueue(expired)
	if err != nil {
		t.Fatalf("Enqueue error")
	}

	for i := 0; i < 6; i++ {
		err := q.Enqueue(time.Now())
		if err != nil {
			t.Fatalf("Enqueue error")
		}
	}

	if q.Len() != 12 {
		t.Fatalf("Len error")
	}

	q.ReleaseExpiry(func(item interface{}) bool {
		// return item.(time.Time).Before(expired)
		return !item.(time.Time).After(expired)
	})

	if q.Len() != 6 {
		t.Fatalf("Len error")
	}
}
