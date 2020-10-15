package ants

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolRevertWorker(t *testing.T) {
	pool, err := NewPool(1)
	assert.Nil(t, err, "NewPool error")

	err = pool.Submit(func() {
		time.Sleep(time.Millisecond * 100)
	})
	assert.Nil(t, err, "Submit error")

	pool.lock.Lock()
	time.Sleep(time.Millisecond * 300)
	// pool.Release()
	atomic.StoreInt32(&pool.state, CLOSED)
	pool.workers.reset()
	pool.lock.Unlock()

	time.Sleep(time.Millisecond * 300)
	assert.Empty(t, pool.Running(), "Memory leak")
}
