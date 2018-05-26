# ants

<div align="center"><img src="https://github.com/panjf2000/ants/blob/master/ants_logo.png"/></div>

<p align="center">A goroutine pool for Go</p>



[![godoc for panjf2000/ants][1]][2]
[![goreportcard for panjf2000/ants][3]][4]
![License](https://img.shields.io/dub/l/vibe-d.svg)

Package ants implements a fixed goroutine pool for managing and recycling a massive number of goroutines, allowing developers to limit the number of goroutines that created by your concurrent programs.

## Features:

- Automatically managing and recycling a massive number of goroutines.
- Friendly interfaces: submitting tasks, getting the number of running goroutines, readjusting capacity of pool dynamically, closing pool.
- Efficient in memory usage and it even achieves higher performance than unlimited goroutines in golang.


## How to install

``` sh
go get -u github.com/panjf2000/ants
```

Or, using glide:

``` sh
glide get github.com/panjf2000/ants
```

## How to use
If your program will generate a massive number of goroutines and you don't want them to consume a vast amount of memory, with ants, all you need to do is to import ants package and submit all your tasks to the default limited pool created when ants was imported:

``` go

package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/ants"
	"time"
)

var sum int32

func myFunc(i interface{}) error {
	n := i.(int)
	atomic.AddInt32(&sum, int32(n))
	fmt.Printf("run with %d\n", n)
	return nil
}

func demoFunc() error {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Hello World!")
	return nil
}

func main() {
	runTimes := 1000

	// use the common pool
	var wg sync.WaitGroup
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		ants.Submit(func() error {
			demoFunc()
			wg.Done()
			return nil
		})
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", ants.Running())
	fmt.Printf("finish all tasks.\n")

	// use the pool with a function
	// set 10 the size of goroutine pool
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) error {
		myFunc(i)
		wg.Done()
		return nil
	})
	// submit tasks
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		p.Serve(i)
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks, result is %d\n", sum)
}
```

## Submit tasks
Tasks can be submitted by calling `ants.Push(func())`
```go
ants.Push(func() {})
```

## Custom limited pool
Ants also supports custom limited pool. You can use the `NewPool` method to create a pool with the given capacity, as following:

``` go
// set 10000 the size of goroutine pool
p, _ := ants.NewPool(10000)
// submit a task
p.Push(func() {})
```

## Readjusting pool capacity
You can change ants pool capacity at any time with `ReSize(int)`:

``` go
pool.ReSize(1000) // Readjust its capacity to 1000
pool.ReSize(100000) // Readjust its capacity to 100000
```

Don't worry about the synchronous problems in this case, this method is thread-safe.

## About sequence
All the tasks submitted to ants pool will not be guaranteed to be processed in order, because those tasks distribute among a series of concurrent workers, thus those tasks are processed concurrently.

## Benchmarks
<div align="center"><img src="https://github.com/panjf2000/ants/blob/master/ants_benchmarks.png"/></div>

 In that benchmark-picture, the first and second benchmarks performed test with 100w tasks and the rest of benchmarks performed test with 1000w tasks, both unlimited goroutines and ants pool, and the capacity of this ants goroutine-pool was limited to 5w.

- BenchmarkGoroutine-4 represent the benchmarks with unlimited goroutines in golang.

- BenchmarkPoolGroutine-4 represent the benchmarks with a ants pool.

The test data above is a basic benchmark and the more detailed benchmarks will be uploaded later.

[1]: https://godoc.org/github.com/panjf2000/ants?status.svg
[2]: https://godoc.org/github.com/panjf2000/ants
[3]: https://goreportcard.com/badge/github.com/panjf2000/ants
[4]: https://goreportcard.com/report/github.com/panjf2000/ants
