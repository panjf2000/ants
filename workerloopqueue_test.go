package ants

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLoopQueue(t *testing.T) {
	size := 100
	q := newWorkerLoopQueue(size)
	assert.EqualValues(t, 0, q.len(), "Len error")
	assert.Equal(t, true, q.isEmpty(), "IsEmpty error")
	assert.Nil(t, q.detach(), "Dequeue error")
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
	assert.EqualValues(t, 5, q.len(), "Len error")
	v := q.detach()
	t.Log(v)
	assert.EqualValues(t, 4, q.len(), "Len error")

	time.Sleep(time.Second)

	for i := 0; i < 6; i++ {
		err := q.insert(&goWorker{recycleTime: time.Now()})
		if err != nil {
			break
		}
	}
	assert.EqualValues(t, 10, q.len(), "Len error")

	err := q.insert(&goWorker{recycleTime: time.Now()})
	assert.Error(t, err, "Enqueue, error")

	q.retrieveExpiry(time.Second)
	assert.EqualValuesf(t, 6, q.len(), "Len error: %d", q.len())
}
