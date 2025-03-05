package ants

import (
	"github.com/stretchr/testify/require"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	sum int32
	wg  sync.WaitGroup
)

func incSumInt(i int32) {
	atomic.AddInt32(&sum, i)
	wg.Done()
}

func TestNewPool(t *testing.T) {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	pool, _ := NewPool(10)
	defer pool.Release()
	// Use the default pool.
	for i := 0; i < runTimes; i++ {
		j := i
		_ = pool.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	require.EqualValues(t, 499500, sum, "The result should be 499500")

	atomic.StoreInt32(&sum, 0)
	wg.Add(runTimes)
	_ = pool.ReleaseTimeout(time.Second) // use both Release and ReleaseTimeout will occur panic
	pool.Reboot()

	for i := 0; i < runTimes; i++ {
		j := i
		_ = pool.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	require.EqualValues(t, 499500, sum, "The result should be 499500")
}

func TestNewPoolWithPreAlloc(t *testing.T) {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	pool, _ := NewPool(10, WithPreAlloc(true))
	defer pool.Release()
	// Use the default pool.
	for i := 0; i < runTimes; i++ {
		j := i
		_ = pool.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	require.EqualValues(t, 499500, sum, "The result should be 499500")

	atomic.StoreInt32(&sum, 0)
	_ = pool.ReleaseTimeout(time.Second)
	pool.Reboot()

	wg.Add(runTimes)
	for i := 0; i < runTimes; i++ {
		j := i
		_ = pool.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	require.EqualValues(t, 499500, sum, "The result should be 499500")
}
