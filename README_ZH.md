<p align="center">
<img src="https://raw.githubusercontent.com/panjf2000/logos/master/ants/logo.png" />
<b>Go 语言的 goroutine 池</b>
<br/><br/>
<a title="Build Status" target="_blank" href="https://travis-ci.com/panjf2000/ants"><img src="https://img.shields.io/travis/com/panjf2000/ants?style=flat-square&logo=travis" /></a>
<a title="Codecov" target="_blank" href="https://codecov.io/gh/panjf2000/ants"><img src="https://img.shields.io/codecov/c/github/panjf2000/ants?style=flat-square&logo=codecov" /></a>
<a title="Release" target="_blank" href="https://github.com/panjf2000/ants/releases"><img src="https://img.shields.io/github/v/release/panjf2000/ants.svg?color=161823&style=flat-square&logo=smartthings" /></a>
<a title="Tag" target="_blank" href="https://github.com/panjf2000/ants/tags"><img src="https://img.shields.io/github/v/tag/panjf2000/ants?color=%23ff8936&logo=fitbit&style=flat-square" /></a>
<br/>
<a title="Chat Room" target="_blank" href="https://gitter.im/ants-pool/ants?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=body_badge"><img src="https://badges.gitter.im/ants-pool/ants.svg" /></a>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/panjf2000/ants"><img src="https://goreportcard.com/badge/github.com/panjf2000/ants?style=flat-square" /></a>
<a title="Doc for ants" target="_blank" href="https://pkg.go.dev/github.com/panjf2000/ants/v2?tab=doc"><img src="https://img.shields.io/badge/go.dev-doc-007d9c?style=flat-square&logo=read-the-docs" /></a>
<a title="Ants on Sourcegraph" target="_blank" href="https://sourcegraph.com/github.com/panjf2000/ants?badge"><img src="https://sourcegraph.com/github.com/panjf2000/ants/-/badge.svg?style=flat-square" /></a>
<a title="Mentioned in Awesome Go" target="_blank" href="https://github.com/avelino/awesome-go#goroutines"><img src="https://awesome.re/mentioned-badge-flat.svg" /></a>
</p>

[英文](README.md) | 🇨🇳中文

## 📖 简介

`ants`是一个高性能的 goroutine 池，实现了对大规模 goroutine 的调度管理、goroutine 复用，允许使用者在开发并发程序的时候限制 goroutine 数量，复用资源，达到更高效执行任务的效果。

## 🚀 功能：

- 自动调度海量的 goroutines，复用 goroutines
- 定期清理过期的 goroutines，进一步节省资源
- 提供了大量有用的接口：任务提交、获取运行中的 goroutine 数量、动态调整 Pool 大小、释放 Pool、重启 Pool
- 优雅处理 panic，防止程序崩溃
- 资源复用，极大节省内存使用量；在大规模批量并发任务场景下比原生 goroutine 并发具有[更高的性能](#-性能小结)
- 非阻塞机制

## ⚔️ 目前测试通过的Golang版本：

- 1.8.x
- 1.9.x
- 1.10.x
- 1.11.x
- 1.12.x
- 1.13.x
- 1.14.x

## 💡 `ants` 是如何运行的

### 流程图

<p align="center">
<img width="845" alt="ants-flowchart-cn" src="https://user-images.githubusercontent.com/7496278/66396519-7ed66e00-ea0c-11e9-9c1a-5ca54bbd61eb.png">
</p>

### 动态图

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-1.png)

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-2.png)

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-3.png)

