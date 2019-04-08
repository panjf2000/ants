# ants

<div align="center"><img src="https://user-images.githubusercontent.com/7496278/51748488-8efd2600-20e7-11e9-91f5-1c5b466dcca1.jpg"/></div>

<p align="center">A goroutine pool for Go</p>

[![Build Status][1]][2]
[![codecov][3]][4]
[![goreportcard for panjf2000/ants][5]][6]
[![godoc for panjf2000/ants][7]][8]
[![MIT Licence][9]][10]

[中文](README_ZH.md) | [Project Tutorial](http://blog.taohuawu.club/article/goroutine-pool)

Library `ants` implements a goroutine pool with fixed capacity, managing and recycling a massive number of goroutines, allowing developers to limit the number of goroutines in your concurrent programs.

## Features:

- Automatically managing and recycling a massive number of goroutines.
- Periodically purging overdue goroutines.
- Friendly interfaces: submitting tasks, getting the number of running goroutines, readjusting capacity of pool dynamically, closing pool.
- Handle panic gracefully to prevent programs from crash.
- Efficient in memory usage and it even achieves higher performance than unlimited goroutines in golang.

## Tested in the following Golang versions:

- 1.8.x
- 1.9.x
- 1.10.x
- 1.11.x
- 1.12.x


## How to install

``` sh
go get -u github.com/panjf2000/ants
```

Or, using glide:

``` sh
glide get github.com/panjf2000/ants
```

## How to use
Just take a imagination that your program starts a massive number of goroutines, from which a vast amount of memory will be consumed. To mitigate that kind of situation, all you need to do is to import `ants` package and submit all your tasks to a default pool with fixed capacity activated when package `ants` has been imported:

``` go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants"
)

var sum int32

func myFunc(i interface{}) {
	n := i.(int32)
	atomic.AddInt32(&sum, n)
	fmt.Printf("run with %d\n", n)
}

func demoFunc() {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Hello World!")
}

func main() {
	defer ants.Release()

	runTimes := 1000

	// Use the common pool.
	var wg sync.WaitGroup
	syncCalculateSum := func() {
		demoFunc()
		wg.Done()
	}
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		ants.Submit(syncCalculateSum)
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", ants.Running())
	fmt.Printf("finish all tasks.\n")

	// Use the pool with a function,
	// set 10 to the capacity of goroutine pool and 1 second for expired duration.
	p, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
		myFunc(i)
		wg.Done()
	})
	defer p.Release()
	// Submit tasks one by one.
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		p.Invoke(int32(i))
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks, result is %d\n", sum)
}
```

## Integrate with http server
```go
package main

import (
	"io/ioutil"
	"net/http"

	"github.com/panjf2000/ants"
)

type Request struct {
	Param  []byte
	Result chan []byte
}

func main() {
	pool, _ := ants.NewPoolWithFunc(100, func(payload interface{}) {
		request, ok := payload.(*Request)
		if !ok {
			return
		}
		reverseParam := func(s []byte) []byte {
			for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
				s[i], s[j] = s[j], s[i]
			}
			return s
		}(request.Param)

		request.Result <- reverseParam
	})
	defer pool.Release()

	http.HandleFunc("/reverse", func(w http.ResponseWriter, r *http.Request) {
		param, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "request error", http.StatusInternalServerError)
		}
		defer r.Body.Close()

		request := &Request{Param: param, Result: make(chan []byte)}

		// Throttle the requests traffic with ants pool. This process is asynchronous and
		// you can receive a result from the channel defined outside.
		if err := pool.Invoke(request); err != nil {
			http.Error(w, "throttle limit error", http.StatusInternalServerError)
		}

		w.Write(<-request.Result)
	})

	http.ListenAndServe(":8080", nil)
}
```

## Submit tasks
Tasks can be submitted by calling `ants.Submit(func())`
```go
ants.Submit(func(){})
```

## Customize limited pool
`ants` also supports customizing the capacity of pool. You can invoke the `NewPool` function to instantiate a pool with a given capacity, as following:

``` go
// Set 10000 the size of goroutine pool
p, _ := ants.NewPool(10000)
// Submit a task
p.Submit(func(){})
```

## Tune pool capacity
You can tune the capacity of  `ants` pool at any time with `Tune(int)`:

``` go
pool.Tune(1000) // Tune its capacity to 1000
pool.Tune(100000) // Tune its capacity to 100000
```

Don't worry about the synchronous problems in this case, the function here is thread-safe (or should be called goroutine-safe).

## Release Pool

```go
pool.Release()
```



## About sequence
All tasks submitted to `ants` pool will not be guaranteed to be addressed in order, because those tasks scatter among a series of concurrent workers, thus those tasks are executed concurrently.

## Benchmarks

```
OS: macOS High Sierra
Processor: 2.7 GHz Intel Core i5
Memory: 8 GB 1867 MHz DDR3

Go Version: 1.9
```

<div align="center"><img src="https://user-images.githubusercontent.com/7496278/51515466-c7ce9e00-1e4e-11e9-89c4-bd3785b3c667.png"/></div>

 In that benchmark-picture, the first and second benchmarks performed test cases with 1M tasks and the rest of benchmarks performed test cases with 10M tasks, both in unlimited goroutines and `ants` pool, and the capacity of this `ants` goroutine-pool was limited to 50K.

- BenchmarkGoroutine-4 represents the benchmarks with unlimited goroutines in golang.

- BenchmarkPoolGroutine-4 represents the benchmarks with a `ants` pool.

The test data above is a basic benchmark and more detailed benchmarks are about to be uploaded later.

### Benchmarks with Pool 

![](https://user-images.githubusercontent.com/7496278/51515499-f187c500-1e4e-11e9-80e5-3df8f94fa70f.png)

In above benchmark picture, the first and second benchmarks performed test cases with 1M tasks and the rest of benchmarks performed test cases with 10M tasks, both in unlimited goroutines and `ants` pool, and the capacity of this `ants` goroutine-pool was limited to 50K.

**As you can see, `ants` can up to 2x faster than goroutines without pool (10M tasks) and it only consumes half the memory comparing with goroutines without pool. (both 1M and 10M tasks)**

### Benchmarks with PoolWithFunc

![](https://user-images.githubusercontent.com/7496278/51515565-1e3bdc80-1e4f-11e9-8a08-452ab91d117e.png)

### Throughput (it is suitable for scenarios where asynchronous tasks are submitted despite of the final results) 

#### 100K tasks

![](https://user-images.githubusercontent.com/7496278/51515590-36abf700-1e4f-11e9-91e4-7bd3dcb5f4a5.png)

#### 1M tasks

![](https://user-images.githubusercontent.com/7496278/51515596-44617c80-1e4f-11e9-89e3-01e19d2979a1.png)

#### 10M tasks

![](https://user-images.githubusercontent.com/7496278/52987732-537c2000-3437-11e9-86a6-177f00d7a1d6.png)

### Performance Summary

![](https://user-images.githubusercontent.com/7496278/52989641-51b65a80-343f-11e9-86c0-e855d97343ea.gif)

**In conclusion, `ants` can up to 2x~6x faster than goroutines without a pool and the memory consumption is reduced by 10 to 20 times.**

[1]: https://travis-ci.com/panjf2000/ants.svg?branch=master
[2]: https://travis-ci.com/panjf2000/ants
[3]: https://codecov.io/gh/panjf2000/ants/branch/master/graph/badge.svg
[4]: https://codecov.io/gh/panjf2000/ants
[5]: https://goreportcard.com/badge/github.com/panjf2000/ants
[6]: https://goreportcard.com/report/github.com/panjf2000/ants
[7]: https://godoc.org/github.com/panjf2000/ants?status.svg
[8]: https://godoc.org/github.com/panjf2000/ants
[9]: https://badges.frapsoft.com/os/mit/mit.svg?v=103
[10]: https://opensource.org/licenses/mit-license.php
