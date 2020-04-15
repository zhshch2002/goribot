package goribot

import (
	"github.com/go-redis/redis"
	"os"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	if os.Getenv("DISABLE_SAVER_TEST") == "" {
		return
	}
	ro := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	sName := "DistributedTest"
	gotItem := false
	m := NewManager(redis.NewClient(ro), sName)
	m.OnItem(func(i interface{}) interface{} {
		Log.Info("got item")
		if i == "goribot" {
			gotItem = true
		}
		return i
	})
	m.SendReq(GetReq("https://httpbin.org/get").SetHeader("goribot", "hello world"))

	gotResp := false
	s := NewSpider(
		RedisDistributed(
			ro,
			sName,
			true,
			func(ctx *Context) {
				Log.Info("got seed resp")
				gotResp = true
				ctx.AddItem("goribot")
			},
		),
	)

	go s.Run()
	go m.Run()
	time.Sleep(10 * time.Second)

	if !gotResp {
		t.Error("lost resp")
	} else if !gotItem {
		t.Error("lost item")
	}
}
