package goribot

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestSavers(t *testing.T) {
	var cvsFile, jsonFile *os.File
	if os.Getenv("DISABLE_SAVER_TEST") != "" {
		cvsFile, _ = os.Open("/dev/null")
		jsonFile, _ = os.Open("/dev/null")
	} else {
		var err error
		cvsFile, err = os.Create("./test.cvs")
		if err != nil {
			panic(err)
		}
		jsonFile, err = os.Create("./test.json")
		if err != nil {
			panic(err)
		}

	}
	defer cvsFile.Close()
	defer jsonFile.Close()
	s := NewSpider(
		SaveItemsAsCSV(cvsFile),
		SaveItemsAsJSON(jsonFile),
	)
	s.AddTask(
		GetReq("https://httpbin.org"),
		func(ctx *Context) {
			ctx.AddItem(CsvItem{
				ctx.Resp.Request.URL.String(),
				ctx.Resp.Dom.Find("title").Text(),
			})
			ctx.AddItem(JsonItem{Data: map[string]interface{}{
				"url":   ctx.Resp.Request.URL.String(),
				"title": ctx.Resp.Dom.Find("title").Text(),
			}})
			ctx.AddItem(CsvItem{
				ctx.Resp.Request.URL.String(),
				ctx.Resp.Dom.Find("title").Text(),
			})
			ctx.AddItem(JsonItem{Data: map[string]interface{}{
				"url":   ctx.Resp.Request.URL.String(),
				"title": ctx.Resp.Dom.Find("title").Text(),
			}})
		},
	)
	s.Run()
}

func TestSpiderLogError(t *testing.T) {
	s := NewSpider(
		SpiderLogError(os.Stdout),
	)
	s.Downloader.(*BaseDownloader).Client.Timeout = 5 * time.Second
	s.AddTask(GetReq("https://httpbin.org/get"), func(ctx *Context) {
		panic("some error!")
	})
	s.AddTask(GetReq("https://httpbin.org/get"), func(ctx *Context) {
		ctx.AddItem(ErrorItem{
			Ctx: ctx,
			Msg: "I left a message.",
		})
	})
	s.AddTask(GetReq("https://githab.com/"))
	s.Run()
}

func TestRetry(t *testing.T) {
	ti := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ti += 1
		if ti < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, _ = fmt.Fprintf(w, "Hello goribot")
	}))
	defer ts.Close()

	got := 0
	s := NewSpider(
		Retry(3, http.StatusOK),
	)
	s.Downloader.(*BaseDownloader).Client.Timeout = 1 * time.Second
	s.AddTask(
		GetReq(ts.URL),
		func(ctx *Context) {
			Log.Info(ctx.Resp.Text, ctx.IsAborted())
			if ctx.Resp.Text == "Hello goribot" {
				got += 1
			}
		},
	)

	s.Run()
	if ti != 3 {
		t.Error("Retry times wrong", ti)
	}
	if got != 1 {
		t.Error("wrong response", got)
	}
}

func TestRobotsTxt(t *testing.T) {
	s := NewSpider(
		RobotsTxt("https://github.com", "Goribot"),
	)
	s.AddTask( // unable to access according to https://github.com/robots.txt
		GetReq("https://github.com/zhshch2002"),
		func(ctx *Context) {
			t.Error("RobotsTxt error")
		},
	)
	s.Run()

	got := false
	s = NewSpider(
		RobotsTxt("https://github.com/", "Googlebot"),
	)
	s.AddTask( // unable to access according to https://github.com/robots.txt
		GetReq("https://github.com/zhshch2002/goribot/wiki"),
		func(ctx *Context) {
			got = true
		},
	)
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

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
	s.SetTaskPoolSize(1)
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
		GetReq("https://httpbin.org/get").SetParam(map[string]string{
			"name": "Goribot",
		}),
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
		GetReq("https://httpbin.org/get").SetParam(map[string]string{
			"name": "Goribot",
		}),
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
			if ctx.Resp.Json("headers.User-Agent").String() == "Go-http-Client/2.0" {
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

func TestSpiderLogPrint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello goribot")
	}))
	defer ts.Close()
	s := NewSpider(SpiderLogPrint())
	s.SetTaskPoolSize(2)
	i := 0
	for i < 20 {
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
}
