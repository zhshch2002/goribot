package main

import (
	"github.com/zhshch2002/goribot"
	"os"
)

func main() {
	cvsFile, err := os.Create("./test.cvs")
	if err != nil {
		panic(err)
	}
	defer cvsFile.Close()
	jsonFile, err := os.Create("./test.json")
	if err != nil {
		panic(err)
	}
	defer cvsFile.Close()
	s := goribot.NewSpider(
		goribot.SaveItemsAsCSV(cvsFile),
		goribot.SaveItemsAsJSON(jsonFile),
	)
	s.AddTask(
		goribot.GetReq("https://httpbin.org"),
		func(ctx *goribot.Context) {
			ctx.AddItem(goribot.CsvItem{
				ctx.Resp.Request.URL.String(),
				ctx.Resp.Dom.Find("title").Text(),
			})
			ctx.AddItem(goribot.JsonItem{Data: map[string]interface{}{
				"url":   ctx.Resp.Request.URL.String(),
				"title": ctx.Resp.Dom.Find("title").Text(),
			}})
			ctx.AddItem(goribot.CsvItem{
				ctx.Resp.Request.URL.String(),
				ctx.Resp.Dom.Find("title").Text(),
			})
			ctx.AddItem(goribot.JsonItem{Data: map[string]interface{}{
				"url":   ctx.Resp.Request.URL.String(),
				"title": ctx.Resp.Dom.Find("title").Text(),
			}})
		},
	)
	s.Run()
}
