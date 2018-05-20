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
	time.Sleep(3 * time.Millisecond)
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
