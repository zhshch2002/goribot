package goribot

import (
	"testing"
)

func TestRefererFiller(t *testing.T) {
	s := NewSpider(
		RefererFiller(),
	)
	got1 := false
	got2 := false
	s.AddTask(
		GetReq("https://httpbin.org/"),
		func(ctx *Context) {
			got1 = true
			t.Log("got first")
			ctx.AddTask(
				GetReq("https://httpbin.org/get").SetHeader("123", "ABC"),
				func(ctx *Context) {
					t.Log("got second")
					got2 = true
					if ctx.Resp.Json("headers.Referer").String() != "https://httpbin.org/" {
						t.Error("wrong Referer", ctx.Resp.Json("headers.Referer").String())
					}
				},
			)
		},
	)
	s.Run()
	if !got1 || !got2 {
		t.Error("didn't get data")
	}
}

func TestSetDepthFirst(t *testing.T) {
	got1, got2 := false, false
	s := NewSpider(
		SetDepthFirst(true),
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			got1 = true
			t.Log("got first")
		},
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			got2 = true
			if got1 {
				t.Error("wrong task order")
			}
			t.Log("got second")
		},
	)
	s.Run()
	if (!got1) || (!got2) {
		t.Error("didn't get data")
	}
}

func TestReqDeduplicate(t *testing.T) {
	got1, got2 := false, false
	s := NewSpider(
		ReqDeduplicate(),
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			got1 = true
			t.Log("got first")
			ctx.AddTask(
				GetReq("https://httpbin.org/get").SetHeader("123", "ABC"),
				func(ctx *Context) {
					t.Log("got second")
					got2 = true
				},
			)
		},
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			t.Error("Deduplicate error")
		},
	)
	s.Run()
	if (!got1) || (!got2) {
		t.Error("didn't get data")
	}
}

func TestRandomUserAgent(t *testing.T) {
	s := NewSpider(
		RandomUserAgent(),
	)
	got := false
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			t.Log("got resp data", ctx.Resp.Text)
			if ctx.Resp.Json("headers.User-Agent").String() == "Go-http-client/2.0" {
				t.Error("wrong ua setting")
			} else {
				got = true
			}
		},
	)
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}
