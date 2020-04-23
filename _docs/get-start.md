# 开始使用
Goribot 是一个使用 Pipeline（流水线）模型的爬虫框架。任意一个 HTTP 请求（即 `goribot.Request` ) 都会被配以一序列回调函数（即 `handlers` ）。在 Goribot 中这些回调函数都是 `func(ctx *goribot.Context)` 的格式，其中的参数 `ctx` 包含了 HTTP 请求、响应以及爬虫的一些信息。

## 运行流程

以起始页面所给出的例子，一个爬虫的生命应该遵循以下流程。

``` Go
s := goribot.NewSpider() // 创建了一个爬虫
```

``` Go
s.AddTask(// 添加一个种子任务。

    // 👇 创建一个 Get 请求，关于请求的创建和配置将在【网络操作】中详细说明
    goribot.GetReq("https://httpbin.org/get"),

    // 👇 这就是上文说的【回调函数】，在请求完成后执行
    func(ctx *goribot.Context) {
        fmt.Println(ctx.Resp.Text)                       // 获取 HTTP 响应结果
        fmt.Println(ctx.Resp.Json("headers.User-Agent")) // 将结果作为 JSON 解析并获取指定内容
    },
)
```

``` Go
s.Run()  // 此时蜘蛛才开始真正运行。此调用会阻塞线程直到没有更多任务给蜘蛛工作。
```

::: warning
`s.AddTask` 只应作为种子任务创建的方式。如果您从一个页面获取了更多链接，此时需要使用 `ctx.Addtask` 。
:::

### 复杂的沿网页扩张爬行例子

``` Go
package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zhshch2002/goribot"
)

func main() {
	s := goribot.NewSpider() // 创建了一个爬虫

    var h goribot.CtxHandlerFun // 这是一个回调函数，用于发现新链接（不把他用 var 单出来声明就没法在回调函数内调用自己）
	h = func(ctx *goribot.Context) {
		fmt.Println(ctx.Resp.Request.URL)
		if ctx.Resp.Dom != nil {
			ctx.Resp.Dom.Find("a").Each(func(i int, sel *goquery.Selection) {
				if u := sel.AttrOr("href", ""); u !="" {
                    // 👇 注意在这里不是 s.AddTask 而是 ctx.AddTask
					ctx.AddTask(goribot.GetReq(u), h)
					// ☝ 蜘蛛会根据ctx里的信息自动处理相对地址，无需手动处理
				}
			})
		}
    }
    // 使用回调函数 h 来创建种子任务
	s.AddTask(goribot.GetReq("https://httpbin.org"), h)

	s.Run()
}

```

## 网络操作

Goribot 对 HTTP 的基本操作——请求（Request）、响应（Response）做了基本封装。

### 请求 Request

在向蜘蛛添加任务时创建的第一个参数就是请求，也就是你要访问哪个地址。

在蜘蛛的回调函数中，你可以使用 `ctx.Req` 来获取本次请求的信息。

创建一个基本的 Get 请求就像这样。

``` Go
req:=goribot.GetReq("https://httpbin.org")
```

但是可能想设置一下请求头的一些信息，于是：

``` Go
req:=goribot.GetReq("https://httpbin.org").SetHeader("hello","world")
```

