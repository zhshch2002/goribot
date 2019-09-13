package goribotExts

import (
	"encoding/json"
	"fmt"
	"github.com/zhshch2002/goribot"
	"net/url"
	"testing"
)

func TestAllowHostPipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewAllowHostPipeline("github.com"))
	got := false
	err := s.Get(nil, "https://www.baidu.com/", func(r *goribot.Response) {
		t.Error("TestAllowHostPipeline error")
	})
	if err != nil {
		t.Error(err)
	}
	err = s.Get(nil, "https://github.com/", func(r *goribot.Response) {
		got = true
	})
	if err != nil {
		t.Error(err)
	}
	s.Run()
	if !got {
		t.Error("TestAllowHostPipeline error")
	}
}

func TestDisallowHostPipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewDisallowHostPipeline("github.com"))
	got := false
	err := s.Get(nil, "https://github.com/", func(r *goribot.Response) {
		t.Error("TestDisallowHostPipeline error")
	})
	if err != nil {
		t.Error(err)
	}
	err = s.Get(nil, "https://www.baidu.com/", func(r *goribot.Response) {
		got = true
	})
	if err != nil {
		t.Error(err)
	}
	s.Run()
	if !got {
		t.Error("TestDisallowHostPipeline error")
	}
}

func TestDeduplicatePipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewDeduplicatePipeline())
	got := false
	err := s.Get(nil, "https://www.baidu.com/", func(r *goribot.Response) {
		got = true
	})
	if err != nil {
		t.Error(err)
	}

	err = s.Get(nil, "https://www.baidu.com/", func(r *goribot.Response) {
		t.Error("TestDeduplicatePipeline error")
	})
	if err != nil {
		t.Error(err)
	}

	s.Run()
	if !got {
		t.Error("TestDeduplicatePipeline error")
	}
}

func TestMaxRequestPipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewMaxRequestPipeline(1))
	got := false
	err := s.Get(nil, "https://www.github.com/", func(r *goribot.Response) {
		got = true
	})
	if err != nil {
		t.Error(err)
	}

	err = s.Get(nil, "https://www.baidu.com/", func(r *goribot.Response) {
		t.Error("TestMaxRequestPipeline error")
	})
	if err != nil {
		t.Error(err)
	}

	s.Run()
	if !got {
		t.Error("TestMaxRequestPipeline error")
	}
}

func TestUrlFilterPipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewUrlFilterPipeline("github"))
	got := false
	err := s.Get(nil, "https://github.com/", func(r *goribot.Response) {
		got = true
	})
	if err != nil {
		t.Error(err)
	}

	err = s.Get(nil, "https://www.baidu.com/", func(r *goribot.Response) {
		t.Error("TestMaxRequestPipeline error")
	})
	if err != nil {
		t.Error(err)
	}

	s.Run()
	if !got {
		t.Error("TestMaxRequestPipeline error")
	}
}

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

func TestRobotstxtPipeline(t *testing.T) {
	s := goribot.NewSpider()
	s.Use(NewRobotstxtPipeline("https://github.com/"))
	err := s.Get(nil, "https://github.com/zhshch2002", func(r *goribot.Response) { // unable to access according to https://github.com/robots.txt
		t.Error("RobotstxtPipeline error")
	})
	if err != nil {
		t.Error(err)
	}
	s.Run()

	s = goribot.NewSpider()
	s.UserAgent = "Googlebot"
	got := false
	s.Use(NewRobotstxtPipeline("https://github.com/"))
	err = s.Get(nil, "https://github.com/zhshch2002/goribot/wiki", func(r *goribot.Response) {
		got = true
	})
	if err != nil {
		t.Error(err)
	}
	s.Run()
	if !got {
		t.Error("RobotstxtPipeline error")
	}
}
