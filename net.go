package goribot

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	Method  string
	Url     url.URL
	Header  http.Header
	Body    []byte
	Proxies string
	Timeout time.Duration
	Handler []ResponseHandler
}

func NewGetRequest(rawurl string) (*Request, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return &Request{
		Method:  "GET",
		Url:     *u,
		Header:  http.Header{},
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
		Header: http.Header{
			"Content-Type": {contentType},
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
	Html         *goquery.Document
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
		for k, vl := range r.Header {
			for _, v := range vl {
				request.Header.Add(k, v)
			}
		}
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

	html, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))

	return &Response{
		Request:      r,
		HttpResponse: resp,
		Status:       resp.StatusCode,
		Headers:      resp.Header,
		Body:         body,
		Text:         string(body),
		Html:         html,
	}, nil
}
