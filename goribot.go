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

// Default User-Agent of spider
const DefaultUA = "Goribot"

// Context is a wrap of response,origin request,new task,etc
type Context struct {
	Text string                 // the response text
	Html *goquery.Document      // spider will try to parse the response as html
	Json map[string]interface{} // spider will try to parse the response as json

	Request  *Request  // origin request
	Response *Response // a response object

	Tasks []*Task                // the new request task which will send to the spider
	Items []interface{}          // the new result data which will send to the spider，use to store
	Meta  map[string]interface{} // the request task created by NewTaskWithMeta func will have a k-y pair

	drop bool // in handlers chain,you can use ctx.Drop() to break the handler chain and stop handling
}

// Drop this context to break the handler chain and stop handling
func (c *Context) Drop() {
	c.drop = true
}

// IsDrop return was the context dropped
func (c *Context) IsDrop() bool {
	return c.drop
}

// AddItem add an item to new item list. After every handler func return,
// spider will collect these items and call OnItem handler func
func (c *Context) AddItem(i interface{}) {
	c.Items = append(c.Items, i)
}

// AddTask add a task to new task list. After every handler func return,spider will collect these tasks
func (c *Context) AddTask(r *Task) {
	c.Tasks = append(c.Tasks, r)
}

// NewTask create a task and add it to new task list After every handler func return,spider will collect these tasks
func (c *Context) NewTask(req *Request, RespHandler ...func(ctx *Context)) {
	c.AddTask(NewTask(req, RespHandler...))
}

// NewTaskWithMeta create a task with meta data and add it to new task list After every handler func return,
// spider will collect these tasks
func (c *Context) NewTaskWithMeta(req *Request, meta map[string]interface{}, RespHandler ...func(ctx *Context)) {
	t := NewTask(req, RespHandler...)
	t.Meta = meta
	c.Tasks = append(c.Tasks, t)

}

// Task is a wrap of request and its handler funcs
type Task struct {
	Request        *Request
	onRespHandlers []func(ctx *Context)
	Meta           map[string]interface{}
}

// NewTask create a new task
func NewTask(req *Request, RespHandler ...func(ctx *Context)) *Task {
	return &Task{Request: req, onRespHandlers: RespHandler}
}

// Spider is the core spider struct
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

// NewSpider create a new spider and run extension func to config the spider
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

// Run the spider and wait to all task done
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
			s.AddTask(ctx, i)
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

// AddTask add a task to the queue
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

// TodoContext -- If a task created by `spider.NewTask` as seed task,the OnTask handler will get TodoContext as ctx param
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

// NewTask create a task and add it to the queue
func (s *Spider) NewTask(req *Request, RespHandler ...func(ctx *Context)) {
	s.AddTask(TodoContext, NewTask(req, RespHandler...))
}

// NewTaskWithMeta create a task with meta data and add it to the queue
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

// OnResp add an On Response handler func to the spider
func (s *Spider) OnResp(h func(ctx *Context)) {
	s.onRespHandlers = append(s.onRespHandlers, h)
}

// OnTask add an On New Task handler func to the spider
func (s *Spider) OnTask(h func(ctx *Context, t *Task) *Task) {
	s.onTaskHandlers = append(s.onTaskHandlers, h)
}

// OnItem add an On New Item handler func to the spider. For some storage
func (s *Spider) OnItem(h func(ctx *Context, i interface{}) interface{}) {
	s.onItemHandlers = append(s.onItemHandlers, h)
}

// OnError add an On Error handler func to the spider
func (s *Spider) OnError(h func(ctx *Context, err error)) {
	s.onErrorHandlers = append(s.onErrorHandlers, h)
}

// NewGetReq create a new get request
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

// MustNewGetReq  create a new get request,if get error will do panic
func MustNewGetReq(rawurl string) *Request {
	res, err := NewGetReq(rawurl)
	if err != nil {
		panic(err)
	}
	return res
}

// NewPostReq create a new post request
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

	req.SetHeader("Content-Type", ct).SetBody(data)

	return req, nil
}

// MustNewPostReq create a new post request,if get error will do panic
func MustNewPostReq(rawurl string, datatype PostDataType, rawdata interface{}) *Request {
	res, err := NewPostReq(rawurl, datatype, rawdata)
	if err != nil {
		panic(err)
	}
	return res
}
