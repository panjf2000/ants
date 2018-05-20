// Copyright (c) 2018 Andy Pan
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package ants_test

import (
	"testing"
	"github.com/panjf2000/ants"
	"sync"
	"runtime"
	"time"
)

var n = 1000000

//func demoFunc() {
//	var n int
//	for i := 0; i < 1000000; i++ {
//		n += i
//	}
//}

//func demoFunc() {
//	var n int
//	for i := 0; i < 10000; i++ {
//		n += i
//	}
//	fmt.Printf("finish task with result:%d\n", n)
//}

func forSleep() {
	time.Sleep(time.Millisecond)
}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			forSleep()
			//demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}

func TestDefaultPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		ants.Push(func() {
			forSleep()
			//demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	//t.Logf("pool capacity:%d", ants.Cap())
	//t.Logf("free workers number:%d", ants.Free())

	t.Logf("running workers number:%d", ants.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}

func TestCustomPool(t *testing.T) {
	p, _ := ants.NewPool(30000)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		p.Push(func() {
			forSleep()
			//demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	//t.Logf("pool capacity:%d", p.Cap())
	//t.Logf("free workers number:%d", p.Free())

	t.Logf("running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}
