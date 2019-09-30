package goribot

import (
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
)

type PostDataType int

const (
	TextPostData       PostDataType = iota // text/plain
	UrlencodedPostData                     // application/x-www-form-urlencoded
	JsonPostData                           // application/json
)

type Request struct {
	Url    *url.URL
	Method string
	Cookie []*http.Cookie
	Header http.Header
	Body   []byte
	Proxy  string
}

func NewRequest() *Request {
	return &Request{
		Url:    &url.URL{},
		Method: "GET",
		Cookie: []*http.Cookie{},
		Header: http.Header{},
		Body:   []byte{},
		Proxy:  "",
	}
}

type Response struct {
	Url    *url.URL
	Status int
	Header http.Header
	Body   []byte

	Request      *Request
	HttpResponse *http.Response

	Text string
	Html *goquery.Document
	Json map[string]interface{}
}

func Download(r *Request) (*Response, error) {
	HttpRequest, err := http.NewRequest(r.Method, r.Url.String(), bytes.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	if r.Header != nil {
		HttpRequest.Header = r.Header.Clone()
	}
	for _, i := range r.Cookie {
		HttpRequest.AddCookie(i)
	}

	c := &http.Client{}
	if r.Proxy != "" {
		p, err := url.Parse(r.Proxy)
		if err != nil {
			return nil, err
		}
		c.Transport = &http.Transport{
			Proxy: http.ProxyURL(p),
		}
	}

	resp, err := c.Do(HttpRequest)
	if err != nil {
		return nil, HttpErr{error: err, Request: r}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))
	var j map[string]interface{}
	_ = json.Unmarshal(body, &j)

	return &Response{
		Url:          r.Url,
		Status:       resp.StatusCode,
		Header:       resp.Header,
		Body:         body,
		Request:      r,
		HttpResponse: resp,
		Text:         string(body),
		Html:         html,
		Json:         j,
	}, nil
}

type HttpErr struct {
	error
	Request *Request
}

func MustParseUrl(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}
