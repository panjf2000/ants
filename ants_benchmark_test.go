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
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants"
)

const (
	RunTimes      = 10000000
	benchParam    = 10
	benchAntsSize = 100000
)

func demoFunc() error {
	n := 10
	time.Sleep(time.Duration(n) * time.Millisecond)
	return nil
}

func demoPoolFunc(args interface{}) error {
	//m := args.(int)
	//var n int
	//for i := 0; i < m; i++ {
	//	n += i
	//}
	//return nil
	n := args.(int)
	time.Sleep(time.Duration(n) * time.Millisecond)
	return nil
}

func BenchmarkGoroutineWithFunc(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demoPoolFunc(benchParam)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkSemaphoreWithFunc(b *testing.B) {
	var wg sync.WaitGroup
	sema := make(chan struct{}, benchAntsSize)

	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			sema <- struct{}{}
			go func() {
				demoPoolFunc(benchParam)
				<-sema
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkAntsPoolWithFunc(b *testing.B) {
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(benchAntsSize, func(i interface{}) error {
		demoPoolFunc(i)
		wg.Done()
		return nil
	})
	defer p.Release()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			p.Serve(benchParam)
		}
		wg.Wait()
		//b.Logf("running goroutines: %d", p.Running())
	}
	b.StopTimer()
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			go demoPoolFunc(benchParam)
		}
	}
}

func BenchmarkSemaphore(b *testing.B) {
	sema := make(chan struct{}, benchAntsSize)
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			sema <- struct{}{}
			go func() {
				demoPoolFunc(benchParam)
				<-sema
			}()
		}
	}
}

func BenchmarkAntsPool(b *testing.B) {
	p, _ := ants.NewPoolWithFunc(benchAntsSize, demoPoolFunc)
	defer p.Release()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			p.Serve(benchParam)
		}
	}
	b.StopTimer()
}
