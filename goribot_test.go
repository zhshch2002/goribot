package goribot

import (
	"net/http"
	"testing"
)

func TestBasic(t *testing.T) {
	s := NewSpider()
	got := false
	s.NewTask(MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello%20world"), func(ctx *Context) {
		t.Log("got resp data", ctx.Text)
		if ctx.Json["args"].(map[string]interface{})["Goribot test"].(string) != "hello world" {
			t.Error("wrong resp data")
		} else {
			got = true
		}
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestCookie(t *testing.T) {
	s := NewSpider()
	got := false
	r := MustNewGetReq("https://httpbin.org/cookies")
	r.Cookie = append(r.Cookie, &http.Cookie{
		Name:  "Goribot test",
		Value: "hello world",
	})
	s.NewTask(r, func(ctx *Context) {
		t.Log("got resp data", ctx.Text)
		if ctx.Json["cookies"].(map[string]interface{})["Goribot test"].(string) != "hello world" {
			t.Error("wrong resp data")
		} else {
			got = true
		}
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestUrlencodedPost(t *testing.T) {
	s := NewSpider()
	got := false
	s.NewTask(MustNewPostReq(
		"https://httpbin.org/post",
		UrlencodedPostData,
		map[string]string{
			"Goribot test": "hello world",
		}),
		func(ctx *Context) {
			t.Log("got resp data", ctx.Text)
			if ctx.Json["form"].(map[string]interface{})["Goribot test"].(string) != "hello world" {
				t.Error("wrong resp data")
			} else {
				got = true
			}
		})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestJsonPost(t *testing.T) {
	s := NewSpider()
	got := false
	s.NewTask(MustNewPostReq(
		"https://httpbin.org/post", JsonPostData, map[string]string{
			"Goribot test": "hello world",
		}),
		func(ctx *Context) {
			t.Log("got resp data", ctx.Text)
			if ctx.Json["json"].(map[string]interface{})["Goribot test"].(string) != "hello world" {
				t.Error("wrong resp data")
			} else {
				got = true
			}
		})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestTextPost(t *testing.T) {
	s := NewSpider()
	got := false
	s.NewTask(MustNewPostReq(
		"https://httpbin.org/post", TextPostData, "hello world"),
		func(ctx *Context) {
			t.Log("got resp data", ctx.Text)
			if ctx.Json["data"].(string) != "hello world" {
				t.Error("wrong resp data")
			} else {
				got = true
			}
		})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestCtxNewReq(t *testing.T) {
	s := NewSpider()
	got := false
	s.NewTask(MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello%20world"), func(ctx *Context) {
		ctx.NewTask(MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello%20world"), func(ctx *Context) {
			t.Log("got resp data", ctx.Text)
			if ctx.Json["args"].(map[string]interface{})["Goribot test"].(string) != "hello world" {
				t.Error("wrong resp data")
			} else {
				got = true
			}
		})
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestOnHandlers(t *testing.T) {
	s := NewSpider()
	resp, task, item, onerr := false, false, false, false
	got := false
	s.AddRespHandler(func(ctx *Context) {
		t.Log("on resp")
		resp = true
	})
	s.AddTaskHandler(func(ctx *Context, k *Task) *Task {
		t.Log("on task", t)
		task = true
		return k
	})
	s.AddItemHandler(func(ctx *Context, i interface{}) interface{} {
		t.Log("on item", i)
		item = true
		return i
	})
	s.AddErrorHandler(func(ctx *Context, err error) {
		t.Log("on error", err)
		onerr = true
	})
	s.NewTask(MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello%20world"), func(ctx *Context) {
		got = true
		ctx.AddItem(1)
	})
	s.NewTask(MustNewGetReq("/"), func(ctx *Context) {})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
	if (!resp) || (!task) || (!item) || (!onerr) {
		t.Error("handlers func wrong")
	}
}

func TestBFS(t *testing.T) {
	s := NewSpider()
	s.ThreadPoolSize = 1
	s.DepthFirst = false
	got := false

	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		t.Log("got No.1")
		if got {
			t.Error("wrong request order")
		}
		got = true
	})
	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		t.Log("got No.2")
		got = true
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestDFS(t *testing.T) {
	s := NewSpider()
	s.ThreadPoolSize = 1
	s.DepthFirst = true
	got := false

	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		got = true
	})
	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		t.Log("got No.1")
		if got {
			t.Error("wrong request order")
		}
		t.Log("got No.2")
		got = true
	})
	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}

func TestCtxDrop(t *testing.T) {
	s := NewSpider()

	s.AddRespHandler(func(ctx *Context) {
		ctx.Drop()
	})

	s.NewTask(MustNewGetReq("https://httpbin.org/get"), func(ctx *Context) {
		t.Error("drop error")
	})

	s.Run()
}

func TestTaskWithMeta(t *testing.T) {
	s := NewSpider()

	s.AddRespHandler(func(ctx *Context) {
		if d, ok := ctx.Meta["test"]; !ok {
			t.Error("can't find meta")
		} else {
			t.Log("Meta 'test' data", d)
		}

	})

	got := false
	s.NewTaskWithMeta(MustNewGetReq("https://httpbin.org/get"), map[string]interface{}{
		"test": 1,
	}, func(ctx *Context) {
		ctx.NewTaskWithMeta(MustNewGetReq("https://httpbin.org/get"), map[string]interface{}{
			"test": 2,
		}, func(ctx *Context) {
			got = true
		})
	})

	s.Run()
	if !got {
		t.Error("didn't get data")
	}
}
