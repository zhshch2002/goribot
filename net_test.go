package goribot

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNet(t *testing.T) {
	d := NewBaseDownloader()

	resp, err := d.Do(
		GetReq("https://httpbin.org/get").
			SetParam(map[string]string{
				"Goribot": "hello world",
			}).
			SetHeader("Goribot", "hello world").
			SetUA("Goribot").
			AddCookie(&http.Cookie{
				Name:  "Goribot",
				Value: "hello world",
			}),
	)
	if err != nil {
		t.Error(err)
	}
	t.Log("got resp data", resp.Text)
	if resp.Json("args.Goribot").String() != "hello world" {
		t.Error("wrong resp data: " + resp.Json("args.Goribot").String())
	}
	if resp.Json("headers.Goribot").String() != "hello world" {
		t.Error("wrong resp data: " + resp.Json("headers.Goribot").String())
	}
	if resp.Json("headers.Cookie").String() != `Goribot="hello world"` {
		t.Error("wrong resp data: " + resp.Json("headers.Cookie").String())
	}
}

func TestPost(t *testing.T) {
	d := NewBaseDownloader()

	resp, err := d.Do(
		PostRawReq("https://httpbin.org/post", []byte("hello world")),
	)
	if err != nil {
		t.Error(err)
	}
	t.Log("got resp data", resp.Text)
	if resp.Json("data").String() != "hello world" {
		t.Error("wrong resp data: " + resp.Json("data").String())
	}

	resp, err = d.Do(
		PostFormReq("https://httpbin.org/post", map[string]string{
			"Goribot": "hello world",
		}),
	)
	if err != nil {
		t.Error(err)
	}
	t.Log("got resp data", resp.Text)
	if resp.Json("form.Goribot").String() != "hello world" {
		t.Error("wrong resp data: " + resp.Json("form.Goribot").String())
	}

	resp, err = d.Do(
		PostJsonReq("https://httpbin.org/post", map[string]interface{}{
			"Goribot": "hello world",
		}),
	)
	if err != nil {
		t.Error(err)
	}
	t.Log("got resp data", resp.Text)
	if resp.Json("json.Goribot").String() != "hello world" {
		t.Error("wrong resp data: " + resp.Json("json.Goribot").String())
	}
}

func TestNetDecode(t *testing.T) {
	d := NewBaseDownloader()
	resp, err := d.Do(GetReq("http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2017/45/14/25/451425202.html"))
	if err != nil {
		t.Error(err)
	}
	t.Log("got resp data", resp.Text)
	if !strings.Contains(resp.Text, "统计用区划代码") {
		t.Error("wrong resp data")
	}
}

func TestMiddleware(t *testing.T) {
	d := NewBaseDownloader()
	got := false
	d.AddMiddleware(func(req *Request, next func(req *Request) (resp *Response, err error)) (resp *Response, err error) {
		resp, err = next(req)
		got = true
		if err == nil {
			fmt.Println(resp.Text)
		}
		return resp, err
	})
	resp, err := d.Do(GetReq("https://httpbin.org/get"))
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(resp.Text)
	}
	if !got {
		t.Error("middleware error")
	}
}

func TestCookieJar(t *testing.T) {
	i := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i == 0 {
			i += 1
			return
		}
		if cookie, err := r.Cookie("Flavor"); err != nil {
			http.SetCookie(w, &http.Cookie{Name: "Flavor", Value: "Chocolate Chip"})
		} else {
			cookie.Value = "Oatmeal Raisin"
			http.SetCookie(w, cookie)
		}
	}))
	d := NewBaseDownloader()
	resp, _ := d.Do(GetReq(ts.URL))
	fmt.Println(resp.Cookies())
	resp, _ = d.Do(GetReq(ts.URL))
	fmt.Println(resp.Cookies())
	resp, _ = d.Do(GetReq(ts.URL))
	fmt.Println(resp.Cookies())
}
