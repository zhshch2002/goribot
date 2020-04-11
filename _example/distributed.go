package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhshch2002/goribot"
)

// docker run --name some-redis -d -p 6379:6379 redis
func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	sName := "DistributedTest"
	//fmt.Println(client.LPop(sName + goribot.ItemsSuffix).Bytes())
	m := goribot.NewManager(client, sName)
	m.OnItem(func(i interface{}) interface{} {
		fmt.Println(i)
		return i
	})
	m.SendReq(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hello world"))
	m.SendReq(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hello world"))
	m.SendReq(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hello world"))

	s := goribot.NewSpider()
	s.Scheduler = goribot.NewRedisScheduler(
		client,
		sName,
		10,
		func(ctx *goribot.Context) {
			goribot.Log.Info("got resp")
			ctx.AddItem(ctx.Resp.Text)
		},
	)

	s.Run()
	m.Run()
}
