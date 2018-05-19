package ants_test

import (
	"testing"
	"github.com/panjf2000/ants"
	"sync"
)

const RunTimes = 100000

func BenchmarkPoolGroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			ants.Push(demoFunc)
		}
		ants.Wait()
	}
}

//func BenchmarkPoolGroutine(b *testing.B) {
//	p := ants.NewPool(size)
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		p.Push(demoFunc)
//	}
//}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				demoFunc()
			}()
		}
		wg.Wait()
	}
}
