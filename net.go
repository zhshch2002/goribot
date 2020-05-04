package goribot

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/saintfish/chardet"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// DownloaderErr is a error create by Downloader
type DownloaderErr struct {
	error
	// Request is the Request object when the error occurred
	Request *Request
	// Response is the Request object when the error occurred.It could be nil.
	Response *Response
}

// Deprecated: will be remove at next major version
var GetReq = Get

// Get creates a get request
func Get(urladdr string) *Request {
	req, err := http.NewRequest("GET", urladdr, nil)
	return &Request{
		Request:                   req,
		Depth:                     -1,
		ResponseCharacterEncoding: "",
		ProxyURL:                  "",
		Meta:                      map[string]interface{}{},
		Err:                       err,
	}
}

// Deprecated: will be remove at next major version
var PostReq = Post

// Post creates a post request
func Post(urladdr string, body io.Reader) *Request {
	req, err := http.NewRequest("POST", urladdr, body)
	return &Request{
		Request:                   req,
		Depth:                     -1,
		ResponseCharacterEncoding: "",
		ProxyURL:                  "",
		Meta:                      map[string]interface{}{},
		Err:                       err,
	}
}

// PostReq creates a post request with raw data
func PostRawReq(urladdr string, body []byte) *Request {
	return PostReq(urladdr, bytes.NewReader(body))
}

