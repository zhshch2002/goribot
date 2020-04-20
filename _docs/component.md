# Goribot 的组成
为了更好地使用 Goribot 的功能，你可以通过更多了解 Goribot 的组成来 Hack 更多的功能。

## 基本结构

Goribot 由一下几个主要部分构成：

1. Spider 蜘蛛：提供一个可以编程的界面和接口，使用这个对象并对其扩展以实现爬虫的主要功能。
2. Downloader 下载器：参考 Scrapy 的设计，作为 Spider 对象的一个组成部分，用于执行 HTTP 请求。Goribot 默认的下载器就是对`http/net`的包装。开发者可以自定义下载器以实现例如使用 Chrome 无头浏览器执行 HTTP 请求。
3. Scheduler 调度器：一个任务队列，作为 Spider 对象的一个组成部分，管理蜘蛛即将执行的 HTTP 请求及其回调函数。对这部分自定义可以实现分布式任务分发的功能。
4. Manager 管理器：[Goribot](https://github.com/zhshch2002/goribot) 自带的分布式工具，用于集中地分发任务并取回结果。

## Spider 蜘蛛
```go
type Spider struct {
        Scheduler                         Scheduler
        Downloader                        Downloader
        AutoStop                          bool
        taskPool, itemPool                *ants.Pool
        onStartHandlers, onFinishHandlers []func(s *Spider)
        onReqHandlers                     []func(ctx *Context, req *Request) *Request
        onAddHandlers                     []func(ctx *Context, req *Task) *Task
        onRespHandlers                    []CtxHandlerFun
        onItemHandlers                    []func(i interface{}) interface{}
        onErrorHandlers                   []func(ctx *Context, err error)
}

func NewSpider(exts ...func(s *Spider)) *Spider
func (s *Spider) AddTask(request *Request, handlers ...CtxHandlerFun)
func (s *Spider) OnAdd(fn func(ctx *Context, t *Task) *Task)
func (s *Spider) OnError(fn func(ctx *Context, err error))
func (s *Spider) OnFinish(fn func(s *Spider))
func (s *Spider) OnItem(fn func(i interface{}) interface{})
func (s *Spider) OnReq(fn func(ctx *Context, req *Request) *Request)
func (s *Spider) OnResp(fn CtxHandlerFun)
func (s *Spider) OnStart(fn func(s *Spider))
func (s *Spider) Run()
func (s *Spider) SetItemPoolSize(i int)
func (s *Spider) SetTaskPoolSize(i int)
func (s *Spider) Use(fn ...func(s *Spider))
func (s *Spider) handleOnAdd(ctx *Context, t *Task) *Task
func (s *Spider) handleOnError(ctx *Context, err error)
func (s *Spider) handleOnFinish()
func (s *Spider) handleOnItem(i interface{})
func (s *Spider) handleOnReq(ctx *Context, req *Request) *Request
func (s *Spider) handleOnResp(ctx *Context)
func (s *Spider) handleOnStart()
```

### 下载器和调度器

`Spider.Downloader`和`Spider.Scheduler`分别就是下载器和调度器了。他们分别都是连个`interface`：

```go
type Scheduler interface {
	GetTask() *Task
	GetItem() interface{}
	AddTask(t *Task)
	AddItem(i interface{})
	IsTaskEmpty() bool
	IsItemEmpty() bool
}

type Downloader interface {
	Do(req *Request) (resp *Response, err error)
	AddMiddleware(func(req *Request, next func(req *Request) (resp *Response, err error)) (resp *Response, err error))
}
```

在调用`goribot.NewSpider()`时会自动装配 Goribot 默认的下载器和调度器，即`BaseDownloader`和`BaseScheduler`。

### Use

`Spider.Use()`函数用于装配一个 Spider 插件，本质上就是个`func(s *Spider)`函数。用于调整一些配置。插件的大部分功能都是靠 Spider 提供的 [Hook 函数](./get-start.html#%E8%9C%98%E8%9B%9B%E7%94%9F%E5%91%BD%E5%91%A8%E6%9C%9F%E5%9B%9E%E8%B0%83-hook-%EF%BC%88%E9%92%A9%E5%AD%90%EF%BC%89) 接口实现的。

### SetTaskPoolSize 与 SetItemPoolSize
在 Goribot 爬虫内，会创建两个线程池，即 Task 线程池和 Item 线程池。分别用于处理爬虫任务和存储爬取结果。这两个函数用于调整线程池大小。这个调整是实时的，也就是爬虫运行后也可以进行调整。

## 下载器 Downloader
```go
type Downloader interface {
	Do(req *Request) (resp *Response, err error)
	AddMiddleware(func(req *Request, next func(req *Request) (resp *Response, err error)) (resp *Response, err error))
}
```
下载器主要的功能就是执行 HTTP 请求，得到响应（或错误）。下载的`Do(req *Request) (resp *Response, err error)`函数用于执行这一功能。

### AddMiddleware
AddMiddleware 函数用于添加下载器扩展，即中间件。

添加的扩展本身是一个函数`func(req *Request, next func(req *Request) (resp *Response, err error)) (resp *Response, err error)`。在这个函数中如果能处理 Request 则返回 resp 或者 err，否则调用 next 函数，即下一个函数，如此套娃。

## Scheduler 调度器
```go
type Scheduler interface {
	GetTask() *Task
	GetItem() interface{}
	AddTask(t *Task)
	AddItem(i interface{})
	IsTaskEmpty() bool
	IsItemEmpty() bool
}
```
管理器用于维护两个队列，以供蜘蛛能获取任务和 Item。

## Manager 管理器
```go
type Manager struct {
        itemPool       *ants.Pool
        redis          *redis.Client
        sName          string
        onItemHandlers []func(i interface{}) interface{}
}

func NewManager(redis *redis.Client, sName string) *Manager
func (s *Manager) GetItem() interface{}
func (s *Manager) OnItem(fn func(i interface{}) interface{})
func (s *Manager) Run()
func (s *Manager) SendReq(req *Request)
func (s *Manager) SetItemPoolSize(i int)
func (s *Manager) handleOnItem(i interface{})
```

管理器实际上是由管理节点运行，向 Redis 数据库添加任务，并从 Redis 上取回 Item。

### SendReq
SendReq 与 Spider 的 AddTask 对应，用于添加种子任务。种子任务会被爬虫节点拉取并按照设定好的回调函数执行，具体请见 [分布式支持](./distributed.html) 相关文档。

**要注意的是：** SendReq 并不是在 Run 函数调用后才执行。

### GetItem
GetItem 与调度器里的 GetItem 类似，是从 Redis 里 pop 出一个 Item。

### OnItem
OnItem 则与 Spider 的 OnItem 的 Hook 函数一样，不过是处理 Redis 回收来的结果。

### Run
Run 与 Spider 一样，不过只启动的是 Item 线程池。（毕竟没有 Task 给 Manager 来处理）

**要注意的是：** SendReq 并不是在 Run 函数调用后才执行。