package ants_test

import (
	"testing"
	"github.com/panjf2000/ants"
	"sync"
)

const RunTimes = 10000000

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			go func() {
				forSleep()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPoolGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			ants.Push(func() {
				forSleep()
				wg.Done()
			})
		}
		wg.Wait()
	}
}
