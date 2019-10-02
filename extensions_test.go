package goribot

import (
	"net/http"
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

func TestHostFilter(t *testing.T) {
	s := NewSpider(
		HostFilter("www.baidu.com"),
	)
	got := false
	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		t.Error("got wrong resp")
	})
	s.NewTask(MustNewGetReq("https://www.baidu.com/"), func(ctx *Context) {
		t.Log("got resp data", ctx.Text)
		got = true
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestRobotsTxt(t *testing.T) {
	s := NewSpider(
		RobotsTxt("https://github.com/", "Goribot"),
	)
	s.NewTask(MustNewGetReq("https://github.com/zhshch2002"), func(ctx *Context) { // unable to access according to https://github.com/robots.txt
		t.Error("RobotsTxt error")
	})
	s.Run()

	s = NewSpider(
		RobotsTxt("https://github.com/", "Googlebot"),
	)
	got := false
	s.NewTask(MustNewGetReq("https://github.com/zhshch2002/goribot/wiki"), func(ctx *Context) {
		got = true
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestReqDeduplicate(t *testing.T) {
	got1, got2 := false, false
	s := NewSpider(
		ReqDeduplicate(),
	)
	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		got1 = true
		t.Log("got first")
		r := MustNewGetReq("https://httpbin.org/get")
		r.Cookie = append(r.Cookie, &http.Cookie{
			Name:  "123",
			Value: "123",
		})
		ctx.NewTask(r, func(ctx *Context) {
			t.Log("got second")
			got2 = true
		})
	})
	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		t.Error("Deduplicate error")
	})
	s.Run()
	if (!got1) || (!got2) {
		t.Error("didn't get data")
	}
}
