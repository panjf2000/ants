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

package main

import (
	"fmt"
	"sync"

	"github.com/panjf2000/ants"
)

var str = "Hello World!"

func myFunc(i interface{}) error {
	s := i.(string)
	fmt.Println(s)
	return nil
}

// func main() {
// 	runTimes := 10000
// 	var wg sync.WaitGroup
// 	// submit all your tasks to ants pool
// 	for i := 0; i < runTimes; i++ {
// 		wg.Add(1)
// 		ants.Push(func() {
// 			myFunc()
// 			wg.Done()
// 		})
// 	}
// 	wg.Wait()
// 	fmt.Println("finish all tasks!")
// }

func main() {
	runTimes := 1000

	// set 100 the size of goroutine pool

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) error {
		myFunc(i)
		wg.Done()
		return nil
	})
	// submit
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		p.Serve(str)
	}
	wg.Wait()
	fmt.Println("finish all tasks!")
}
