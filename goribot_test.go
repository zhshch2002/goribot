package goribot

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	s := NewSpider()
	r := 0
	s.OnStart(func(s *Spider) {
		t.Log("OnStart")
		r += 1
	})
	s.OnAdd(func(ctx *Context, ta *Task) *Task {
		t.Log("OnAdd")
		r += 1
		return ta
	})
	s.OnReq(func(ctx *Context, req *Request) *Request {
		t.Log("OnReq")
		r += 1
		return req
	})
	s.OnResp(func(ctx *Context) {
		t.Log("OnResp")
		r += 1
	})
	s.AddTask(
		GetReq("https://httpbin.org/get").SetParam(map[string]string{
			"Goribot test": "hello world",
		}).WithMeta("test", "hello world"),
		func(ctx *Context) {
			if ctx.Meta["test"] != "hello world" {
				t.Error("wrong meta data")
			}
			r += 1
			t.Log("got resp data", ctx.Resp.Text)
			if ctx.Resp.Json("args.Goribot test").String() != "hello world" {
				t.Error("wrong resp data: " + ctx.Resp.Json("args.Goribot test").String() + " " + ctx.Resp.Text)
			}
			ctx.AddItem(struct{}{})
		},
		func(ctx *Context) {
			t.Log("Handler 2")
			panic("some test error")
		},
	)
	s.OnItem(func(i interface{}) interface{} {
		t.Log("OnItem")
		r += 1
		//panic("unexpect error")
		return i
	})
	s.OnError(func(ctx *Context, err error) {
		t.Log(err)
		r += 1
	})
	s.OnFinish(func(s *Spider) {
		t.Log("OnFinish")
		r += 1
	})
	s.Run()
	if r != 8 {
		t.Error("handlers miss " + fmt.Sprint(r))
	}
}

func TestAbort(t *testing.T) {
	s := NewSpider()
	got := false
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			got = true
			t.Log("got resp data")
			ctx.Abort()
		},
		func(ctx *Context) {
			t.Error("Abort error")
		},
	)
	s.Run()
	if !got {
		t.Error("didn't get response")
	}
}
