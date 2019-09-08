package goribot

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	Method  string
	Url     url.URL
	Header  Dict
	Cookie  Dict
	Body    []byte
	Proxies string
	Timeout time.Duration
}

func NewGetRequest(rawurl string) (*Request, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return &Request{
		Method:  "GET",
		Url:     *u,
		Timeout: 5 * time.Second,
	}, nil
}

func NewPostRequest(rawurl string, data []byte, contentType string) (*Request, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}
	return &Request{
		Method: "GET",
		Url:    *u,
		Header: Dict{
			"Content-Type": contentType,
		},
		Body:    data,
		Timeout: 5 * time.Second,
	}, nil
}

func NewPostData(data Dict) []byte {
	var DataTmp []string
	for k, v := range data {
		DataTmp = append(DataTmp, k+"="+v)
	}
	return []byte(strings.Join(DataTmp, "&"))
}

type Response struct {
	Request      *Request
	HttpResponse *http.Response
	Status       int
	Headers      http.Header
	Body         []byte
	Text         string
}

func DoRequest(r *Request) (*Response, error) {
	client := &http.Client{
		Timeout: r.Timeout,
	}
	if r.Body == nil {
		r.Body = []byte{}
	}
	request, err := http.NewRequest(r.Method, r.Url.String(), bytes.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	if r.Header != nil {
		for k, v := range r.Header {
			request.Header.Set(k, v)
		}
	}

	if r.Cookie != nil {
		var cookieTmp []string
		for k, v := range r.Cookie {
			cookieTmp = append(cookieTmp, k+"="+v)
		}
		request.Header.Set("Cookie", strings.Join(cookieTmp, "; "))
	}

	if r.Proxies != "" {
		p, err := url.Parse(r.Proxies)
		if err != nil {
			return nil, err
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(p),
		}
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Response{
		Request:      r,
		HttpResponse: resp,
		Status:       resp.StatusCode,
		Headers:      resp.Header,
		Body:         body,
		Text:         string(body),
	}, nil
}
