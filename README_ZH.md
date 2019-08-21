# ants
<p align="center">
<img src="https://user-images.githubusercontent.com/7496278/51748488-8efd2600-20e7-11e9-91f5-1c5b466dcca1.jpg"/>
A goroutine pool for Go
<br/><br/>
<a title="Build Status" target="_blank" href="https://travis-ci.com/panjf2000/ants"><img src="https://img.shields.io/travis/com/panjf2000/ants?style=flat-square"></a>
<a title="Codecov" target="_blank" href="https://codecov.io/gh/panjf2000/ants"><img src="https://img.shields.io/codecov/c/github/panjf2000/ants?style=flat-square"></a>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/panjf2000/ants"><img src="https://goreportcard.com/badge/github.com/panjf2000/ants?style=flat-square"></a>
<a title="Ants on Sourcegraph" target="_blank" href="https://sourcegraph.com/github.com/panjf2000/ants?badge"><img src="https://sourcegraph.com/github.com/panjf2000/ants/-/badge.svg?style=flat-square"></a>
<br/>
<a title="" target="_blank" href="https://golangci.com/r/github.com/panjf2000/ants"><img src="https://golangci.com/badges/github.com/panjf2000/ants.svg"></a>
<a title="Godoc for ants" target="_blank" href="https://godoc.org/github.com/panjf2000/ants"><img src="https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square"></a>
<a title="Release" target="_blank" href="https://github.com/panjf2000/ants/releases"><img src="https://img.shields.io/github/release/panjf2000/ants.svg?style=flat-square"></a>
<a title="License" target="_blank" href="https://opensource.org/licenses/mit-license.php"><img src="https://img.shields.io/aur/license/pac?style=flat-square"></a>
</p>

[英文](README.md) | [项目博客](https://taohuawu.club/high-performance-implementation-of-goroutine-pool)

`ants`是一个高性能的协程池，实现了对大规模 goroutine 的调度管理、goroutine 复用，允许使用者在开发并发程序的时候限制协程数量，复用资源，达到更高效执行任务的效果。

## 功能：

- 实现了自动调度并发的 goroutine，复用 goroutine
- 定时清理过期的 goroutine，进一步节省资源
- 提供了友好的接口：任务提交、获取运行中的协程数量、动态调整协程池大小
- 优雅处理 panic，防止程序崩溃
- 资源复用，极大节省内存使用量；在大规模批量并发任务场景下比原生 goroutine 并发具有更高的性能

## 目前测试通过的Golang版本：

- 1.8.x
- 1.9.x
- 1.10.x
- 1.11.x
- 1.12.x


## 安装

``` sh
go get -u github.com/panjf2000/ants
```

## 使用
写 go 并发程序的时候如果程序会启动大量的 goroutine ，势必会消耗大量的系统资源（内存，CPU），通过使用 `ants`，可以实例化一个协程池，复用 goroutine ，节省资源，提升性能：

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
		_ = ants.Submit(syncCalculateSum)
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
		_ = p.Invoke(int32(i))
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks, result is %d\n", sum)
}
```

## 与 http server 集成
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
  pool, _ := ants.NewPoolWithFunc(100000, func(payload interface{}) {
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

## Pool 配置

```go
type Options struct {
	// ExpiryDuration set the expired time (second) of every worker.
	ExpiryDuration time.Duration

	// PreAlloc indicate whether to make memory pre-allocation when initializing Pool.
	PreAlloc bool

	// Max number of goroutine blocking on pool.Submit.
	// 0 (default value) means no such limit.
	MaxBlockingTasks int

	// When Nonblocking is true, Pool.Submit will never be blocked.
	// ErrPoolOverload will be returned when Pool.Submit cannot be done at once.
	// When Nonblocking is true, MaxBlockingTasks is inoperative.
	Nonblocking bool

	// PanicHandler is used to handle panics from each worker goroutine.
	// if nil, panics will be thrown out again from worker goroutines.
	PanicHandler func(interface{})
}

func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

func WithExpiryDuration(expiryDuration time.Duration) Option {
	return func(opts *Options) {
		opts.ExpiryDuration = expiryDuration
	}
}

func WithPreAlloc(preAlloc bool) Option {
	return func(opts *Options) {
		opts.PreAlloc = preAlloc
	}
}

func WithMaxBlockingTasks(maxBlockingTasks int) Option {
	return func(opts *Options) {
		opts.MaxBlockingTasks = maxBlockingTasks
	}
}

func WithNonblocking(nonblocking bool) Option {
	return func(opts *Options) {
		opts.Nonblocking = nonblocking
	}
}

