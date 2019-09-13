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
	Url     *url.URL
	Header  http.Header
	Body    []byte
	Proxies string
	Meta    map[string]interface{}
	Handler []ResponseHandler
}

type Response struct {
	Request      *Request
	HttpResponse *http.Response
	Status       int
	Headers      http.Header
	Body         []byte
	Text         string
	Meta         map[string]interface{}
	Html         *goquery.Document
}

var DefaultClient = &http.Client{
	Jar:     nil,
	Timeout: 5 * time.Second,
}

func DoRequest(r *Request) (*Response, error) {
	client := DefaultClient
	if r.Body == nil {
		r.Body = []byte{}
	}
	HttpRequest, err := http.NewRequest(r.Method, r.Url.String(), bytes.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	if r.Header != nil {
		HttpRequest.Header = r.Header
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

	resp, err := client.Do(HttpRequest)
	if err != nil {
		return nil, HttpErr{error: err, Request: r}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))

	res := &Response{
		Request:      r,
		HttpResponse: resp,
		Status:       resp.StatusCode,
		Headers:      resp.Header,
		Body:         body,
		Text:         string(body),
		Html:         html,
		Meta:         map[string]interface{}{},
	}
	for key, value := range r.Meta {
		res.Meta[key] = value
	}
	return res, nil
}

func NewRequest(method string, rawurl string, body []byte) (*Request, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return &Request{
		Method: method,
		Url:    u,
		Body:   body,
		Header: http.Header{},
		Meta:   map[string]interface{}{},
	}, nil
}

func NewGetRequest(rawurl string) (*Request, error) {
	return NewRequest("GET", rawurl, []byte{})
}

func NewPostRequest(rawurl string, data []byte, contentType string) (*Request, error) {
	r, err := NewRequest("POST", rawurl, data)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}
	r.Header.Set("Content-Type", contentType)
	return r, nil
}

func NewPostData(data Dict) []byte {
	var DataTmp []string
	for k, v := range data {
		DataTmp = append(DataTmp, k+"="+v)
	}
	return []byte(strings.Join(DataTmp, "&"))
}

type HttpErr struct {
	error
	Request *Request
}
