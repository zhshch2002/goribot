package goribot

import (
	"testing"
	"time"
)

func TestLimiterDelay(t *testing.T) {
	start := time.Now()
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob: "httpbin.org",
			//Allow: Allow,
			Delay: 5 * time.Second,
		}),
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			Log.Info("got 1")
		},
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			Log.Info("got 2")
		},
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			Log.Info("got 3")
		},
	)
	s.AddTask(
		GetReq("https://github.com"),
		func(ctx *Context) {
			t.Error("shouldn't get")
		},
	)
	s.Run()
	if time.Since(start) < 10*time.Second {
		t.Error("wrong time")
	}
}

func TestLimiterRate(t *testing.T) {
	//start := time.Now()
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob: "httpbin.org",
			Rate: 2,
		}),
	)
	i := 0
	for i < 20 {
		ii := i
		s.AddTask(
			GetReq("https://httpbin.org/get"),
			func(ctx *Context) {
				Log.Info("got", ii)
			},
		)
		i += 1
	}
	s.Run()
	//if time.Since(start) < 5*time.Second {
	//	t.Error("wrong time")
	//}
}

func TestLimiterParallelism(t *testing.T) {
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob:        "httpbin.org",
			Parallelism: 1,
		}),
	)
	got := 0
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			if got != 0 {
				t.Error("wrong order")
			}
			got = 1
			Log.Info("got", 1)
		},
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			if got != 1 {
				t.Error("wrong order")
			}
			got = 2
			Log.Info("got", 2)
		},
	)
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			if got != 2 {
				t.Error("wrong order")
			}
			got = 3
			Log.Info("got", 3)
		},
	)
	s.Run()
	if got != 3 {
		t.Error("lost response")
	}
}
