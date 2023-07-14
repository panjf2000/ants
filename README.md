<p align="center">
<img src="https://raw.githubusercontent.com/panjf2000/logos/master/ants/logo.png" />
<b>A goroutine pool for Go</b>
<br/><br/>
<a title="Build Status" target="_blank" href="https://github.com/panjf2000/ants/actions?query=workflow%3ATests"><img src="https://img.shields.io/github/actions/workflow/status/panjf2000/ants/test.yml?branch=master&style=flat-square&logo=github-actions" /></a>
<a title="Codecov" target="_blank" href="https://codecov.io/gh/panjf2000/ants"><img src="https://img.shields.io/codecov/c/github/panjf2000/ants?style=flat-square&logo=codecov" /></a>
<a title="Release" target="_blank" href="https://github.com/panjf2000/ants/releases"><img src="https://img.shields.io/github/v/release/panjf2000/ants.svg?color=161823&style=flat-square&logo=smartthings" /></a>
<a title="Tag" target="_blank" href="https://github.com/panjf2000/ants/tags"><img src="https://img.shields.io/github/v/tag/panjf2000/ants?color=%23ff8936&logo=fitbit&style=flat-square" /></a>
<br/>
<a title="Chat Room" target="_blank" href="https://gitter.im/ants-pool/ants?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=body_badge"><img src="https://badges.gitter.im/ants-pool/ants.svg" /></a>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/panjf2000/ants"><img src="https://goreportcard.com/badge/github.com/panjf2000/ants?style=flat-square" /></a>
<a title="Doc for ants" target="_blank" href="https://pkg.go.dev/github.com/panjf2000/ants/v2?tab=doc"><img src="https://img.shields.io/badge/go.dev-doc-007d9c?style=flat-square&logo=read-the-docs" /></a>
<a title="Mentioned in Awesome Go" target="_blank" href="https://github.com/avelino/awesome-go#goroutines"><img src="https://awesome.re/mentioned-badge-flat.svg" /></a>
</p>

English | [中文](README_ZH.md)

## 📖 Introduction

Library `ants` implements a goroutine pool with fixed capacity, managing and recycling a massive number of goroutines, allowing developers to limit the number of goroutines in your concurrent programs.

## 🚀 Features:

- Managing and recycling a massive number of goroutines automatically
- Purging overdue goroutines periodically
- Abundant APIs: submitting tasks, getting the number of running goroutines, tuning the capacity of the pool dynamically, releasing the pool, rebooting the pool
- Handle panic gracefully to prevent programs from crash
- Efficient in memory usage and it even achieves [higher performance](#-performance-summary) than unlimited goroutines in Golang
- Nonblocking mechanism

## 💡 How `ants` works

### Flow Diagram

<p align="center">
<img width="1011" alt="ants-flowchart-en" src="https://user-images.githubusercontent.com/7496278/66396509-7b42e700-ea0c-11e9-8612-b71a4b734683.png">
</p>

### Activity Diagrams

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-1.png)

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-2.png)

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-3.png)

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-4.png)

## 🧰 How to install

### For `ants` v1

``` powershell
go get -u github.com/panjf2000/ants
```

### For `ants` v2 (with GO111MODULE=on)

```powershell
go get -u github.com/panjf2000/ants/v2
```

## 🛠 How to use
Just imagine that your program starts a massive number of goroutines, resulting in a huge consumption of memory. To mitigate that kind of situation, all you need to do is to import `ants` package and submit all your tasks to a default pool with fixed capacity, activated when package `ants` is imported:

``` go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
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

###  Functional options for ants pool

```go
// Option represents the optional function.
type Option func(opts *Options)

// Options contains all options which will be applied when instantiating a ants pool.
type Options struct {
	// ExpiryDuration is a period for the scavenger goroutine to clean up those expired workers,
	// the scavenger scans all workers every `ExpiryDuration` and clean up those workers that haven't been
	// used for more than `ExpiryDuration`.
	ExpiryDuration time.Duration

	// PreAlloc indicates whether to make memory pre-allocation when initializing Pool.
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

	// Logger is the customized logger for logging info, if it is not set,
	// default standard logger from log package is used.
	Logger Logger
}

// WithOptions accepts the whole options config.
func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

