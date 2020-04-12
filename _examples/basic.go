package main

import (
	"fmt"
	"github.com/zhshch2002/goribot"
)

func main() {
	goribot.Log.Info("Start spider")
	s := goribot.NewSpider()

	s.AddTask(
		goribot.GetReq("https://httpbin.org/get"),
		func(ctx *goribot.Context) {
			fmt.Println(ctx.Resp.Text)
			fmt.Println(ctx.Resp.Json("headers.User-Agent"))
		},
	)

	s.Run()
}
