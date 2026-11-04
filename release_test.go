/*
 * Copyright (c) 2019. Ants Authors. All rights reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package ants

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A caller of Submit()/Invoke() may pass the IsClosed() check right before the
// pool is released, retrieveWorker() must not hand it a worker after that.
func TestRetrieveWorkerAfterRelease(t *testing.T) {
	p, err := NewPool(10)
	require.NoError(t, err)
	pf, err := NewPoolWithFunc(10, func(any) {})
	require.NoError(t, err)
	pfg, err := NewPoolWithFuncGeneric(10, func(int) {})
	require.NoError(t, err)

	for _, pc := range []*poolCommon{p.poolCommon, pf.poolCommon, pfg.poolCommon} {
		pc.Release()
		w, err := pc.retrieveWorker()
		require.ErrorIs(t, err, ErrPoolClosed)
		require.Nil(t, w)
		require.EqualValues(t, 0, pc.Running())
		select {
		case <-pc.allDone:
		case <-time.After(time.Second):
			t.Fatal("allDone is not closed after Release")
		}
	}
}

// Release and reboot a pool repeatedly while tasks are being submitted:
// ReleaseTimeout() must not return before all workers are gone, and a worker
// exiting from the previous run must not close allDone of the rebooted pool.
func TestReleaseRebootWithConcurrentSubmit(t *testing.T) {
	p, err := NewPool(4, WithExpiryDuration(5*time.Millisecond))
	require.NoError(t, err)

	var stop atomic.Bool
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !stop.Load() {
				_ = p.Submit(func() { time.Sleep(time.Microsecond) })
			}
		}()
	}
	defer func() {
		stop.Store(true)
		wg.Wait()
		p.Release()
	}()

	for i := 0; i < 200; i++ {
		require.NoError(t, p.ReleaseTimeout(5*time.Second), "iteration %d", i)
		require.Zerof(t, p.Running(), "iteration %d: workers still running after ReleaseTimeout", i)
		p.Reboot()
	}
}
