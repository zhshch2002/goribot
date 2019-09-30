package goribot

import (
	"net/http"
	"testing"
)

func TestNet(t *testing.T) {
	req := &Request{
		Url:    MustParseUrl("https://httpbin.org/get?Goribot%20test=hello%20world"),
		Method: http.MethodGet,
		Cookie: nil,
		Header: nil,
		Body:   nil,
		Proxy:  "",
	}
	resp, err := Download(req)
	if err != nil {
		t.Error(err)
	}
	t.Log("got resp data", resp.Text)
	if resp.Json["args"].(map[string]interface{})["Goribot test"].(string) != "hello world" {
		t.Error("wrong resp data")
	}
}
