package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zhshch2002/goribot"
	"strings"
)

func main() {
	s := goribot.NewSpider(
		goribot.Limiter(true, &goribot.LimitRule{
			Glob: "*.bilibili.com",
			Rate: 2,
		}),
		goribot.RefererFiller(),
		goribot.RandomUserAgent(),
		goribot.SetDepthFirst(true),
	)
	var getVideoInfo = func(ctx *goribot.Context) {
		res := map[string]interface{}{
			"bvid":  ctx.Resp.Json("data.bvid").String(),
			"title": ctx.Resp.Json("data.title").String(),
			"des":   ctx.Resp.Json("data.des").String(),
			"pic":   ctx.Resp.Json("data.pic").String(),   // 封面图
			"tname": ctx.Resp.Json("data.tname").String(), // 分类名
			"owner": map[string]interface{}{ //视频作者
				"name": ctx.Resp.Json("data.owner.name").String(),
				"mid":  ctx.Resp.Json("data.owner.mid").String(),
				"face": ctx.Resp.Json("data.owner.face").String(), // 头像
			},
			"ctime":   ctx.Resp.Json("data.ctime").String(),   // 创建时间
			"pubdate": ctx.Resp.Json("data.pubdate").String(), // 发布时间
			"stat": map[string]interface{}{ // 视频数据
				"view":     ctx.Resp.Json("data.stat.view").Int(),
				"danmaku":  ctx.Resp.Json("data.stat.danmaku").Int(),
				"reply":    ctx.Resp.Json("data.stat.reply").Int(),
				"favorite": ctx.Resp.Json("data.stat.favorite").Int(),
				"coin":     ctx.Resp.Json("data.stat.coin").Int(),
				"share":    ctx.Resp.Json("data.stat.share").Int(),
				"like":     ctx.Resp.Json("data.stat.like").Int(),
				"dislike":  ctx.Resp.Json("data.stat.dislike").Int(),
			},
		}
		ctx.AddItem(res)
	}
	var findVideo goribot.CtxHandlerFun
	findVideo = func(ctx *goribot.Context) {
		u := ctx.Req.URL.String()
		fmt.Println(u)
		if strings.HasPrefix(u, "https://www.bilibili.com/video/") {
			if strings.Contains(u, "?") {
				u = u[:strings.Index(u, "?")]
			}
			u = u[31:]
			fmt.Println(u)
			ctx.AddTask(goribot.GetReq("https://api.bilibili.com/x/web-interface/view?bvid="+u), getVideoInfo)
		}
		ctx.Resp.Dom.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
			if h, ok := sel.Attr("href"); ok {
				ctx.AddTask(goribot.GetReq(h), findVideo)
			}
		})
	}
	s.OnItem(func(i interface{}) interface{} {
		fmt.Println(i)
		return i
	})
	s.AddTask(goribot.GetReq("https://www.bilibili.com/video/BV1at411a7RS").SetHeader("cookie", "_uuid=1B9F036F-8652-DCDD-D67E-54603D58A9B904750infoc; buvid3=5D62519D-8AB5-449B-A4CF-72D17C3DFB87155806infoc; sid=9h5nzg2a; LIVE_BUVID=AUTO7815811574205505; CURRENT_FNVAL=16; im_notify_type_403928979=0; rpdid=|(k|~uu|lu||0J'ul)ukk)~kY; _ga=GA1.2.533428114.1584175871; PVID=1; DedeUserID=403928979; DedeUserID__ckMd5=08363945687b3545; SESSDATA=b4f022fe%2C1601298276%2C1cf0c*41; bili_jct=2f00b7d205a97aa2ec1475f93bfcb1a3; bp_t_offset_403928979=375484225910036050"), findVideo)
	s.Run()
}
