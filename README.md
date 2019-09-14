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

# Example:Bilibili.com spider
```go
package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/zhshch2002/goribot"
	"github.com/zhshch2002/goribot/goribotExts"
	"log"
	"os"
	"strings"
)

type BiliVideoItem struct {
	Title, Url string
}

type BiliPipe struct {
	goribot.Pipeline
	itemCount int
	f         *os.File
}

func (s *BiliPipe) Init(spider *goribot.Spider) {
	f, err := os.OpenFile("bilibili.txt", os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	s.f = f
}
func (s *BiliPipe) OnItem(spider *goribot.Spider, item interface{}) interface{} {
	s.itemCount += 1
	log.Println("got item", s.itemCount)
	if i, ok := item.(BiliVideoItem); ok {
		_, _ = s.f.Write([]byte(i.Title + "\t" + i.Url + "\n"))
	}
	return item
}
func (s *BiliPipe) Finish(spider *goribot.Spider) {
	_ = s.f.Close()
}

func main() {
	s := goribot.NewSpider()

	s.Use(goribotExts.NewAllowHostPipeline("www.bilibili.com"))
	s.Use(goribotExts.NewDeduplicatePipeline())
	s.Use(goribotExts.NewRandomUaPipeline())
	s.Use(goribotExts.NewRetryPipelineWithErrorCode(1, 404, 403))
	s.Use(&BiliPipe{})

	//s.RandSleepRange = [2]time.Duration{5 * time.Millisecond, 1 * time.Second}

	var biliVideoHandler goribot.ResponseHandler
	biliVideoHandler = func(r *goribot.Response) {
		s.NewItem(BiliVideoItem{
			Title: r.Html.Find("title").Text(),
			Url:   r.Request.Url.String(),
		})

		r.Html.Find("a[href]").Each(func(i int, selection *goquery.Selection) {
			rawurl, _ := selection.Attr("href")
			if !strings.HasPrefix(rawurl, "/video/av") {
				return
			}
			u, err := r.Request.Url.Parse(rawurl)
			if err != nil {
				return
			}
			u.RawQuery = ""
			if strings.HasSuffix(u.Path, "/") {
				u.Path = u.Path[0 : len(u.Path)-1]
			}
			//log.Println(u.String())
			_ = s.Get(r, u.String(), biliVideoHandler)
		})
	}

	_ = s.Get(nil, "https://www.bilibili.com/video/av66703342", biliVideoHandler)

	s.Run()
}
```

# TODO
* Basic
  * [x] 实现请求随机延时
* Pipeline
  * [ ] AllowHostPipeline
  * [x] DeduplicatePipeline
  * [x] UrlFilterPipeline