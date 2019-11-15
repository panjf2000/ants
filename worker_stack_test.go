// +build !windows

package ants

import (
	"testing"
	"time"
)

func TestNewWorkerStack(t *testing.T) {
	size := 100
	q := newWorkerStack(size)
	if q.len() != 0 {
		t.Fatal("Len error")
	}

	if !q.isEmpty() {
		t.Fatal("IsEmpty error")
	}

	if q.detach() != nil {
		t.Fatal("Dequeue error")
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
		t.Fatal("Len error")
	}

	expired := time.Now()

	err := q.insert(&goWorker{recycleTime: expired})
	if err != nil {
		t.Fatal("Enqueue error")
	}

	time.Sleep(time.Second)

	for i := 0; i < 6; i++ {
		err := q.insert(&goWorker{recycleTime: time.Now()})
		if err != nil {
			t.Fatal("Enqueue error")
		}
	}

	if q.len() != 12 {
		t.Fatal("Len error")
	}

	q.retrieveExpiry(time.Second)

	if q.len() != 6 {
		t.Fatal("Len error")
	}
}

// It seems that something wrong with time.Now() on Windows, not sure whether it is a bug on Windows, so exclude this test
// from Windows platform temporarily.
func TestSearch(t *testing.T) {
	q := newWorkerStack(0)

	// 1
	expiry1 := time.Now()

	_ = q.insert(&goWorker{recycleTime: time.Now()})

	index := q.binarySearch(0, q.len()-1, time.Now())
	if index != 0 {
		t.Fatal("index should be 0")
	}

	index = q.binarySearch(0, q.len()-1, expiry1)
	if index != -1 {
		t.Fatal("index should be -1")
	}

	// 2
	expiry2 := time.Now()
	_ = q.insert(&goWorker{recycleTime: time.Now()})

	index = q.binarySearch(0, q.len()-1, expiry1)
	if index != -1 {
		t.Fatal("index should be -1")
	}

	index = q.binarySearch(0, q.len()-1, expiry2)
	if index != 0 {
		t.Fatal("index should be 0")
	}

	index = q.binarySearch(0, q.len()-1, time.Now())
	if index != 1 {
		t.Fatal("index should be 1")
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

	index = q.binarySearch(0, q.len()-1, expiry3)
	if index != 7 {
		t.Fatal("index should be 7")
	}
}
