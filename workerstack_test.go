// +build !windows

package ants

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkerStack(t *testing.T) {
	size := 100
	q := newWorkerStack(size)
	assert.EqualValues(t, 0, q.len(), "Len error")
	assert.Equal(t, true, q.isEmpty(), "IsEmpty error")
	assert.Nil(t, q.detach(), "Dequeue error")
}

func TestWorkerStack(t *testing.T) {
	q := newWorkerArray(arrayType(-1), 0)

	for i := 0; i < 5; i++ {
		err := q.insert(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}
	assert.EqualValues(t, 5, q.len(), "Len error")

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
	assert.EqualValues(t, 12, q.len(), "Len error")
	q.retrieveExpiry(time.Second)
	assert.EqualValues(t, 6, q.len(), "Len error")
}

// It seems that something wrong with time.Now() on Windows, not sure whether it is a bug on Windows,
// so exclude this test from Windows platform temporarily.
func TestSearch(t *testing.T) {
	q := newWorkerStack(0)

	// 1
	expiry1 := time.Now()

	_ = q.insert(&goWorker{recycleTime: time.Now()})

	assert.EqualValues(t, 0, q.binarySearch(0, q.len()-1, time.Now()), "index should be 0")
	assert.EqualValues(t, -1, q.binarySearch(0, q.len()-1, expiry1), "index should be -1")

	// 2
	expiry2 := time.Now()
	_ = q.insert(&goWorker{recycleTime: time.Now()})

	assert.EqualValues(t, -1, q.binarySearch(0, q.len()-1, expiry1), "index should be -1")

	assert.EqualValues(t, 0, q.binarySearch(0, q.len()-1, expiry2), "index should be 0")

	assert.EqualValues(t, 1, q.binarySearch(0, q.len()-1, time.Now()), "index should be 1")

	// more
	for i := 0; i < 5; i++ {
		_ = q.insert(&goWorker{recycleTime: time.Now()})
	}

	expiry3 := time.Now()

	_ = q.insert(&goWorker{recycleTime: expiry3})

	for i := 0; i < 10; i++ {
		_ = q.insert(&goWorker{recycleTime: time.Now()})
	}

	assert.EqualValues(t, 7, q.binarySearch(0, q.len()-1, expiry3), "index should be 7")
}
