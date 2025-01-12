//go:build !windows

package ants

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewWorkerStack(t *testing.T) {
	size := 100
	q := newWorkerStack(size)
	require.EqualValues(t, 0, q.len(), "Len error")
	require.Equal(t, true, q.isEmpty(), "IsEmpty error")
	require.Nil(t, q.detach(), "Dequeue error")
}

func TestWorkerStack(t *testing.T) {
	q := newWorkerQueue(queueType(-1), 0)

	for i := 0; i < 5; i++ {
		err := q.insert(&goWorker{lastUsed: time.Now()})
		if err != nil {
			break
		}
	}
	require.EqualValues(t, 5, q.len(), "Len error")

	expired := time.Now()

	err := q.insert(&goWorker{lastUsed: expired})
	if err != nil {
		t.Fatal("Enqueue error")
	}

	time.Sleep(time.Second)

	for i := 0; i < 6; i++ {
		err := q.insert(&goWorker{lastUsed: time.Now()})
		if err != nil {
			t.Fatal("Enqueue error")
		}
	}
	require.EqualValues(t, 12, q.len(), "Len error")
	q.refresh(time.Second)
	require.EqualValues(t, 6, q.len(), "Len error")
}

// It seems that something wrong with time.Now() on Windows, not sure whether it is a bug on Windows,
// so exclude this test from Windows platform temporarily.
func TestSearch(t *testing.T) {
	q := newWorkerStack(0)

	// 1
	expiry1 := time.Now()

	_ = q.insert(&goWorker{lastUsed: time.Now()})

	require.EqualValues(t, 0, q.binarySearch(0, q.len()-1, time.Now()), "index should be 0")
	require.EqualValues(t, -1, q.binarySearch(0, q.len()-1, expiry1), "index should be -1")

	// 2
	expiry2 := time.Now()
	_ = q.insert(&goWorker{lastUsed: time.Now()})

	require.EqualValues(t, -1, q.binarySearch(0, q.len()-1, expiry1), "index should be -1")

	require.EqualValues(t, 0, q.binarySearch(0, q.len()-1, expiry2), "index should be 0")

	require.EqualValues(t, 1, q.binarySearch(0, q.len()-1, time.Now()), "index should be 1")

	// more
	for i := 0; i < 5; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}

	expiry3 := time.Now()

	_ = q.insert(&goWorker{lastUsed: expiry3})

	for i := 0; i < 10; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}

	require.EqualValues(t, 7, q.binarySearch(0, q.len()-1, expiry3), "index should be 7")
}
