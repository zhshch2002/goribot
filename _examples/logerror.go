package main

import (
	"github.com/zhshch2002/goribot"
	"os"
	"time"
)

func main() {
	f, err := os.Create("./test.log")
	if err != nil {
		goribot.Log.Fatal(err)
	}
	s := goribot.NewSpider(
		goribot.SpiderLogError(f),
	)
	s.Downloader.(*goribot.BaseDownloader).Client.Timeout = 5 * time.Second
	s.AddTask(goribot.GetReq("https://httpbin.org/get"), func(ctx *goribot.Context) {
		panic("some error!")
	})
	s.AddTask(goribot.GetReq("https://httpbin.org/get"), func(ctx *goribot.Context) {
		ctx.AddItem(goribot.ErrorItem{
			Ctx: ctx,
			Msg: "I left a message.",
		})
	})
	s.AddTask(goribot.GetReq("https://githab.com/"))
	s.Run()
}
