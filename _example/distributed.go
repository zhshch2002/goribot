package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhshch2002/goribot"
)

// docker run --name some-redis -d -p 6379:6379 redis
func main() {
	ro := &redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	sName := "DistributedTest"
	m := goribot.NewManager(redis.NewClient(ro), sName)
	m.OnItem(func(i interface{}) interface{} {
		fmt.Println(i)
		return i
	})
	m.SendReq(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hello world"))

	h := func(ctx *goribot.Context) {
		goribot.Log.Info("got resp")
		ctx.AddItem(ctx.Resp.Text)
	}
	s := goribot.NewSpider(
		goribot.RedisDistributed(
			ro,
			sName,
			true,
			func(ctx *goribot.Context) {
				goribot.Log.Info("got seed resp")
				ctx.AddItem(ctx.Resp.Text)
				ctx.AddTask(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hi!"), h)
				ctx.AddTask(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hi!"), h)
			},
		),
	)

	go s.Run()
	m.Run()
}
