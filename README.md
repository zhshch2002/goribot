# Goribot
A golang spider framework.

# Example
```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/zhshch2002/goribot"
)

func main() {
    s := goribot.NewSpider()
	_ = s.Get(nil, "https://httpbin.org/get?Goribot%20test=hello%20world", func(r *goribot.Response) {
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(r.Text), &m)
		if err != nil {
			fmt.Println(err)
		}
        fmt.Println(m)
	})
}
```

# Pipeline
Pipeline是一个实现了`PipelineInterface`的`struct`。包含了数个钩子函数。
```go
package goribot

type PipelineInterface interface {
	Init(spider *goribot.Spider)
	OnDoRequest(spider *Spider, request *goribot.Request) *goribot.Request
	OnNewRequest(spider *goribot.Spider, preResp *goribot.Response, request *goribot.Request) *goribot.Request
	OnResponse(spider *goribot.Spider, response *goribot.Response) *goribot.Response
	OnItem(spider *goribot.Spider, item interface{}) interface{}
	OnError(spider *goribot.Spider, err error)
	Finish(spider *goribot.Spider)
}
```

在实例化的Spider对象上，使用`func (s *Spider) Use(p PipelineInterface)`方法可以装入一个Pipeline。每次Spider会按照注册的顺序依次调用注册过的Pipeline里的函数。
```go
s := goribot.NewSpider()
s.Use(goribotExts.NewAllowHostPipeline("www.bilibili.com")) // 使用域名过滤Pipeline
```
## 接力棒逻辑
在Pipeline中具有返回值的函数，可以通过返回一个新的结果来改变相应的`Request`、`Response`和`Item`。`Spider`会顺序调用所有注册的`Pipeline`的钩子函数（例如有新`Request`时会调用每个`OnNewRequest`函数），意味着上一个`OnNewRequest`函数传出的结构会被送给下一个`OnNewRequest`函数。这就是接力棒逻辑。
当接力棒的某一个函数返回了`nil`结果，那么这个`Request`、`Response`或`Item`会被抛弃，不在参与之后的运行。

## 钩子函数
* 在使用`Use()`函数注册时调用 `Init(spider *goribot.Spider)`
* 在向蜘蛛队列加入新的`Request`时调用 `OnNewRequest(spider *goribot.Spider, preResp *goribot.Response, request *goribot.Request) *goribot.Request`
* 在`Downloader`下载新的`Request`时调用 `OnDoRequest(spider *Spider, request *goribot.Request) *goribot.Request`
* 在`Downloader`执行后调用 `OnResponse(spider *goribot.Spider, response *goribot.Response) *goribot.Response`
* 在`Spider`的`NewItem`函数被执行后调用 `OnItem(spider *goribot.Spider, item interface{}) interface{}`
* 在`Downloader`执行出错后调用 `OnError(spider *goribot.Spider, err error)`
* 在`Run()`函数执行完毕后调用 `Finish(spider *goribot.Spider)` (注意，一个Spider的Run函数可能被执行多次)

## 已经实现了的Pipeline
TODO：在goribotExts目录下。

# TODO
* Basic
  * [x] 实现请求随机延时
* Pipeline
  * [ ] AllowHostPipeline
  * [x] DeduplicatePipeline
  * [x] UrlFilterPipeline