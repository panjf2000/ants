# ants

<div align="center"><img src="ants_logo.png"/></div>

<p align="center">A goroutine pool for Go</p>



[![godoc for panjf2000/ants][1]][2] [![goreportcard for panjf2000/ants][3]][4] [![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)

[英文说明页](README.md) | [项目介绍文章传送门](http://blog.taohuawu.club/article/42)

`ants`是一个高性能的协程池，实现了对大规模goroutine的调度管理、goroutine复用，允许使用者在开发并发程序的时候限制协程数量，复用资源，达到更高效执行任务的效果。

## 功能:

- 实现了自动调度并发的goroutine，复用goroutine
- 提供了友好的接口：任务提交、获取运行中的协程数量、动态调整协程池大小
- 资源复用，极大节省内存使用量；在大规模批量并发任务场景下比原生goroutine并发具有更高的性能


## 安装

``` sh
go get -u github.com/panjf2000/ants
```

使用包管理工具 glide 安装:

``` sh
glide get github.com/panjf2000/ants
```

## 使用
写 go 并发程序的时候如果程序会启动大量的 goroutine ，势必会消耗大量的系统资源（内存，CPU），通过使用 `ants`，可以实例化一个协程池，复用 goroutine ，节省资源，提升性能：

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

## 任务提交
提交任务通过调用 `ants.Submit(func())`方法：
```go
ants.Submit(func() {})
```

## 自定义池
`ants`支持实例化使用者自己的一个 Pool ，指定具体的池容量；通过调用 `NewPool` 方法可以实例化一个新的带有指定容量的 Pool ，如下：

``` go
// set 10000 the size of goroutine pool
p, _ := ants.NewPool(10000)
// submit a task
p.Submit(func() {})
```

## 动态调整协程池容量
需要动态调整协程池容量可以通过调用`ReSize(int)`：

``` go
pool.ReSize(1000) // Readjust its capacity to 1000
pool.ReSize(100000) // Readjust its capacity to 100000
```

该方法是线程安全的。

## Benchmarks

系统参数：

```
OS : macOS High Sierra
Processor : 2.7 GHz Intel Core i5
Memory : 8 GB 1867 MHz DDR3
```



<div align="center"><img src="ants_benchmarks.png"/></div>

上图中的前两个 benchmark 测试结果是基于100w任务量的条件，剩下的几个是基于1000w任务量的测试结果，`ants`的默认池容量是5w。

- BenchmarkGoroutine-4 代表原生goroutine

- BenchmarkPoolGroutine-4 代表使用协程池`ants`

### Benchmarks with Pool 

![](benchmark_pool.png)



### Benchmarks with PoolWithFunc

![](ants_bench_poolwithfunc.png)

### 吞吐量测试

#### 10w 任务量

![](ants_bench_10w.png)

#### 100w 任务量

![](ants_bench_100w.png)

#### 1000w 任务量

![](ants_bench_1000w.png)

1000w任务量的场景下，我的电脑已经无法支撑 golang 的原生 goroutine 并发，所以只测出了使用`ants`池的测试结果。

[1]: https://godoc.org/github.com/panjf2000/ants?status.svg
[2]: https://godoc.org/github.com/panjf2000/ants
[3]: https://goreportcard.com/badge/github.com/panjf2000/ants
[4]: https://goreportcard.com/report/github.com/panjf2000/ants
