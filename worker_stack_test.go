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
	"runtime"
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
	if runtime.GOOS == "windows" { // time.Now() doesn't seem to be precise on Windows
		t.Skip("Skip this test on Windows platform")
	}

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
