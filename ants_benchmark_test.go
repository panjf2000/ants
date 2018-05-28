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
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
	GiB // 1073741824
	TiB // 1099511627776             (超过了int32的范围)
	PiB // 1125899906842624
	EiB // 1152921504606846976
	ZiB // 1180591620717411303424    (超过了int64的范围)
	YiB // 1208925819614629174706176
)
const RunTimes = 1000000
const loop = 1000000

func demoFunc() error {
	n := 10
	time.Sleep(time.Duration(n) * time.Millisecond)
	return nil
}

func demoPoolFunc(args interface{}) error {
	m := args.(int)
	var n int
	for i := 0; i < m; i++ {
		n += i
	}
	return nil
	// n := args.(int)
	// time.Sleep(time.Duration(n) * time.Millisecond)
	// return nil
}

func BenchmarkGoroutineWithFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			go func() {
				demoPoolFunc(loop)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkAntsPoolWithFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		p, _ := ants.NewPoolWithFunc(50000, func(i interface{}) error {
			demoPoolFunc(i)
			wg.Done()
			return nil
		})
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			p.Serve(loop)
		}
		wg.Wait()
		b.Logf("running goroutines: %d", p.Running())
	}
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			go demoPoolFunc(loop)
		}
	}
}

func BenchmarkAntsPool(b *testing.B) {
	p, _ := ants.NewPoolWithFunc(50000, demoPoolFunc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			p.Serve(loop)
		}
		// b.Logf("running goroutines: %d", p.Running())
	}
}