// WithExpiryDuration sets up the interval time of cleaning up goroutines.
func WithExpiryDuration(expiryDuration time.Duration) Option {
	return func(opts *Options) {
		opts.ExpiryDuration = expiryDuration
	}
}

// WithPreAlloc indicates whether it should malloc for workers.
func WithPreAlloc(preAlloc bool) Option {
	return func(opts *Options) {
		opts.PreAlloc = preAlloc
	}
}

// WithMaxBlockingTasks sets up the maximum number of goroutines that are blocked when it reaches the capacity of pool.
func WithMaxBlockingTasks(maxBlockingTasks int) Option {
	return func(opts *Options) {
		opts.MaxBlockingTasks = maxBlockingTasks
	}
}

// WithNonblocking indicates that pool will return nil when there is no available workers.
func WithNonblocking(nonblocking bool) Option {
	return func(opts *Options) {
		opts.Nonblocking = nonblocking
	}
}

// WithPanicHandler sets up panic handler.
func WithPanicHandler(panicHandler func(interface{})) Option {
	return func(opts *Options) {
		opts.PanicHandler = panicHandler
	}
}

// WithLogger sets up a customized logger.
func WithLogger(logger Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}
```

`ants.Options`contains all optional configurations of the ants pool, which allows you to customize the goroutine pool by invoking option functions to set up each configuration in `NewPool`/`NewPoolWithFunc`method.

### Customize limited pool

`ants` also supports customizing the capacity of the pool. You can invoke the `NewPool` method to instantiate a pool with a given capacity, as follows:

``` go
p, _ := ants.NewPool(10000)
```

### Submit tasks
Tasks can be submitted by calling `ants.Submit(func())`
```go
ants.Submit(func(){})
```

### Tune pool capacity in runtime
You can tune the capacity of  `ants` pool in runtime with `Tune(int)`:

``` go
pool.Tune(1000) // Tune its capacity to 1000
pool.Tune(100000) // Tune its capacity to 100000
```

Don't worry about the contention problems in this case, the method here is thread-safe (or should be called goroutine-safe).

### Pre-malloc goroutine queue in pool

`ants` allows you to pre-allocate the memory of the goroutine queue in the pool, which may get a performance enhancement under some special certain circumstances such as the scenario that requires a pool with ultra-large capacity, meanwhile, each task in goroutine lasts for a long time, in this case, pre-mallocing will reduce a lot of memory allocation in goroutine queue.

```go
// ants will pre-malloc the whole capacity of pool when you invoke this method
p, _ := ants.NewPool(100000, ants.WithPreAlloc(true))
```

### Release Pool

```go
pool.Release()
```

### Reboot Pool

```go
// A pool that has been released can be still used once you invoke the Reboot().
pool.Reboot()
```

## ⚙️ About sequence

All tasks submitted to `ants` pool will not be guaranteed to be addressed in order, because those tasks scatter among a series of concurrent workers, thus those tasks would be executed concurrently.

## 🧲 Benchmarks

<div align="center"><img src="https://user-images.githubusercontent.com/7496278/51515466-c7ce9e00-1e4e-11e9-89c4-bd3785b3c667.png"/></div>
 In this benchmark result, the first and second benchmarks performed test cases with 1M tasks, and the rest of the benchmarks performed test cases with 10M tasks, both in unlimited goroutines and `ants` pool, and the capacity of this `ants` goroutine pool was limited to 50K.

- BenchmarkGoroutine-4 represents the benchmarks with unlimited goroutines in Golang.

- BenchmarkPoolGroutine-4 represents the benchmarks with an `ants` pool.

### Benchmarks with Pool 

![](https://user-images.githubusercontent.com/7496278/51515499-f187c500-1e4e-11e9-80e5-3df8f94fa70f.png)

In the above benchmark result, the first and second benchmarks performed test cases with 1M tasks, and the rest of the benchmarks performed test cases with 10M tasks, both in unlimited goroutines and `ants` pool and the capacity of this `ants` goroutine-pool was limited to 50K.

**As you can see, `ants` performs 2 times faster than goroutines without a pool (10M tasks) and it only consumes half the memory compared with goroutines without a pool. (both in 1M and 10M tasks)**

### Benchmarks with PoolWithFunc

![](https://user-images.githubusercontent.com/7496278/51515565-1e3bdc80-1e4f-11e9-8a08-452ab91d117e.png)

### Throughput (it is suitable for scenarios where tasks are submitted asynchronously without waiting for the final results) 

#### 100K tasks

![](https://user-images.githubusercontent.com/7496278/51515590-36abf700-1e4f-11e9-91e4-7bd3dcb5f4a5.png)

#### 1M tasks

![](https://user-images.githubusercontent.com/7496278/51515596-44617c80-1e4f-11e9-89e3-01e19d2979a1.png)

#### 10M tasks

![](https://user-images.githubusercontent.com/7496278/52987732-537c2000-3437-11e9-86a6-177f00d7a1d6.png)

## 📊 Performance Summary

![](https://user-images.githubusercontent.com/7496278/63449727-3ae6d400-c473-11e9-81e3-8b3280d8288a.gif)

**In conclusion, `ants` performs 2~6 times faster than goroutines without a pool and the memory consumption is reduced by 10 to 20 times.**

## 👏 Contributors

Please read our [Contributing Guidelines](CONTRIBUTING.md) before opening a PR and thank you to all the developers who already made contributions to `ants`!

<a href="https://github.com/panjf2000/ants/graphs/contributors">
	<img src="https://contrib.rocks/image?repo=panjf2000/ants" />
</a>

## 📄 License

The source code in `ants` is available under the [MIT License](/LICENSE).

## 📚 Relevant Articles

-  [Goroutine 并发调度模型深度解析之手撸一个高性能 goroutine 池](https://taohuawu.club/high-performance-implementation-of-goroutine-pool)
-  [Visually Understanding Worker Pool](https://medium.com/coinmonks/visually-understanding-worker-pool-48a83b7fc1f5)
-  [The Case For A Go Worker Pool](https://brandur.org/go-worker-pool)
-  [Go Concurrency - GoRoutines, Worker Pools and Throttling Made Simple](https://twin.sh/articles/39/go-concurrency-goroutines-worker-pools-and-throttling-made-simple)

## 🖥 Use cases

### business companies

The following companies/organizations use `ants` in production.

<a href="https://www.tencent.com"><img src="http://img.taohuawu.club/gallery/tencent_logo.png" width="250" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.bytedance.com/" target="_blank"><img src="http://img.taohuawu.club/gallery/ByteDance_Logo.png" width="250" align="middle"/></a>&nbsp;&nbsp;<a href="https://tieba.baidu.com/" target="_blank"><img src="http://img.taohuawu.club/gallery/baidu-tieba-logo.png" width="300" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.sina.com.cn/" target="_blank"><img src="http://img.taohuawu.club/gallery/sina-logo.png" width="200" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.163.com/" target="_blank"><img src="http://img.taohuawu.club/gallery/netease-logo.png" width="150" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.tencentmusic.com/" target="_blank"><img src="http://img.taohuawu.club/gallery/tencent-music-logo.png" width="250" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.futuhk.com/" target="_blank"><img src="http://img.taohuawu.club/gallery/futu-logo.png" width="250" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.shopify.com/" target="_blank"><img src="http://img.taohuawu.club/gallery/shopify-logo.png" width="250" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.wechat.com/en/" target="_blank"><img src="http://img.taohuawu.club/gallery/wechat-logo.png" width="250" align="middle"/></a><a href="https://www.baidu.com/" target="_blank"><img src="http://img.taohuawu.club/gallery/baidu-mobile.png" width="250" align="middle"/></a>&nbsp;&nbsp;<a href="https://www.360.com" target="_blank"><img src="http://img.taohuawu.club/gallery/360-logo.png" width="250" align="middle"/></a>

### open-source software

- [gnet](https://github.com/panjf2000/gnet):  A high-performance, lightweight, non-blocking, event-driven networking framework written in pure Go.
- [nps](https://github.com/ehang-io/nps): A lightweight, high-performance, powerful intranet penetration proxy server, with a powerful web management terminal.
- [milvus](https://github.com/milvus-io/milvus): An open-source vector database for scalable similarity search and AI applications.
- [siyuan](https://github.com/siyuan-note/siyuan): SiYuan is a local-first personal knowledge management system that supports complete offline use, as well as end-to-end encrypted synchronization.
- [osmedeus](https://github.com/j3ssie/osmedeus): A Workflow Engine for Offensive Security.
- [jitsu](https://github.com/jitsucom/jitsu): An open-source Segment alternative. Fully-scriptable data ingestion engine for modern data teams. Set-up a real-time data pipeline in minutes, not days.
- [triangula](https://github.com/RH12503/triangula): Generate high-quality triangulated and polygonal art from images.
- [teler](https://github.com/kitabisa/teler): Real-time HTTP Intrusion Detection.
- [bsc](https://github.com/binance-chain/bsc): A Binance Smart Chain client based on the go-ethereum fork.
- [jaeles](https://github.com/jaeles-project/jaeles): The Swiss Army knife for automated Web Application Testing.
- [devlake](https://github.com/apache/incubator-devlake): The open-source dev data platform & dashboard for your DevOps tools.

#### All use cases:

- [Repositories that depend on ants/v2](https://github.com/panjf2000/ants/network/dependents?package_id=UGFja2FnZS0yMjY2ODgxMjg2)

- [Repositories that depend on ants/v1](https://github.com/panjf2000/ants/network/dependents?package_id=UGFja2FnZS0yMjY0ODMzNjEw)

If you have `ants` integrated into projects, feel free to open a pull request refreshing this list of use cases.

## 🔋 JetBrains OS licenses

`ants` had been being developed with GoLand under the **free JetBrains Open Source license(s)** granted by JetBrains s.r.o., hence I would like to express my thanks here.

<a href="https://www.jetbrains.com/?from=ants" target="_blank"><img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/jetbrains/jetbrains-variant-4.png" width="250" align="middle"/></a>

## 💰 Backers

Support us with a monthly donation and help us continue our activities.

<a href="https://opencollective.com/ants#backers" target="_blank"><img src="https://opencollective.com/ants/backers.svg"></a>

## 💎 Sponsors

Become a bronze sponsor with a monthly donation of $10 and get your logo on our README on GitHub.

<a href="https://opencollective.com/ants#sponsors" target="_blank"><img src="https://opencollective.com/ants/sponsors.svg"></a>

## ☕️ Buy me a coffee

> Please be sure to leave your name, GitHub account, or other social media accounts when you donate by the following means so that I can add it to the list of donors as a token of my appreciation.

<img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/payments/WeChatPay.JPG" width="250" align="middle"/>&nbsp;&nbsp;
<img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/payments/AliPay.JPG" width="250" align="middle"/>&nbsp;&nbsp;
<a href="https://www.paypal.me/R136a1X" target="_blank"><img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/payments/PayPal.JPG" width="250" align="middle"/></a>&nbsp;&nbsp;

## 💵 Patrons

<a target="_blank" href="https://github.com/patrick-othmer"><img src="https://avatars1.githubusercontent.com/u/8964313" width="100" alt="Patrick Othmer" /></a>&nbsp;<a target="_blank" href="https://github.com/panjf2000/ants"><img src="https://avatars2.githubusercontent.com/u/50285334" width="100" alt="Jimmy" /></a>&nbsp;<a target="_blank" href="https://github.com/cafra"><img src="https://avatars0.githubusercontent.com/u/13758306" width="100" alt="ChenZhen" /></a>&nbsp;<a target="_blank" href="https://github.com/yangwenmai"><img src="https://avatars0.githubusercontent.com/u/1710912" width="100" alt="Mai Yang" /></a>&nbsp;<a target="_blank" href="https://github.com/BeijingWks"><img src="https://avatars3.githubusercontent.com/u/33656339" width="100" alt="王开帅" /></a>&nbsp;<a target="_blank" href="https://github.com/refs"><img src="https://avatars3.githubusercontent.com/u/6905948" width="100" alt="Unger Alejandro" /></a>&nbsp;<a target="_blank" href="https://github.com/Wuvist"><img src="https://avatars.githubusercontent.com/u/657796" width="100" alt="Weng Wei" /></a>

## 🔋 Sponsorship

<p>
  <a href="https://www.digitalocean.com/">
    <img src="https://opensource.nyc3.cdn.digitaloceanspaces.com/attribution/assets/PoweredByDO/DO_Powered_by_Badge_blue.svg" width="201px">
  </a>
</p>
