package main

import (
	"fmt"
	"github.com/zhshch2002/goribot"
	"time"
)

func main() {
	s := goribot.NewSpider(
		goribot.SpiderLogPrint(),
		goribot.Retry(3),
	)

	s.Downloader.(*goribot.BaseDownloader).Client.Timeout = 1 * time.Second

	s.AddTask(
		goribot.GetReq("https://githab.com/"),
		func(ctx *goribot.Context) {
			fmt.Println(ctx.Resp.Text)
		},
	)

	s.OnError(func(ctx *goribot.Context, err error) {
		if e, ok := err.(goribot.DownloaderErr); ok {
			fmt.Println(e)
		}
	})

	s.Run()
}
