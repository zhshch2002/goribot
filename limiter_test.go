package goribot

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello goribot")
	})
	Log.Info("Server Start")
	go func() {
		err := http.ListenAndServe("127.0.0.1:1229", nil)
		if err != nil {
			Log.Error("ListenAndServe: ", err)
		}
	}()
}

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
	start := time.Now()
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob: "127.0.0.1:1229",
			Rate: 2,
		}),
	)
	i := 0
	for i < 10 {
		ii := i
		s.AddTask(
			GetReq("http://127.0.0.1:1229/"),
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
	start := time.Now()
	s := NewSpider(
		Limiter(true, &LimitRule{
			Glob:        "127.0.0.1:1229",
			Parallelism: 1,
		}),
	)
	i := 0
	for i < 5 {
		ii := i
		s.AddTask(
			GetReq("http://127.0.0.1:1229/"),
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
