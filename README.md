# Goribot
A golang spider framework.

[![Codecov](https://img.shields.io/codecov/c/gh/zhshch2002/goribot)](https://codecov.io/gh/zhshch2002/goribot)
[![go-report](https://goreportcard.com/badge/github.com/zhshch2002/goribot)](https://goreportcard.com/report/github.com/zhshch2002/goribot)
![license](https://img.shields.io/github/license/zhshch2002/goribot)
![code-size](https://img.shields.io/github/languages/code-size/zhshch2002/goribot.svg)

# Features
* Clean API
* Pipeline-style handle logic
* Robots.txt support
* Extensions

# Example
一个简单的实例：
```go
package main

import (
    "fmt"
    "github.com/zhshch2002/goribot"
)

func main() {
    s := goribot.NewSpider()
    s.NewTask(
        goribot.MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello%20world"),
        func(ctx *goribot.Context) {
            fmt.Println("got resp data", ctx.Text)
        })
    s.Run()
}
```
[完整的B站视频爬虫的例子](#another-example)

# Start to use
## install
```shell
go get -u github.com/zhshch2002/goribot
```

## 基本使用
### 创建蜘蛛
```go
s := goribot.NewSpider()
```
此期间允许使用插件来初始化蜘蛛，例如使用随机UA组件：
```go
s := NewSpider(RandomUserAgent)
```

### 创建任务
建立一个请求：
```go
req:=goribot.MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello")
// 或者 req,err := goribot.NewGetReq("https://httpbin.org/get?Goribot%20test=hello")

// 配置请求
req.Header.Set("test", "test") // 设置请求头部
req.Cookie = append(req.Cookie, &http.Cookie{ // 添加Cookie
    Name:  "test",
    Value: "test",
})
req.Proxy = "http://127.0.0.1:1080" // 配置代理
```

向蜘蛛添加任务：
```go
var thirdHandler func(*goribot.Context) // 可选的创建方式
thirdHandler= func(ctx *goribot.Context) {
    //bu la bu la,do sth
}

s.NewTask(
    req, // 上文创建的请求
    func(ctx *goribot.Context) {
        // first handler
        fmt.Println("got resp data", ctx.Text)
    },
    func(ctx *goribot.Context) { // 此处可以为一个请求设置多个处理回调函数，或者数个请求共用一个函数
    // second handler
        fmt.Println("got resp data", ctx.Text)
    },
    thirdHandler,
)
```

### Context
`Context`是处理函数收到的唯一参数。使用这个参数可以获得蜘蛛从网络获取的数据，也可以像蜘蛛提交新的任务`Task`。

```go
type Context struct {
    Text string // 收到的数据转换为字符串类型
    Html *goquery.Document // 自动尝试将收到的内容解析为Html
    Json map[string]interface{} // 自动尝试将收到的内容解析为Json

    Request  *Request // 上文中创建的请求
    Response *Response // 蜘蛛收到的网络响应，包含响应码、响应头、字节码原始数据等

    Tasks []*Task // 执行结束后将要提交给蜘蛛的新任务
    Items []interface{} // 执行结束后将要提交给蜘蛛的结果数据，用作数据持久化设计
    Meta  map[string]interface{} // 使用 NewTaskWithMeta 函数添加的任务可以携带一个Key-Val对应的数据

    drop bool // 在多个handler回调函数中可以调用ctx.Drop()函数来中断处理队列，使之后的回调函数不再执行
}
```

从Task内提交新的Task，以及使用Meta变量：
```go
s.NewTaskWithMeta(MustNewGetReq("https://httpbin.org/get"), map[string]interface{}{
    "test": 1,
}, func(ctx *Context) {
    fmt.Println(ctx.Meta["test"]) // 内层函数可以使用外层提供的数据
    
    // 在Handler内创建新的任务，注意这里是ctx.NewTaskWithMeta而非外层的s.NewTaskWithMeta
    ctx.NewTaskWithMeta(MustNewGetReq("https://httpbin.org/get"), map[string]interface{}{
        "test": 2,
    }, func(ctx *Context) {
        fmt.Println(ctx.Meta["test"]) // 内层函数可以使用外层提供的数据
    })
})
```
注：`s.NewTaskWithMeta`一般用来爬取种子地址，在用到钩子函数的时候（如：使用RandomUserAgent组件时），使用`ctx.NewTaskWithMeta`和`s.NewTaskWithMeta`对于钩子函数是有区别的。

### Run it！
经过上述初始化操作，需要执行`s.Run()`来启动蜘蛛。调用该函数将阻塞程序直到所有任务执行完毕。

# Another Example
A bilibili video spider:
```go
package main

import (
    "github.com/PuerkitoBio/goquery"
    "github.com/zhshch2002/goribot"
    "log"
    "strings"
)

type BiliVideoItem struct {
    Title, Url string
}

func main() {
    s := goribot.NewSpider(goribot.HostFilter("www.bilibili.com"), goribot.RandomUserAgent)
    s.DepthFirst = false
    s.ThreadPoolSize = 1

    var biliVideoHandler, getNewLinkHandler func(ctx *goribot.Context)
    getNewLinkHandler = func(ctx *goribot.Context) {
        ctx.Html.Find("a[href]").Each(func(i int, selection *goquery.Selection) {
            rawurl, _ := selection.Attr("href")
            if !strings.HasPrefix(rawurl, "/video/av") {
                return
            }
            u, err := ctx.Request.Url.Parse(rawurl)
            if err != nil {
                return
            }
            u.RawQuery = ""
            if strings.HasSuffix(u.Path, "/") {
                u.Path = u.Path[0 : len(u.Path)-1]
            }
            //log.Println(u.String())
            if r, err := goribot.NewGetReq(u.String()); err == nil {
                ctx.NewTask(r, getNewLinkHandler, biliVideoHandler)
            }
        })
    }
    biliVideoHandler = func(ctx *goribot.Context) {
        ctx.AddItem(BiliVideoItem{
            Title: ctx.Html.Find("title").Text(),
            Url:   ctx.Request.Url.String(),
        })
    }

    s.NewTask(goribot.MustNewGetReq("https://www.bilibili.com/video/av66703342"), getNewLinkHandler, biliVideoHandler)
    s.OnItem(func(ctx *goribot.Context, i interface{}) interface{} {
        log.Println(i) // 可以做一些数据存储工作
        return i
    })

    s.Run()
}
```

# TODO
* 制作更多的内建插件`extensions`
    * [x] 随机UA
    * [ ] 自动添加Referer
    * [x] Robots.txt解析
    * [ ] 最大请求数限制
    * [x] Host过滤
    * [ ] URL过滤
    * [ ] 随机代理
* `Spider`主体功能
    * [ ] 随机延时
    * [ ] 去重复