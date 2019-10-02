package goribot

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const DefaultUA = "Goribot"

type Context struct {
	Text string
	Html *goquery.Document
	Json map[string]interface{}

	Request  *Request
	Response *Response

	Tasks []*Task
	Items []interface{}
	Meta  map[string]interface{}

	drop bool
}

func (c *Context) Drop() {
	c.drop = true
}
func (c *Context) IsDrop() bool {
	return c.drop
}

func (c *Context) AddItem(i interface{}) {
	c.Items = append(c.Items, i)
}
func (c *Context) AddTask(r *Task) {
	c.Tasks = append(c.Tasks, r)
}
func (c *Context) NewTask(req *Request, RespHandler ...func(ctx *Context)) {
	c.AddTask(NewTask(req, RespHandler...))
}
func (c *Context) NewTaskWithMeta(req *Request, meta map[string]interface{}, RespHandler ...func(ctx *Context)) {
	t := NewTask(req, RespHandler...)
	t.Meta = meta
	c.Tasks = append(c.Tasks, t)

}

type Task struct {
	Request        *Request
	onRespHandlers []func(ctx *Context)
	Meta           map[string]interface{}
}

func NewTask(req *Request, RespHandler ...func(ctx *Context)) *Task {
	return &Task{Request: req, onRespHandlers: RespHandler}
}

type Spider struct {
	ThreadPoolSize uint64
	DepthFirst     bool
	Downloader     func(r *Request) (*Response, error)

	onRespHandlers  []func(ctx *Context)
	onTaskHandlers  []func(ctx *Context, req *Task) *Task
	onItemHandlers  []func(ctx *Context, i interface{}) interface{}
	onErrorHandlers []func(ctx *Context, err error)

	taskQueue     *TaskQueue
	workingThread uint64
}

func NewSpider(exts ...func(s *Spider)) *Spider {
	s := &Spider{
		taskQueue:      NewTaskQueue(),
		Downloader:     Download,
		DepthFirst:     true,
		ThreadPoolSize: 30,
	}
	for _, e := range exts {
		e(s)
	}
	return s
}

func (s *Spider) Run() {
	worker := func(t *Task) {
		defer atomic.AddUint64(&s.workingThread, ^uint64(0))

		resp, err := s.Downloader(t.Request)
		ctx := &Context{
			Request:  t.Request,
			Response: resp,
			Items:    []interface{}{},
			Meta:     t.Meta,
			drop:     false,
		}
		if err != nil {
			log.Println("Downloader error", err)
			s.handleError(ctx, err)
		} else {
			ctx.Text = resp.Text
			ctx.Html = resp.Html
			ctx.Json = resp.Json

			s.handleResp(ctx)
			if !ctx.IsDrop() {
				for _, h := range t.onRespHandlers {
					h(ctx)
					if ctx.IsDrop() {
						break
					}
				}
			}
		}

		for _, i := range ctx.Tasks {
			i = s.handleTask(ctx, i)
			if i != nil {
				s.AddTask(ctx, i)
			}
		}
		s.handleItem(ctx)
	}

	for (!s.taskQueue.IsEmpty()) || atomic.LoadUint64(&s.workingThread) > 0 {
		if (!s.taskQueue.IsEmpty()) && (atomic.LoadUint64(&s.workingThread) < s.ThreadPoolSize || s.ThreadPoolSize == 0) {
			atomic.AddUint64(&s.workingThread, 1)
			go worker(s.taskQueue.Pop())
		} else {
			time.Sleep(100 * time.Nanosecond)
		}
	}
}

func (s *Spider) AddTask(ctx *Context, t *Task) {
	t = s.handleTask(ctx, t)
	if t == nil {
		return
	}

	if t.Request.Header.Get("User-Agent") == "" {
		t.Request.Header.Set("User-Agent", DefaultUA)
	}

	if s.DepthFirst {
		s.taskQueue.PushInHead(t)
	} else {
		s.taskQueue.Push(t)
	}
}

var TodoContext = &Context{
	Text:     "",
	Html:     &goquery.Document{},
	Json:     map[string]interface{}{},
	Request:  &Request{},
	Response: &Response{},
	Tasks:    []*Task{},
	Items:    []interface{}{},
	Meta:     map[string]interface{}{},
	drop:     false,
}

func (s *Spider) NewTask(req *Request, RespHandler ...func(ctx *Context)) {
	s.AddTask(TodoContext, NewTask(req, RespHandler...))
}
func (s *Spider) NewTaskWithMeta(req *Request, meta map[string]interface{}, RespHandler ...func(ctx *Context)) {
	t := NewTask(req, RespHandler...)
	t.Meta = meta
	s.AddTask(TodoContext, t)
}

func (s *Spider) handleResp(ctx *Context) {
	for _, h := range s.onRespHandlers {
		h(ctx)
		if ctx.IsDrop() == true {
			return
		}
	}
}
func (s *Spider) handleTask(ctx *Context, t *Task) *Task {
	for _, h := range s.onTaskHandlers {
		t = h(ctx, t)
		if t == nil {
			return nil
		}
	}
	return t
}
func (s *Spider) handleItem(ctx *Context) {
	for _, h := range s.onItemHandlers {
		for _, i := range ctx.Items {
			i = h(ctx, i)
			if i == nil {
				return
			}
		}
	}
}
func (s *Spider) handleError(ctx *Context, err error) {
	for _, h := range s.onErrorHandlers {
		h(ctx, err)
	}
}

func (s *Spider) OnResp(h func(ctx *Context)) {
	s.onRespHandlers = append(s.onRespHandlers, h)
}
func (s *Spider) OnTask(h func(ctx *Context, t *Task) *Task) {
	s.onTaskHandlers = append(s.onTaskHandlers, h)
}
func (s *Spider) OnItem(h func(ctx *Context, i interface{}) interface{}) {
	s.onItemHandlers = append(s.onItemHandlers, h)
}
func (s *Spider) OnError(h func(ctx *Context, err error)) {
	s.onErrorHandlers = append(s.onErrorHandlers, h)
}

func NewGetReq(rawurl string) (*Request, error) {
	req := NewRequest()
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req.Url = u
	req.Method = http.MethodGet
	return req, nil
}
func MustNewGetReq(rawurl string) *Request {
	res, err := NewGetReq(rawurl)
	if err != nil {
		panic(err)
	}
	return res
}

func NewPostReq(rawurl string, datatype PostDataType, rawdata interface{}) (*Request, error) {
	req := NewRequest()
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req.Url = u
	req.Method = http.MethodPost

	var data []byte
	ct := ""
	switch datatype {
	case TextPostData:
		ct = "text/plain"
		data = []byte(rawdata.(string))
		break
	case UrlencodedPostData:
		ct = "application/x-www-form-urlencoded"
		var urlS url.URL
		q := urlS.Query()
		for k, v := range rawdata.(map[string]string) {
			q.Add(k, v)
		}
		data = []byte(q.Encode())
		break
	case JsonPostData:
		ct = "application/json"
		tdata, err := json.Marshal(rawdata)
		if err != nil {
			return nil, err
		}
		data = tdata
		break
	}

	req.Header.Set("Content-Type", ct)
	req.Body = data

	return req, nil
}
func MustNewPostReq(rawurl string, datatype PostDataType, rawdata interface{}) *Request {
	res, err := NewPostReq(rawurl, datatype, rawdata)
	if err != nil {
		panic(err)
	}
	return res
}
