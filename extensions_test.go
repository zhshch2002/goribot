package goribot

import (
	"testing"
)

func TestRandomUserAgent(t *testing.T) {
	s := NewSpider(
		RandomUserAgent,
	)
	got := false
	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		t.Log("got resp data", ctx.Text)
		if ctx.Json["headers"].(map[string]interface{})["User-Agent"].(string) == DefaultUA {
			t.Error("wrong ua setting")
		} else {
			got = true
		}
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}
