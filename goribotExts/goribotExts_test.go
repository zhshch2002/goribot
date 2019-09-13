package goribotExts

import (
	"encoding/json"
	"fmt"
	"github.com/zhshch2002/goribot"
	"net/url"
	"testing"
)

func TestRefererPipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewRefererPipeline())
	u, _ := url.Parse("https://github.com/")
	err := s.Get(&goribot.Response{
		Request: &goribot.Request{
			Url: u,
		},
	}, "https://httpbin.org/headers", func(r *goribot.Response) {
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(r.Text), &m)
		if err != nil {
			t.Error("json load error", err)
		}
		if m["headers"].(map[string]interface{})["Referer"].(string) != "https://github.com/" {
			fmt.Println(r.Text)
			t.Error("error")
		}
		t.Log("RefererPipeline test ok")
	})
	if err != nil {
		t.Error(err)
	}
	s.Run()
}

func TestRandomUaPipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewRandomUaPipeline())
	err := s.Get(nil, "https://httpbin.org/user-agent", func(r *goribot.Response) {
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(r.Text), &m)
		if err != nil {
			t.Error("json load error", err)
		}
		if m["user-agent"].(string) == s.UserAgent {
			t.Error(
				"got:", "'"+m["user-agent"].(string)+"'")
		}
		t.Log("RandomUaPipeline test ok")
	})
	if err != nil {
		t.Error(err)
	}
	s.Run()
}
