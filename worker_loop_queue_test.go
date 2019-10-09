package ants

import (
	"testing"
	"time"
)

func TestNewLoopQueue(t *testing.T) {
	size := 100
	q := newWorkerLoopQueue(size)
	if q.len() != 0 {
		t.Fatalf("Len error")
	}

	if !q.isEmpty() {
		t.Fatalf("IsEmpty error")
	}

	if q.detach() != nil {
		t.Fatalf("Dequeue error")
	}
}

func TestLoopQueue(t *testing.T) {
	size := 10
	q := newWorkerLoopQueue(size)

	for i := 0; i < 5; i++ {
		err := q.insert(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}

	if q.len() != 5 {
		t.Fatalf("Len error")
	}

	v := q.detach()
	t.Log(v)

	if q.len() != 4 {
		t.Fatalf("Len error")
	}

	time.Sleep(time.Second)

	for i := 0; i < 6; i++ {
		err := q.insert(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}

	if q.len() != 10 {
		t.Fatalf("Len error")
	}

	err := q.insert(&goWorker{recycleTime: time.Now()})
	if err == nil {
		t.Fatalf("Enqueue error")
	}

	q.findOutExpiry(time.Second)

	if q.len() != 6 {
		t.Fatalf("Len error: %d", q.len())
	}
}
