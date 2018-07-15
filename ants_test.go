// MIT License

// Copyright (c) 2018 Andy Pan

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ants_test

import (
	"runtime"
	"sync"
	"testing"

	"github.com/panjf2000/ants"
)

var n = 1000000

func TestAntsPoolWithFunc(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(AntsSize, func(i interface{}) error {
		demoPoolFunc(i)
		wg.Done()
		return nil
	})
	defer p.Release()

	for i := 0; i < n; i++ {
		wg.Add(1)
		p.Serve(Param)
	}
	wg.Wait()

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/GiB)
}

func TestAntsPool(t *testing.T) {
	defer ants.Release()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		ants.Submit(func() error {
			demoFunc()
			wg.Done()
			return nil
		})
	}
	wg.Wait()

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d MB", mem.TotalAlloc/MiB)
}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d MB", mem.TotalAlloc/MiB)
}

func TestCodeCov(t *testing.T) {
	p0, _ := ants.NewPool(AntsSize)
	defer p0.Release()
	for i := 0; i < n; i++ {
		p0.Submit(demoFunc)
	}
	t.Logf("pool, capacity:%d", p0.Cap())
	t.Logf("pool, running workers number:%d", p0.Running())
	t.Logf("pool, free workers number:%d", p0.Free())
	p0.ReSize(AntsSize / 2)
	t.Logf("pool, after resize, capacity:%d", p0.Cap())

	p, _ := ants.NewPoolWithFunc(AntsSize, demoPoolFunc)
	defer p.Release()
	for i := 0; i < n; i++ {
		p.Serve(Param)
	}
	t.Logf("pool with func, capacity:%d", p.Cap())
	t.Logf("pool with func, running workers number:%d", p.Running())
	t.Logf("pool with func, free workers number:%d", p.Free())
	p.ReSize(AntsSize / 2)
	t.Logf("pool with func, after resize, capacity:%d", p.Cap())
}

// func TestNoPool(t *testing.T) {
// 	var wg sync.WaitGroup
// 	for i := 0; i < n; i++ {
// 		wg.Add(1)
// 		go func() {
// 			demoPoolFunc(n)
// 			wg.Done()
// 		}()
// 	}

// 	wg.Wait()
// 	mem := runtime.MemStats{}
// 	runtime.ReadMemStats(&mem)
// 	t.Logf("memory usage:%d", mem.TotalAlloc/GiB)
// }

//func TestCustomPool(t *testing.T) {
//	p, _ := ants.NewPool(30000)
//	var wg sync.WaitGroup
//	for i := 0; i < n; i++ {
//		wg.Add(1)
//		p.Submit(func() {
//			demoFunc()
//			//demoFunc()
//			wg.Done()
//		})
//	}
//	wg.Wait()
//
//	//t.Logf("pool capacity:%d", p.Cap())
//	//t.Logf("free workers number:%d", p.Free())
//
//	t.Logf("running workers number:%d", p.Running())
//	mem := runtime.MemStats{}
//	runtime.ReadMemStats(&mem)
//	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
//}
