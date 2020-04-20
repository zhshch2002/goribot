package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
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
	s.OnResp(func(ctx *goribot.Context) {
		fmt.Println("OnResp")
	})
	s.OnHTML("title", func(ctx *goribot.Context, sel *goquery.Selection) {
		fmt.Println("on html,title:", sel.Text())
	})
	s.OnJSON("args", func(ctx *goribot.Context, j gjson.Result) {
		fmt.Println("on json")
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
	s.AddTask(goribot.GetReq("https://httpbin.org"))
	s.Run()
}
