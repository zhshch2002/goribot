package goribot

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"sync/atomic"
	"time"
)

// DefaultUA is the default User-Agent of spider
const DefaultUA = "Goribot"

// Spider is the core spider struct
type Spider struct {
	ThreadPoolSize uint64
	DepthFirst     bool
	Downloader     func(r *Request) (*Response, error)

	Cache *CacheManger

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
		Cache:          NewCacheManger(),
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