// PostFormReq creates a post request with form data
func PostFormReq(urladdr string, requestData map[string]string) *Request {
	var urlS url.URL
	q := urlS.Query()
	for k, v := range requestData {
		q.Add(k, v)
	}
	req := PostRawReq(urladdr, []byte(q.Encode()))
	req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// PostJsonReq creates a post request with json data
func PostJsonReq(urladdr string, requestData interface{}) *Request {
	body, err := json.Marshal(requestData)
	req := PostReq(urladdr, bytes.NewReader(body))
	if req.Err == nil {
		req.Err = err
	}
	req.SetHeader("Content-Type", "application/json")
	return req
}

// Request is a object of HTTP request
type Request struct {
	*http.Request
	Depth int
	// ResponseCharacterEncoding is the character encoding of the response body.
	// Leave it blank to allow automatic character encoding of the response body.
	// It is empty by default and it can be set in OnRequest callback.
	ResponseCharacterEncoding string
	// ProxyURL is the proxy address that handles the request
	ProxyURL string
	// Meta contains data between a Request and a Response
	Meta map[string]interface{}
	Err  error

	body []byte
}

// GetBody returns the body as bytes of request
func (s *Request) GetBody() []byte {
	if s.Err == nil {
		if s.Request.Body == nil {
			return []byte{}
		}
		s.body, _ = ioutil.ReadAll(s.Request.Body)
		s.Request.Body = ioutil.NopCloser(bytes.NewReader(s.body))
	}
	return []byte{}
}

// AddCookie adds a cookie to the request.
func (s *Request) AddCookie(c *http.Cookie) *Request {
	if s.Err == nil {
		s.Request.AddCookie(c)
	}
	return s
}

// SetHeader sets the header entries associated with key
// to the single element value.
func (s *Request) SetHeader(key, value string) *Request {
	if s.Err == nil {
		s.Request.Header.Set(key, value)
	}
	return s
}

// SetProxy sets proxy url of request.
func (s *Request) SetProxy(p string) *Request {
	if s.Err == nil {
		s.ProxyURL = p
	}
	return s
}

// SetProxy sets user-agent url of request header.
func (s *Request) SetUA(ua string) *Request {
	if s.Err == nil {
		s.SetHeader("User-Agent", ua)
	}
	return s
}

// AddParam adds a query param of request url.
func (s *Request) AddParam(k, v string) *Request {
	if s.Err == nil {
		if len(s.Request.URL.RawQuery) > 0 {
			s.Request.URL.RawQuery += "&"
		}
		s.Request.URL.RawQuery += url.QueryEscape(k) + "=" + url.QueryEscape(v)
	}
	return s
}

// SetParam sets query param of request url.
// Deprecated: will be remove at next major version
func (s *Request) SetParam(p map[string]string) *Request {
	for k, v := range p {
		s.AddParam(k, v)
	}
	return s
}

// SetParam sets the meta data of request.
func (s *Request) WithMeta(k string, v interface{}) *Request {
	s.Meta[k] = v
	return s
}

// Response is a object of HTTP response
type Response struct {
	*http.Response
	// Body is the content of the Response
	Body []byte
	// Text is the content of the Response parsed as string
	Text string
	// Request is the Req object from goribot of the response.Tip: there is another Request attr come from *http.Response
	Req *Request
	// Dom is the parsed html object
	Dom *goquery.Document
	// Meta contains data between a Request and a Response
	Meta map[string]interface{}
}

// DecodeAndParas decodes the body to text and try to parse it to html or json.
func (s *Response) DecodeAndParse() error {
	if len(s.Body) == 0 {
		return nil
	}
	contentType := strings.ToLower(s.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "/json") {
		if !strings.Contains(contentType, "charset") {
			if s.Req.ResponseCharacterEncoding != "" {
				contentType += "; charset=" + s.Req.ResponseCharacterEncoding
			} else {
				r, err := chardet.NewTextDetector().DetectBest(s.Body)
				if err != nil {
					return err
				}
				contentType += "; charset=" + r.Charset
			}
		}
		if strings.Contains(contentType, "utf-8") || strings.Contains(contentType, "utf8") {
			s.Text = string(s.Body)
		} else {
			tmpBody, err := encodeBytes(s.Body, contentType)
			if err != nil {
				return err
			}
			s.Body = tmpBody
			s.Text = string(s.Body)
		}
		if s.IsHTML() {
			d, err := goquery.NewDocumentFromReader(bytes.NewReader(s.Body))
			s.Dom = d
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Json returns json result parsed from response
func (s *Response) Json(q string) gjson.Result {
	return gjson.Get(s.Text, q)
}

func (s *Response) IsHTML() bool {
	contentType := strings.ToLower(s.Header.Get("Content-Type"))
	return strings.Contains(contentType, "/html")
}

func (s *Response) IsJSON() bool {
	contentType := strings.ToLower(s.Header.Get("Content-Type"))
	return strings.Contains(contentType, "/json")
}

// Downloader tool download response from request
type Downloader interface {
	Do(req *Request) (resp *Response, err error)
	AddMiddleware(func(req *Request, next func(req *Request) (resp *Response, err error)) (resp *Response, err error))
}

var D = NewBaseDownloader()
var Do = D.Do

// BaseDownloader is default downloader of goribot
type BaseDownloader struct {
	Client   *http.Client
	handlers []func(req *Request, next func(req *Request) (resp *Response, err error)) (resp *Response, err error)
}

func NewBaseDownloader() *BaseDownloader {
	j, _ := cookiejar.New(nil)
	return &BaseDownloader{Client: &http.Client{Jar: j}}
}

func (s *BaseDownloader) AddMiddleware(fn func(req *Request, next func(*Request) (*Response, error)) (*Response, error)) {
	s.handlers = append(s.handlers, fn)
}

func (s *BaseDownloader) defaultHandler(req *Request) (resp *Response, err error) {
	if req.Err != nil {
		return nil, err
	}
	client := s.Client

	if req.ProxyURL != "" {
		s.Client.Transport = &http.Transport{
			Proxy: func(request *http.Request) (u *url.URL, err error) {
				return url.Parse(req.ProxyURL)
			},
		}
	}
	res, err := client.Do(req.Request)
	if err != nil {
		return nil, DownloaderErr{err, req, resp}
	}
	defer res.Body.Close()

	resp = &Response{
		Response: res,
		Text:     "",
		Req:      req,
		Meta:     req.Meta,
	}

	bodyReader := res.Body
	contentEncoding := strings.ToLower(res.Header.Get("Content-Encoding"))
	if !res.Uncompressed && (strings.Contains(contentEncoding, "gzip") || (contentEncoding == "" && strings.Contains(strings.ToLower(res.Header.Get("Content-Type")), "gzip")) || strings.HasSuffix(strings.ToLower(req.URL.Path), ".xml.gz")) {
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			return nil, DownloaderErr{err, req, resp}
		}
		defer bodyReader.(*gzip.Reader).Close()
	}

	resp.Body, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, DownloaderErr{err, req, resp}
	}
	_ = resp.DecodeAndParse()
	return resp, nil
}

func (s *BaseDownloader) nextHandler(i int) func(req *Request) (resp *Response, err error) {
	if i == -1 {
		return s.defaultHandler
	}
	return func(req *Request) (resp *Response, err error) {
		return s.handlers[i](req, s.nextHandler(i-1))
	}
}

func (s *BaseDownloader) Do(req *Request) (resp *Response, err error) {
	return s.nextHandler(len(s.handlers) - 1)(req)
}
