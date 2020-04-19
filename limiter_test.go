package goribot

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	if time.Since(start) <= 10*time.Second {
		t.Error("wrong time")
	}
}

func TestLimiterRate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello goribot")
	}))
	defer ts.Close()
	start := time.Now()
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob: "*",
			Rate: 2,
		}),
	)
	i := 0
	for i < 10 {
		ii := i
		s.AddTask(
			GetReq(ts.URL),
			func(ctx *Context) {
				Log.Info("got", ii)
			},
		)
		i += 1
	}
	s.Run()
	if time.Since(start) <= 5*time.Second {
		t.Error("wrong time")
	}
}

func TestLimiterParallelism(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello goribot")
	}))
	defer ts.Close()
	start := time.Now()
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob:        "*",
			Parallelism: 1,
		}),
	)
	i := 0
	for i < 5 {
		ii := i
		s.AddTask(
			GetReq(ts.URL),
			func(ctx *Context) {
				Log.Info("got", ii)
			},
		)
		i += 1
	}
	s.Run()
	if time.Since(start) <= 20*time.Second {
		t.Error("wrong time")
	}
}

func TestMaxReq(t *testing.T) {
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob:   "httpbin.org",
			MaxReq: 3,
		}),
	)
	got := 0
	i := 0
	for i < 5 {
		ii := i
		s.AddTask(
			GetReq("https://httpbin.org/get"),
			func(ctx *Context) {
				Log.Info("got", ii)
				got += 1
			},
		)
		i += 1
	}
	s.Run()
	if got != 3 {
		t.Error("wrong req got", got)
	}
}

func TestMaxDepth(t *testing.T) {
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob:     "httpbin.org",
			MaxDepth: 2,
		}),
	)
	got := 0
	s.AddTask(
		GetReq("https://httpbin.org/get"),
		func(ctx *Context) {
			got += 1
			ctx.AddTask(GetReq("https://httpbin.org/get"), func(ctx *Context) {
				got += 1
				ctx.AddTask(GetReq("https://httpbin.org/get"), func(ctx *Context) {
					got += 1
					ctx.AddTask(GetReq("https://httpbin.org/get"), func(ctx *Context) {
						got += 1
					})
				})
			})
		},
	)

	s.Run()
	if got != 2 {
		t.Error("wrong req got", got)
	}
}