![](https://raw.githubusercontent.com/panjf2000/illustrations/master/go/ants-pool-4.png)

## 🧰 安装

### 使用 `ants` v1 版本:

``` powershell
go get -u github.com/panjf2000/ants
```

### 使用 `ants` v2 版本 (开启 GO111MODULE=on):

```powershell
go get -u github.com/panjf2000/ants/v2
```

## 🛠 使用
写 go 并发程序的时候如果程序会启动大量的 goroutine ，势必会消耗大量的系统资源（内存，CPU），通过使用 `ants`，可以实例化一个 goroutine 池，复用 goroutine ，节省资源，提升性能：

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

### Pool 配置

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

通过在调用`NewPool`/`NewPoolWithFunc`之时使用各种 optional function，可以设置`ants.Options`中各个配置项的值，然后用它来定制化 goroutine pool.


### 自定义池
`ants`支持实例化使用者自己的一个 Pool ，指定具体的池容量；通过调用 `NewPool` 方法可以实例化一个新的带有指定容量的 Pool ，如下：

``` go
// Set 10000 the size of goroutine pool
p, _ := ants.NewPool(10000)
```

### 任务提交

提交任务通过调用 `ants.Submit(func())`方法：
```go
ants.Submit(func(){})
```

### 动态调整 goroutine 池容量
需要动态调整 goroutine 池容量可以通过调用`Tune(int)`：

``` go
pool.Tune(1000) // Tune its capacity to 1000
pool.Tune(100000) // Tune its capacity to 100000
```

该方法是线程安全的。

### 预先分配 goroutine 队列内存

`ants`允许你预先把整个池的容量分配内存， 这个功能可以在某些特定的场景下提高 goroutine 池的性能。比如， 有一个场景需要一个超大容量的池，而且每个 goroutine 里面的任务都是耗时任务，这种情况下，预先分配 goroutine 队列内存将会减少不必要的内存重新分配。

```go
// ants will pre-malloc the whole capacity of pool when you invoke this function
p, _ := ants.NewPool(100000, ants.WithPreAlloc(true))
```

### 释放 Pool

```go
pool.Release()
```

### 重启 Pool

```go
// 只要调用 Reboot() 方法，就可以重新激活一个之前已经被销毁掉的池，并且投入使用。
pool.Reboot()
```

## ⚙️ 关于任务执行顺序

`ants` 并不保证提交的任务被执行的顺序，执行的顺序也不是和提交的顺序保持一致，因为在 `ants` 是并发地处理所有提交的任务，提交的任务会被分派到正在并发运行的 workers 上去，因此那些任务将会被并发且无序地被执行。

## 🧲 Benchmarks

<div align="center"><img src="https://user-images.githubusercontent.com/7496278/51515466-c7ce9e00-1e4e-11e9-89c4-bd3785b3c667.png"/></div>
上图中的前两个 benchmark 测试结果是基于100w 任务量的条件，剩下的几个是基于 1000w 任务量的测试结果，`ants` 的默认池容量是 5w。

- BenchmarkGoroutine-4 代表原生 goroutine

- BenchmarkPoolGroutine-4 代表使用 goroutine 池 `ants`

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

## 📊 性能小结

![](https://user-images.githubusercontent.com/7496278/63449727-3ae6d400-c473-11e9-81e3-8b3280d8288a.gif)

**从该 demo 测试吞吐性能对比可以看出，使用`ants`的吞吐性能相较于原生 goroutine 可以保持在 2-6 倍的性能压制，而内存消耗则可以达到 10-20 倍的节省优势。** 

## 👏 贡献者

请在提 PR 之前仔细阅读 [Contributing Guidelines](CONTRIBUTING.md)，感谢那些为 `ants` 贡献过代码的开发者！

[![](https://opencollective.com/ants/contributors.svg?width=890&button=false)](https://github.com/panjf2000/ants/graphs/contributors)

## 📄 证书

`ants` 的源码允许用户在遵循 [MIT 开源证书](/LICENSE) 规则的前提下使用。

## 📚 相关文章

-  [Goroutine 并发调度模型深度解析之手撸一个高性能 goroutine 池](https://taohuawu.club/high-performance-implementation-of-goroutine-pool)
-  [Visually Understanding Worker Pool](https://medium.com/coinmonks/visually-understanding-worker-pool-48a83b7fc1f5)
-  [The Case For A Go Worker Pool](https://brandur.org/go-worker-pool)
-  [Go Concurrency - GoRoutines, Worker Pools and Throttling Made Simple](https://twinnation.org/articles/39/go-concurrency-goroutines-worker-pools-and-throttling-made-simple)

## 🖥 用户案例

欢迎在这里添加你的案例~~

<a href="https://github.com/panjf2000/gnet" target="_blank"><img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/go/gnet-logo.svg" width="250" align="middle"/></a>&nbsp;&nbsp;
<a href="https://www.tencent.com"><img src="https://www.tencent.com/img/index/tencent_logo.png" width="250" align="middle"/></a>&nbsp;&nbsp;

## 🔋 JetBrains 开源证书支持

`ants` 项目一直以来都是在 JetBrains 公司旗下的 GoLand 集成开发环境中进行开发，基于 **free JetBrains Open Source license(s)** 正版免费授权，在此表达我的谢意。

<a href="https://www.jetbrains.com/?from=ants" target="_blank"><img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/jetbrains/jetbrains-variant-4.png" width="250" align="middle"/></a>

## 💰 支持

如果有意向，可以通过每个月定量的少许捐赠来支持这个项目。

<a href="https://opencollective.com/ants#backers" target="_blank"><img src="https://opencollective.com/ants/backers.svg"></a>

## 💎 赞助

每月定量捐赠 10 刀即可成为本项目的赞助者，届时您的 logo 或者 link 可以展示在本项目的 README 上。

<a href="https://opencollective.com/ants#sponsors" target="_blank"><img src="https://opencollective.com/ants/sponsors.svg"></a>

## ☕️ 打赏

> 当您通过以下方式进行捐赠时，请务必留下姓名、Github账号或其他社交媒体账号，以便我将其添加到捐赠者名单中，以表谢意。

<img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/payments/WeChatPay.JPG" width="250" align="middle"/>&nbsp;&nbsp;
<img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/payments/AliPay.JPG" width="250" align="middle"/>&nbsp;&nbsp;
<a href="https://www.paypal.me/R136a1X" target="_blank"><img src="https://raw.githubusercontent.com/panjf2000/illustrations/master/payments/PayPal.JPG" width="250" align="middle"/></a>&nbsp;&nbsp;

### 捐赠者名单

<a target="_blank" href="https://github.com/patrick-othmer"><img src="https://avatars1.githubusercontent.com/u/8964313" width="100" alt="Patrick Othmer" /></a>&nbsp;<a target="_blank" href="https://github.com/panjf2000/gnet"><img src="https://avatars2.githubusercontent.com/u/50285334" width="100" alt="Jimmy" /></a>&nbsp;<a target="_blank" href="https://github.com/cafra"><img src="https://avatars0.githubusercontent.com/u/13758306" width="100" alt="ChenZhen" /></a>&nbsp;<a target="_blank" href="https://github.com/yangwenmai"><img src="https://avatars0.githubusercontent.com/u/1710912" width="100" alt="Mai Yang" /></a>