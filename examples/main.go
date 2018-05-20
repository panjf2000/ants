package main

import (
	"fmt"
	"github.com/panjf2000/ants"
	"sync"
)

func myFunc() {
	fmt.Println("Hello World!")
}

func main() {
	//
	runTimes := 10000

	// set 100 the size of goroutine pool
	p, _ := ants.NewPool(100)

	var wg sync.WaitGroup
	// submit
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		p.Push(func() {
			myFunc()
			wg.Done()
		})
	}
	wg.Wait()
	fmt.Println("finish all tasks!")
}
