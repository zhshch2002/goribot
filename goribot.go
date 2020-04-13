package goribot

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"github.com/panjf2000/ants/v2"
	"os"
	"runtime"
	"time"
)

var Log = logging.MustGetLogger("goribot")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfile} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func init() {
	backend := logging.NewLogBackend(os.Stdout, "Goribot ", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
}

var ErrRunFinishedSpider = errors.New("running a spider which is finished,you could recreate this spider and run the new one")

type Task struct {
	Request  *Request
	Handlers []CtxHandlerFun
}

func NewTask(request *Request, handlers ...CtxHandlerFun) *Task {
	return &Task{Request: request, Handlers: handlers}
}

type CtxHandlerFun func(ctx *Context)

type Spider struct {
	Scheduler                         Scheduler
	Downloader                        Downloader
	AutoStop                          bool
	taskPool, itemPool                *ants.Pool
	onStartHandlers, onFinishHandlers []func(s *Spider)
	onReqHandlers                     []func(ctx *Context, req *Request) *Request
	onAddHandlers                     []func(ctx *Context, req *Task) *Task
	onRespHandlers                    []CtxHandlerFun
	onItemHandlers                    []func(i interface{}) interface{}
	onErrorHandlers                   []func(ctx *Context, err error)
}

func NewSpider(exts ...func(s *Spider)) *Spider {
	tp, err := ants.NewPool(runtime.NumCPU() * 2)
	if err != nil {
		panic(err)
	}
	ip, err := ants.NewPool(runtime.NumCPU())
	if err != nil {
		panic(err)
	}
	s := &Spider{
		Scheduler:  NewBaseScheduler(false),
		Downloader: NewBaseDownloader(),
		taskPool:   tp,
		itemPool:   ip,
		AutoStop:   true,
	}
	s.Use(exts...)
	return s
}

func (s *Spider) SetTaskPoolSize(i int) {
	s.taskPool.Tune(i)
}

func (s *Spider) SetItemPoolSize(i int) {
	s.itemPool.Tune(i)
}

func (s *Spider) AddTask(request *Request, handlers ...CtxHandlerFun) {
	if request.Depth == -1 {
		request.Depth = 1
	}
	t := NewTask(request, handlers...)
	t = s.handleOnAdd(nil, t)
	if t != nil {
		s.Scheduler.AddTask(t)
	}

}

func (s *Spider) Use(fn ...func(s *Spider)) {
	for _, f := range fn {
		f(s)
	}
}

func (s *Spider) Run() {
	defer s.taskPool.Release()
	defer s.itemPool.Release()
	s.handleOnStart()
	taskRunning := true
	if s.itemPool.Cap() > 0 {
		go func() {
			for taskRunning {
				if s.itemPool.Free() > 0 {
					if i := s.Scheduler.GetItem(); i != nil {
						err := s.itemPool.Submit(func() {
							s.handleOnItem(i)
						})
						if errors.Is(err, ants.ErrPoolClosed) {
							panic(ErrRunFinishedSpider)
						}
					}
				} else {
					time.Sleep(500 * time.Microsecond)
				}
				runtime.Gosched()
			}
		}()
	}

	for {
		if s.taskPool.Free() > 0 {
			if t := s.Scheduler.GetTask(); t != nil {
				err := s.taskPool.Submit(func() {
					ctx := &Context{
						Req:      t.Request,
						Resp:     nil,
						tasks:    []*Task{},
						items:    []interface{}{},
						Meta:     t.Request.Meta,
						Handlers: t.Handlers,
						abort:    false,
					}
					defer func() { // 回收Task和Item
						defer func() { // 回收时的错误处理
							if err := recover(); err != nil {
								s.handleOnError(ctx, errors.New(fmt.Sprintf("%+v", err)))
							}
						}()
						for _, i := range ctx.tasks {
							i := s.handleOnAdd(ctx, i)
							if i != nil {
								if !i.Request.URL.IsAbs() {
									i.Request.URL = ctx.Resp.Request.URL.ResolveReference(i.Request.URL)
								}
								if i.Request.Depth == -1 {
									i.Request.Depth = ctx.Req.Depth + 1
								}
								s.Scheduler.AddTask(i)
							}
						}
						for _, i := range ctx.items {
							s.Scheduler.AddItem(i)
						}
					}()
					defer func() { // 主回调函数异常处理
						if err := recover(); err != nil {
							s.handleOnError(ctx, errors.New(fmt.Sprintf("%+v", err)))
						}
					}()
					req := s.handleOnReq(ctx, t.Request)
					if req.Err != nil {
						s.handleOnError(ctx, req.Err)
						return
					}
					if req != nil {
						resp, err := s.Downloader.Do(req)
						ctx.Resp = resp
						if err == nil {
							ctx.Meta = resp.Meta
							if ctx.Resp.Text == "" {
								_ = ctx.Resp.DecodeAndParse()
							}
							s.handleOnResp(ctx)
							for _, fn := range t.Handlers {
								if ctx.IsAborted() {
									break
								}
								fn(ctx)
							}
						} else {
							s.handleOnError(ctx, err)
						}
					}
				})
				if errors.Is(err, ants.ErrPoolClosed) {
					panic(ErrRunFinishedSpider)
				}
			} else if s.taskPool.Running() == 0 {
				if s.AutoStop {
					break
				} else {
					Log.Info("Waiting for more tasks")
					time.Sleep(5 * time.Second)
				}
			}
		} else {
			if s.Scheduler.IsTaskEmpty() {
				time.Sleep(500 * time.Microsecond)
			}
		}
		runtime.Gosched()
	}
	taskRunning = false
	s.handleOnFinish()
}

/*************************************************************************************/
func (s *Spider) OnStart(fn func(s *Spider)) {
	s.onStartHandlers = append(s.onStartHandlers, fn)
}
func (s *Spider) handleOnStart() {
	for _, fn := range s.onStartHandlers {
		fn(s)
	}
}

/*************************************************************************************/
func (s *Spider) OnFinish(fn func(s *Spider)) {
	s.onFinishHandlers = append(s.onFinishHandlers, fn)
}
func (s *Spider) handleOnFinish() {
	for _, fn := range s.onFinishHandlers {
		fn(s)
	}
}

/*************************************************************************************/
func (s *Spider) OnReq(fn func(ctx *Context, req *Request) *Request) {
	s.onReqHandlers = append(s.onReqHandlers, fn)
}
func (s *Spider) handleOnReq(ctx *Context, req *Request) *Request {
	for _, fn := range s.onReqHandlers {
		req = fn(ctx, req)
		if req == nil {
			return req
		}
	}
	return req
}

/*************************************************************************************/
func (s *Spider) OnAdd(fn func(ctx *Context, t *Task) *Task) {
	s.onAddHandlers = append(s.onAddHandlers, fn)
}
func (s *Spider) handleOnAdd(ctx *Context, t *Task) *Task {
	for _, fn := range s.onAddHandlers {
		t = fn(ctx, t)
		if t == nil {
			return t
		}
	}
	return t
}

/*************************************************************************************/
func (s *Spider) OnResp(fn CtxHandlerFun) {
	s.onRespHandlers = append(s.onRespHandlers, fn)
}
func (s *Spider) handleOnResp(ctx *Context) {
	for _, fn := range s.onRespHandlers {
		if ctx.IsAborted() {
			Log.Info("Aborted")
			return
		}
		fn(ctx)
	}
}

/*************************************************************************************/
func (s *Spider) OnItem(fn func(i interface{}) interface{}) {
	s.onItemHandlers = append(s.onItemHandlers, fn)
}
func (s *Spider) handleOnItem(i interface{}) {
	for _, fn := range s.onItemHandlers {
		i = fn(i)
		if i == nil {
			return
		}
	}
}

/*************************************************************************************/
func (s *Spider) OnError(fn func(ctx *Context, err error)) {
	s.onErrorHandlers = append(s.onErrorHandlers, fn)
}
func (s *Spider) handleOnError(ctx *Context, err error) {
	for _, fn := range s.onErrorHandlers {
		fn(ctx, err)
	}
}
