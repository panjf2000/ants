//go:build !windows

package ants

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLoopQueue(t *testing.T) {
	size := 100
	q := newWorkerLoopQueue(size)
	require.EqualValues(t, 0, q.len(), "Len error")
	require.Equal(t, true, q.isEmpty(), "IsEmpty error")
	require.Nil(t, q.detach(), "Dequeue error")

	require.Nil(t, newWorkerLoopQueue(0))
}

func TestLoopQueue(t *testing.T) {
	size := 10
	q := newWorkerLoopQueue(size)

	for i := 0; i < 5; i++ {
		err := q.insert(&goWorker{lastUsed: time.Now()})
		if err != nil {
			break
		}
	}
	require.EqualValues(t, 5, q.len(), "Len error")
	_ = q.detach()
	require.EqualValues(t, 4, q.len(), "Len error")

	time.Sleep(time.Second)

	for i := 0; i < 6; i++ {
		err := q.insert(&goWorker{lastUsed: time.Now()})
		if err != nil {
			break
		}
	}
	require.EqualValues(t, 10, q.len(), "Len error")

	err := q.insert(&goWorker{lastUsed: time.Now()})
	require.Error(t, err, "Enqueue, error")

	q.refresh(time.Second)
	require.EqualValuesf(t, 6, q.len(), "Len error: %d", q.len())
}

func TestRotatedQueueSearch(t *testing.T) {
	size := 10
	q := newWorkerLoopQueue(size)

	// 1
	expiry1 := time.Now()

	_ = q.insert(&goWorker{lastUsed: time.Now()})

	require.EqualValues(t, 0, q.binarySearch(time.Now()), "index should be 0")
	require.EqualValues(t, -1, q.binarySearch(expiry1), "index should be -1")

	// 2
	expiry2 := time.Now()
	_ = q.insert(&goWorker{lastUsed: time.Now()})

	require.EqualValues(t, -1, q.binarySearch(expiry1), "index should be -1")

	require.EqualValues(t, 0, q.binarySearch(expiry2), "index should be 0")

	require.EqualValues(t, 1, q.binarySearch(time.Now()), "index should be 1")

	// more
	for i := 0; i < 5; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}

	expiry3 := time.Now()
	_ = q.insert(&goWorker{lastUsed: expiry3})

	var err error
	for err != errQueueIsFull {
		err = q.insert(&goWorker{lastUsed: time.Now()})
	}

	require.EqualValues(t, 7, q.binarySearch(expiry3), "index should be 7")

	// rotate
	for i := 0; i < 6; i++ {
		_ = q.detach()
	}

	expiry4 := time.Now()
	_ = q.insert(&goWorker{lastUsed: expiry4})

	for i := 0; i < 4; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}
	//	head = 6, tail = 5, insert direction ->
	// [expiry4, time, time, time,  time, nil/tail,  time/head, time, time, time]
	require.EqualValues(t, 0, q.binarySearch(expiry4), "index should be 0")

	for i := 0; i < 3; i++ {
		_ = q.detach()
	}
	expiry5 := time.Now()
	_ = q.insert(&goWorker{lastUsed: expiry5})

	//	head = 6, tail = 5, insert direction ->
	// [expiry4, time, time, time,  time, expiry5,  nil/tail, nil, nil, time/head]
	require.EqualValues(t, 5, q.binarySearch(expiry5), "index should be 5")

	for i := 0; i < 3; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}
	//	head = 9, tail = 9, insert direction ->
	// [expiry4, time, time, time,  time, expiry5,  time, time, time, time/head/tail]
	require.EqualValues(t, -1, q.binarySearch(expiry2), "index should be -1")

	require.EqualValues(t, 9, q.binarySearch(q.items[9].lastUsedTime()), "index should be 9")
	require.EqualValues(t, 8, q.binarySearch(time.Now()), "index should be 8")
}

func TestRetrieveExpiry(t *testing.T) {
	size := 10
	q := newWorkerLoopQueue(size)
	expirew := make([]worker, 0)
	u, _ := time.ParseDuration("1s")

	// test [ time+1s, time+1s, time+1s, time+1s, time+1s, time, time, time, time, time]
	for i := 0; i < size/2; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}
	expirew = append(expirew, q.items[:size/2]...)
	time.Sleep(u)

	for i := 0; i < size/2; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}
	workers := q.refresh(u)

	require.EqualValues(t, expirew, workers, "expired workers aren't right")

	// test [ time, time, time, time, time, time+1s, time+1s, time+1s, time+1s, time+1s]
	time.Sleep(u)

	for i := 0; i < size/2; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}
	expirew = expirew[:0]
	expirew = append(expirew, q.items[size/2:]...)

	workers2 := q.refresh(u)

	require.EqualValues(t, expirew, workers2, "expired workers aren't right")

	// test [ time+1s, time+1s, time+1s, nil, nil, time+1s, time+1s, time+1s, time+1s, time+1s]
	for i := 0; i < size/2; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}
	for i := 0; i < size/2; i++ {
		_ = q.detach()
	}
	for i := 0; i < 3; i++ {
		_ = q.insert(&goWorker{lastUsed: time.Now()})
	}
	time.Sleep(u)

	expirew = expirew[:0]
	expirew = append(expirew, q.items[0:3]...)
	expirew = append(expirew, q.items[size/2:]...)

	workers3 := q.refresh(u)

	require.EqualValues(t, expirew, workers3, "expired workers aren't right")
}
