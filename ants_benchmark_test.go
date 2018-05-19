package ants_test

import (
	"testing"
	"github.com/panjf2000/ants"
)

func BenchmarkPoolGroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ants.Push(demoFunc)
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
		go demoFunc()
	}
}