func WithPanicHandler(panicHandler func(interface{})) Option {
	return func(opts *Options) {
		opts.PanicHandler = panicHandler
	}
}
```

通过在调用`NewPool`/`NewPoolWithFunc`之时使用各种 optional function，可以设置`ants.Options`中各个配置项的值，然后用它来定制化 goroutine pool.


## 自定义池
`ants`支持实例化使用者自己的一个 Pool ，指定具体的池容量；通过调用 `NewPool` 方法可以实例化一个新的带有指定容量的 Pool ，如下：

``` go
// Set 10000 the size of goroutine pool
p, _ := ants.NewPool(10000)
```

## 任务提交

提交任务通过调用 `ants.Submit(func())`方法：
```go
ants.Submit(func(){})
```

## 动态调整协程池容量
需要动态调整协程池容量可以通过调用`Tune(int)`：

``` go
pool.Tune(1000) // Tune its capacity to 1000
pool.Tune(100000) // Tune its capacity to 100000
```

该方法是线程安全的。

## 预先分配 goroutine 队列内存

`ants`允许你预先把整个池的容量分配内存， 这个功能可以在某些特定的场景下提高协程池的性能。比如， 有一个场景需要一个超大容量的池，而且每个 goroutine 里面的任务都是耗时任务，这种情况下，预先分配 goroutine 队列内存将会减少 re-slice 时的复制内存损耗。

```go
// ants will pre-malloc the whole capacity of pool when you invoke this function
p, _ := ants.NewPool(100000, ants.WithPreAlloc(true))
```



## 销毁协程池

```go
pool.Release()
```

## Benchmarks

系统参数：

```
OS: macOS High Sierra
Processor: 2.7 GHz Intel Core i5
Memory: 8 GB 1867 MHz DDR3

Go Version: 1.9
```



<div align="center"><img src="https://user-images.githubusercontent.com/7496278/51515466-c7ce9e00-1e4e-11e9-89c4-bd3785b3c667.png"/></div>
上图中的前两个 benchmark 测试结果是基于100w 任务量的条件，剩下的几个是基于 1000w 任务量的测试结果，`ants`的默认池容量是 5w。

- BenchmarkGoroutine-4 代表原生 goroutine

- BenchmarkPoolGroutine-4 代表使用协程池`ants`

### Benchmarks with Pool 

![](https://user-images.githubusercontent.com/7496278/51515499-f187c500-1e4e-11e9-80e5-3df8f94fa70f.png)

**这里为了模拟大规模 goroutine 的场景，两次测试的并发次数分别是 100w 和 1000w，前两个测试分别是执行 100w 个并发任务不使用 Pool 和使用了`ants`的 Goroutine Pool 的性能，后两个则是 1000w 个任务下的表现，可以直观的看出在执行速度和内存使用上，`ants`的 Pool 都占有明显的优势。100w 的任务量，使用`ants`，执行速度与原生 goroutine 相当甚至略快，但只实际使用了不到 5w 个 goroutine 完成了全部任务，且内存消耗仅为原生并发的 40%；而当任务量达到 1000w，优势则更加明显了：用了 70w 左右的 goroutine 完成全部任务，执行速度比原生 goroutine 提高了 100%，且内存消耗依旧保持在不使用 Pool 的 40% 左右。**

### Benchmarks with PoolWithFunc

![](https://user-images.githubusercontent.com/7496278/51515565-1e3bdc80-1e4f-11e9-8a08-452ab91d117e.png)

**因为`PoolWithFunc`这个 Pool 只绑定一个任务函数，也即所有任务都是运行同一个函数，所以相较于`Pool`对原生 goroutine 在执行速度和内存消耗的优势更大，上面的结果可以看出，执行速度可以达到原生 goroutine 的 300%，而内存消耗的优势已经达到了两位数的差距，原生 goroutine 的内存消耗达到了`ants`的35倍且原生 goroutine 的每次执行的内存分配次数也达到了`ants`45倍，1000w 的任务量，`ants`的初始分配容量是 5w，因此它完成了所有的任务依旧只使用了 5w 个 goroutine！事实上，`ants`的 Goroutine Pool 的容量是可以自定义的，也就是说使用者可以根据不同场景对这个参数进行调优直至达到最高性能。**

### 吞吐量测试（适用于那种只管提交异步任务而无须关心结果的场景）

#### 10w 任务量

![](https://user-images.githubusercontent.com/7496278/51515590-36abf700-1e4f-11e9-91e4-7bd3dcb5f4a5.png)

#### 100w 任务量

![](https://user-images.githubusercontent.com/7496278/51515596-44617c80-1e4f-11e9-89e3-01e19d2979a1.png)

#### 1000w 任务量

![](https://user-images.githubusercontent.com/7496278/52987732-537c2000-3437-11e9-86a6-177f00d7a1d6.png)

### 性能小结

![](https://user-images.githubusercontent.com/7496278/52989641-51b65a80-343f-11e9-86c0-e855d97343ea.gif)

**从该 demo 测试吞吐性能对比可以看出，使用`ants`的吞吐性能相较于原生 goroutine 可以保持在 2-6 倍的性能压制，而内存消耗则可以达到 10-20 倍的节省优势。** 