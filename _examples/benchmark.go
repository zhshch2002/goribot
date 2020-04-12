package main

import (
	"github.com/zhshch2002/goribot"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1)
	s := goribot.NewSpider(
		goribot.SpiderLogPrint(),
	)
	s.SetTaskPoolSize(500)
	s.SetItemPoolSize(0)
	i := 0
	target := 100_000
	for i < target {
		s.AddTask(goribot.GetReq("http://127.0.0.1:1229"), func(ctx *goribot.Context) {
			if ctx.Resp.Text != "Hello goribot" {
				goribot.Log.Error("wrong response text", ctx.Resp.Text)
			}
		})
		i += 1
	}
	t := time.Now()
	s.Run()
	goribot.Log.Info("Total used", time.Since(t).Seconds(), "sec")
	goribot.Log.Info(float64(target)/time.Since(t).Seconds(), "task/sec")
}
