# Goribot
A golang spider framework.

[中文文档](README_zh.md)

[![Codecov](https://img.shields.io/codecov/c/gh/zhshch2002/goribot)](https://codecov.io/gh/zhshch2002/goribot)
[![go-report](https://goreportcard.com/badge/github.com/zhshch2002/goribot)](https://goreportcard.com/report/github.com/zhshch2002/goribot)
![license](https://img.shields.io/github/license/zhshch2002/goribot)
![code-size](https://img.shields.io/github/languages/code-size/zhshch2002/goribot.svg)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fzhshch2002%2Fgoribot.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fzhshch2002%2Fgoribot?ref=badge_shield)

# Features
* Clean API
* Pipeline-style handle logic
* Robots.txt support (ues `RobotsTxt` extensions)
* Request Deduplicate (ues `ReqDeduplicate` extensions)
* Extensions

# Example
a basic example：
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
[a complete bilibili.com video spider example](#another-example)

# Start to use
## install
```shell
go get -u github.com/zhshch2002/goribot
```

## basic ues
### create spider
```go
s := goribot.NewSpider()
```
you can also init the spider by extensions,like the RandomUserAgent extension:
```go
s := NewSpider(RandomUserAgent())
```

### New task
create a request：
```go
req:=goribot.MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello")
// or req,err := goribot.NewGetReq("https://httpbin.org/get?Goribot%20test=hello")

// config the request
req.Header.Set("test", "test")
req.Cookie = append(req.Cookie, &http.Cookie{
    Name:  "test",
    Value: "test",
})
req.Proxy = "http://127.0.0.1:1080"
```

Add the request to spider task queue：
```go
var thirdHandler func(*goribot.Context)
thirdHandler= func(ctx *goribot.Context) {
    //bu la bu la,do sth
}

s.NewTask(
    req, // the request you have created
    func(ctx *goribot.Context) {
        // first handler
        fmt.Println("got resp data", ctx.Text)
    },
    func(ctx *goribot.Context) { // you can set a group of handler func as a chain,or set same func for different request task.
    // second handler
        fmt.Println("got resp data", ctx.Text)
    },
    thirdHandler,
)
```

### Context
`Context` is the only param the handler get.You can get the http response or the origin request from it,in addition you can use `ctx` send new request task to spider.

```go
type Context struct {
    Text string // the response text
    Html *goquery.Document // spider will try to parse the response as html
    Json map[string]interface{} // spider will try to parse the response as json

    Request  *Request // origin request
    Response *Response // a response object

    Tasks []*Task // the new request task which will send to the spider
    Items []interface{} // the new result data which will send to the spider，use to store
    Meta  map[string]interface{} // the request task created by NewTaskWithMeta func will have a k-y pair

    drop bool // in handlers chain,you can use ctx.Drop() to break the handler chain and stop handling
}
```

create new task inside of handle fun or with meta data：
```go
s.NewTaskWithMeta(MustNewGetReq("https://httpbin.org/get"), map[string]interface{}{
    "test": 1,
}, func(ctx *Context) {
    fmt.Println(ctx.Meta["test"]) // get the meta data
    
    // waring: here is the ctx.NewTaskWithMeta func rather than s.NewTaskWithMeta!
    ctx.NewTaskWithMeta(MustNewGetReq("https://httpbin.org/get"), map[string]interface{}{
        "test": 2,
    }, func(ctx *Context) {
        fmt.Println(ctx.Meta["test"]) // get the meta data
    })
})
```
Tip:It is different between `s.NewTaskWithMeta` and `ctx.NewTaskWithMeta`,when you use the extensions or spider hook func.

### Run it！
Call the `s.Run()` to run the spider.

## ues the hook func and make extensions
wait to write.

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
    s := goribot.NewSpider(goribot.HostFilter("www.bilibili.com"), goribot.ReqDeduplicate(), goribot.RandomUserAgent())
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
    * [x] 自动添加Referer
    * [x] Robots.txt解析
    * [x] 最大请求数限制
    * [x] Host过滤
    * [x] URL过滤
    * [ ] 随机代理
    * [x] 去重复
* `Spider`主体功能
    * [ ] 随机延时
    * [ ] 分布式调度

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fzhshch2002%2Fgoribot.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fzhshch2002%2Fgoribot?ref=badge_large)
