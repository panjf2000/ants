/*
 * Copyright (c) 2025. Andy Pan. All rights reserved.
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

package ants_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
)

var (
	sum int32
	wg  sync.WaitGroup
)

func incSum(i any) {
	incSumInt(i.(int32))
}

func incSumInt(i int32) {
	atomic.AddInt32(&sum, i)
	wg.Done()
}

func ExamplePool() {
	ants.Reboot() // ensure the default pool is available

	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)
	// Use the default pool.
	for i := 0; i < runTimes; i++ {
		j := i
		_ = ants.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	fmt.Printf("The result is %d\n", sum)

	atomic.StoreInt32(&sum, 0)
	wg.Add(runTimes)
	// Use the new pool.
	pool, _ := ants.NewPool(10)
	defer pool.Release()
	for i := 0; i < runTimes; i++ {
		j := i
		_ = pool.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()
	fmt.Printf("The result is %d\n", sum)

	// Output:
	// The result is 499500
	// The result is 499500
}

func ExamplePoolWithFunc() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	pool, _ := ants.NewPoolWithFunc(10, incSum)
	defer pool.Release()

	for i := 0; i < runTimes; i++ {
		_ = pool.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExamplePoolWithFuncGeneric() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	pool, _ := ants.NewPoolWithFuncGeneric(10, incSumInt)
	defer pool.Release()

	for i := 0; i < runTimes; i++ {
		_ = pool.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExampleMultiPool() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	mp, _ := ants.NewMultiPool(10, runTimes/10, ants.RoundRobin)
	defer mp.ReleaseTimeout(time.Second) // nolint:errcheck

	for i := 0; i < runTimes; i++ {
		j := i
		_ = mp.Submit(func() {
			incSumInt(int32(j))
		})
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExampleMultiPoolWithFunc() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	mp, _ := ants.NewMultiPoolWithFunc(10, runTimes/10, incSum, ants.RoundRobin)
	defer mp.ReleaseTimeout(time.Second) // nolint:errcheck

	for i := 0; i < runTimes; i++ {
		_ = mp.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}

func ExampleMultiPoolWithFuncGeneric() {
	atomic.StoreInt32(&sum, 0)
	runTimes := 1000
	wg.Add(runTimes)

	mp, _ := ants.NewMultiPoolWithFuncGeneric(10, runTimes/10, incSumInt, ants.RoundRobin)
	defer mp.ReleaseTimeout(time.Second) // nolint:errcheck

	for i := 0; i < runTimes; i++ {
		_ = mp.Invoke(int32(i))
	}
	wg.Wait()

	fmt.Printf("The result is %d\n", sum)

	// Output: The result is 499500
}
