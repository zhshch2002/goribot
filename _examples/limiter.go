package main

import (
	"github.com/zhshch2002/goribot"
	"time"
)

func main() {
	s := goribot.NewSpider(
		goribot.Limiter(true, &goribot.LimitRule{
			Glob: "httpbin.org",
			//Allow: goribot.Allow,
			Delay: 5 * time.Second,
		}),
	)

	s.AddTask(
		goribot.GetReq("https://httpbin.org/get"),
		func(ctx *goribot.Context) {
			goribot.Log.Info("got 1")
		},
	)
	s.AddTask(
		goribot.GetReq("https://httpbin.org/get"),
		func(ctx *goribot.Context) {
			goribot.Log.Info("got 2")
		},
	)
	s.AddTask(
		goribot.GetReq("https://httpbin.org/get"),
		func(ctx *goribot.Context) {
			goribot.Log.Info("got 3")
		},
	)
	s.AddTask(
		goribot.GetReq("https://github.com"),
		func(ctx *goribot.Context) {
			goribot.Log.Info("shouldn't get")
		},
	)

	s.Run()
}
