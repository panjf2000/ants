package ants

import (
	"testing"
	"time"
)

func TestNewLoopQueue(t *testing.T) {
	size := 100
	q := NewLoopQueue(size)
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

func TestLoopQueue(t *testing.T) {
	size := 10
	q := NewLoopQueue(size)

	for i := 0; i < 5; i++ {
		err := q.Enqueue(time.Now())
		if err != nil {
			break
		}
	}

	expired := time.Now()

	if q.Len() != 5 {
		t.Fatalf("Len error")
	}

	v := q.Dequeue().(time.Time)
	t.Log(v)

	if q.Len() != 4 {
		t.Fatalf("Len error")
	}

	for i := 0; i < 6; i++ {
		err := q.Enqueue(time.Now())
		if err != nil {
			break
		}
	}

	if q.Len() != 10 {
		t.Fatalf("Len error")
	}

	err := q.Enqueue(time.Now())
	if err == nil {
		t.Fatalf("Enqueue error")
	}

	q.ReleaseExpiry(func(item interface{}) bool {
		return item.(time.Time).Before(expired)
	})

	if q.Len() != 6 {
		t.Fatalf("Len error")
	}
}
