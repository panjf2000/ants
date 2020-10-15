package ants

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolFuncRevertWorker(t *testing.T) {
	pool, err := NewPoolWithFunc(1, func(i interface{}) {
		time.Sleep(time.Millisecond * 100)
	})
	assert.Nil(t, err, "NewPool error")

	_ = pool.Invoke("test")
	assert.Nil(t, err, "Invoke error")

	pool.lock.Lock()
	time.Sleep(time.Millisecond * 300)
	// pool.Release()
	atomic.StoreInt32(&pool.state, CLOSED)
	idleWorkers := pool.workers
	for _, w := range idleWorkers {
		w.args <- nil
	}
	pool.workers = nil
	pool.lock.Unlock()

	time.Sleep(time.Millisecond * 300)
	assert.Empty(t, pool.Running(), "Memory leak")
}
