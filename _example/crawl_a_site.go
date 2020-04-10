package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zhshch2002/goribot"
)

func main() {
	s := goribot.NewSpider()

	var h goribot.CtxHandlerFun
	h = func(ctx *goribot.Context) {
		fmt.Println(ctx.Resp.Request.URL)
		if ctx.Resp.Dom != nil {
			ctx.Resp.Dom.Find("a").Each(func(i int, sel *goquery.Selection) {
				if u := sel.AttrOr("href", ""); u != "" {
					ctx.AddTask(goribot.GetReq(u), h)
				}
			})
		}
	}
	s.AddTask(goribot.GetReq("https://httpbin.org"), h)

	s.Run()
}
