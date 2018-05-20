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
			wg.Done()
		})
	}
	wg.Wait()

	//t.Logf("pool capacity:%d", ants.Cap())
	//t.Logf("running workers number:%d", ants.Running())
	//t.Logf("free workers number:%d", ants.Free())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Logf("memory usage:%d", mem.TotalAlloc/1024)
}

//func TestCustomPool(t *testing.T) {
//	p := ants.NewPool(1000)
//	for i := 0; i < n; i++ {
//		p.Push(demoFunc)
//	}
//
//	t.Logf("pool capacity:%d", p.Cap())
//	t.Logf("running workers number:%d", p.Running())
//	t.Logf("free workers number:%d", p.Free())
//
//	mem := runtime.MemStats{}
//	runtime.ReadMemStats(&mem)
//
//}