Goribot 的请求配置遵循链式操作，如果在链式操作中有某个环节发生错误，会返回一个带有 `Err` 属性的请求，同时其之后的链式操作不会被执行，显然的，蜘蛛也不会执行一个带有 `Err` 的请求，但会调用 `OnError` 回调函数（会在 [【回调函数】](#%E5%9B%9E%E8%B0%83%E5%87%BD%E6%95%B0) 中详细说明）。

一个 Request 结构如下所示。

``` Go
// goribot/net.go
type Request struct {
    // 继承 http.Request 的功能
    *http.Request
    // 记录爬取深度，这个初始化为 - 1 后由爬虫维护，种子任务记为 Depth=-1
	Depth int
	// 这个请求的代理配置，不适用即为空
	ProxyURL string
	// 一个可以自定义配置的地方，会沿着 Request->Response->Context 的方向传递。
    Meta map[string]interface{}
    // 链式配置时标记出错处
	Err  error
}
```

#### 创建请求

``` Go
// 创建 Get 请求
func GetReq(urladdr string) *Request
// 创建基本 Post 请求，允许传入一个 io reader。此后 Post 创建函数基于此。
func PostReq(urladdr string, body io.Reader) *Request
// 创建 Post 请求并传入 bytes 数据
func PostRawReq(urladdr string, body []byte) *Request
// 创建 Post 请求并设置 Form 参数，此函数将自动设置 Content-Type 请求头
func PostFormReq(urladdr string, requestData map[string]string) *Request
// 创建 Post 请求并设置 Json 参数，此函数将自动设置 Content-Type 请求头
func PostJsonReq(urladdr string, requestData interface{}) *Request
```

#### 链式操作

``` Go
// 设置 Get 参数
func (s *Request) AddParam(k, v string) *Request
// 添加 Cookie
func (s *Request) AddCookie(c *http.Cookie) *Request
// 设置 Header
func (s *Request) SetHeader(key, value string) *Request
// 设置代理
func (s *Request) SetProxy(p string) *Request
// 设置 UA
func (s *Request) SetUA(ua string) *Request
// 设置 Meta 参数，将在【回调函数 > Context】章节讲到
func (s *Request) WithMeta(k, v string) *Request
```

### 响应 Response

在蜘蛛的回调函数中，你可以使用 `ctx.Resp` 来获取响应结果。

``` Go
type Response struct {
    // 继承自 * http.Response。
	*http.Response
	// 覆盖了 * http.Response 的 Body 属性，这个 Body 会针对 Content-Type 为文本的结果做编码解码，也会对 gzip 响应做解压。
	Body []byte
	// 对 Content-Type 为文本的结果做解码而得来
	Text string
	// 响应所对应的请求
	Req *goribot.Request
	// 对 Content-Type 为 HTML 的结果解析为 goquery 的 Document 对象
	Dom *goquery.Document
	// 呈递自 Request 时配置的 Meta 信息
	Meta map[string]interface{}
}
```

响应自动解析函数：

``` Go
func (s *Response) DecodeAndParse() error
```

调用该函数会自动解码响应结果，包括编码识别和解压，理论上这一函数已经在 Http 请求后被蜘蛛调用。

#### Json、HTML 数据解析

针对 Content-Type 中标明 HTML 和 Json 的响应，蜘蛛已经实现了自动处理。其中：

HTML 对象可以使用 [goquery](https://github.com/PuerkitoBio/goquery) 来访问。

``` Go
a:=ctx.Resp.Dom.Find("a")
```

Json 对象使用了 [gjson](https://github.com/tidwall/gjson) 支持。可以使用 Response 的 Json 方法访问。

``` Go
d:=ctx.Resp.Json("data").String()
```

## 回调函数

回调函数是 Goribot 中处理数据的主要方式，其分为两种，一类是蜘蛛本身身生命周期的回调函数，另外是每个请求都可以带有一系列回调函数。

### 请求携带的回调函数

``` Go
s.AddTask(
    goribot.GetReq("https://httpbin.org/get"),
    func(ctx *goribot.Context) { // 👈 这个函数只会在这个请求收到响应后触发
        fmt.Println(ctx.Resp.Text)
        fmt.Println(ctx.Resp.Json("headers.User-Agent"))
    },
    func(ctx *goribot.Context) { // 👈 还可以多来几个回调，这就是 Pipeline 流水线模型
        fmt.Println("second handler")
    },
)
```

### 蜘蛛生命周期回调 - Hook （钩子）

以下函数触发在蜘蛛运行的不同时期，每个函数都遵守 Pipeline 流水线模式，也就是可以添加好几次，蜘蛛会安添加次序在相应的时期顺序执行。Goribot 中的很多扩展功能（后文将讲到）都是通过这些回调实现的。

体验 Goribot 生命周期钩子函数：
::: details 展开

``` Go
package main

import (
	"fmt"
	"github.com/zhshch2002/goribot"
)

func main() {
	s := goribot.NewSpider()
	s.OnStart(func(s *goribot.Spider) {
		fmt.Println("OnStart")
	})
	s.OnAdd(func(ctx *goribot.Context, ta *goribot.Task) *goribot.Task {
		fmt.Println("OnAdd")
		return ta
	})
	s.OnReq(func(ctx *goribot.Context, req *goribot.Request) *goribot.Request {
		fmt.Println("OnReq")
		return req
	})
	s.AddTask(
		goribot.GetReq("https://httpbin.org/get?Goribot%20test=hello%20world").SetParam(map[string]string{
			"Goribot test": "hello world",
		}),
		func(ctx *goribot.Context) {
			fmt.Println("got resp data", ctx.Resp.Text)
			ctx.AddItem(ctx.Resp.Text)
		},
		func(ctx *goribot.Context) {
			fmt.Println("Handler 2")
			panic("some error")
		},
	)
	s.OnItem(func(i interface{}) interface{} {
		fmt.Println("OnItem")
		return i
	})
	s.OnError(func(ctx *goribot.Context, err error) {
		fmt.Println(err)
	})
	s.OnFinish(func(s *goribot.Spider) {
		fmt.Println("OnFinish")
	})
	s.Run()
}
```

:::

``` Go
// 在蜘蛛执行 s.Run() 时一开始执行一次
func (s *Spider) OnStart(fn func(s *Spider))
// 在所有线程结束后，蜘蛛即将退出时调用一次
func (s *Spider) OnFinish(fn func(s *Spider))
// 有新的任务添加到队列里之前执行
// ❗ 这个函数不是线程安全的，他可能被在多线程环境下调用
// ❗❗ 其中参数 ctx 的值可能为空，是因为创建种子任务时无上下文环境
func (s *Spider) OnAdd(fn func(ctx *Context, t *Task) *Task)
// 在发出新的 Http 请求前执行
// ❗ 这个函数不是线程安全的，他可能被在多线程环境下调用
func (s *Spider) OnReq(fn func(ctx *Context, req *Request) *Request)
// 有新的 Http 响应时执行，请求携带的回调函数在此之后运行
// ❗ 这个函数不是线程安全的，他可能被在多线程环境下调用
func (s *Spider) OnResp(fn func(ctx *Context))
// 有新的 Item 提交到队列后执行
// ❗ 这个函数不是线程安全的，他可能被在多线程环境下调用
func (s *Spider) OnItem(fn func(i interface{}) interface{})
// 蜘蛛内有 error 或 panic 发生 recover 后执行
// ❗ 这个函数不是线程安全的，他可能被在多线程环境下调用
func (s *Spider) OnError(fn func(ctx *Context, err error))
```

#### 串联的 Hook 函数

你会注意到类似 `func (s *Spider) OnReq(fn func(ctx *Context, req *Request) *Request)` 中的 Hook 函数 `func(ctx *Context, req *Request) *Request` 传入了 `*Request` 后又返回了 `*Request` 。由此我们可以在函数内修改 `req` 的内容，然后返回，之后的下一个 `OnReq` Hook 函数将收到新的 `req` 内容。

::: tip 提示
返回一个 `nil` 将会使 Hook 函数队列 **抛弃** 当前的 `req` 内容，继续执行其他的任务。
:::

有同样设计的 Hook 函数有：

* `func (s *Spider) OnReq(fn func(ctx *Context, req *Request) *Request)` 
* `func (s *Spider) OnAdd(fn func(ctx *Context, t *Task) *Task)` 
* `func (s *Spider) OnItem(fn func(i interface{}) interface{})` 

### Context

我们已经反复提及 Goribot 中代表运行上下文的 `Context` ，他实际上是这样的：

``` Go
type Context struct {
	// 发起的请求
	Req *Request
	// 得到的响应（注意在未得到响应前，例如 OnReq 中，此属性为空）
	Resp *Response
	// Meta 一开始由 Request 创建时设置，用于在不同 Handler 之间携带数据
	Meta map[string]interface{}
}
```

### 中断 Handler 处理链

当你调用 `ctx.Abort()` 后，之后的 Handler 将不再被执行，但蜘蛛仍会从中收集新的 Task 和 Item。

## 爬虫结果收集

你在之前的内容中已经发现了 `ctx.Additem()` 和 `s.OnItem()` 。这两兄弟就是 Goribot 用于收集爬虫所获取的数据的工具。

诚然，你可以在蜘蛛的回调函数里收集结果、提交到数据库或者写出到文件。但蜘蛛的回调函数应该是处理 HTTP 请求和响应的，读写数据库和文件将会占用时间、影响爬取效率、还有可能造成意外的 panic。

Goribot 内置了基于这两个接口的一些数据收集插件，如下：

* [SaveItemsAsJson](./extensions.html#saveitemsasjson-%e4%bf%9d%e5%ad%98-item-%e5%88%b0-json-%e6%96%87%e4%bb%b6)
* [SaveItemsAsCSV](./extensions.html#saveitemsascsv-%e4%bf%9d%e5%ad%98-item-%e5%88%b0-csv-%e6%96%87%e4%bb%b6)

## 写一个 Goribot 扩展吧！

Goribot 扩展就是一个形如 `func(s *Spider)` 的函数，传入一个 `Spider` 指针来修改蜘蛛的配置或者添加 Hook 函数。

比如一个自动添加 Referer 头部的插件：
```Go
func RefererFiller() func(s *Spider) {
	return func(s *Spider) {
		s.OnAdd(func(ctx *Context, t *Task) *Task {
			if ctx != nil {
				t.Request.Header.Set("Referer", ctx.Resp.Request.URL.String())
			}
			return t
		})
	}
}
```

之后在创建 Spider 时：
```Go
s := goribot.NewSpider(
	goribot.RefererFiller(),
)
```
或者（效果是一样的）：
```Go
s := goribot.NewSpider()
s.Use(goribot.RefererFiller())
```
