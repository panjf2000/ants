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

	if !q.isEmpty() {
		t.Fatalf("IsEmpty error")
	}

	if q.detach() != nil {
		t.Fatalf("Dequeue error")
	}
}

func TestWorkerStack(t *testing.T) {
	q := newWorkerArray(arrayType(-1), 0)

	for i := 0; i < 5; i++ {
		err := q.insert(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}
	if q.len() != 5 {
		t.Fatalf("Len error")
	}

	expired := time.Now()

	err := q.insert(&goWorker{recycleTime: expired})
	if err != nil {
		t.Fatalf("Enqueue error")
	}

	time.Sleep(time.Second)

	for i := 0; i < 6; i++ {
		err := q.insert(&goWorker{recycleTime: time.Now()})
		if err != nil {
			t.Fatalf("Enqueue error")
		}
	}

	if q.len() != 12 {
		t.Fatalf("Len error")
	}

	q.findOutExpiry(time.Second)

	if q.len() != 6 {
		t.Fatalf("Len error")
	}
}

func TestSearch(t *testing.T) {
	q := newWorkerStack(0)

	// 1
	expiry1 := time.Now()

	_ = q.insert(&goWorker{recycleTime: time.Now()})

	index := q.search(0, q.len()-1, time.Now())
	if index != 0 {
		t.Fatalf("should is 0")
	}

	index = q.search(0, q.len()-1, expiry1)
	if index != -1 {
		t.Fatalf("should is -1")
	}

	// 2
	expiry2 := time.Now()
	_ = q.insert(&goWorker{recycleTime: time.Now()})

	index = q.search(0, q.len()-1, expiry1)
	if index != -1 {
		t.Fatalf("should is -1")
	}

	index = q.search(0, q.len()-1, expiry2)
	if index != 0 {
		t.Fatalf("should is 0")
	}

	index = q.search(0, q.len()-1, time.Now())
	if index != 1 {
		t.Fatalf("should is 1")
	}

	// more
	for i := 0; i < 5; i++ {
		_ = q.insert(&goWorker{recycleTime: time.Now()})
	}

	expiry3 := time.Now()

	_ = q.insert(&goWorker{recycleTime: expiry3})

	for i := 0; i < 10; i++ {
		_ = q.insert(&goWorker{recycleTime: time.Now()})
	}

	index = q.search(0, q.len()-1, expiry3)
	if index != 7 {
		t.Fatalf("should is 7")
	}
}
