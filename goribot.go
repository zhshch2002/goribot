package goribot

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	UserAgent      = "GoRibot" //TODO 设置UA
	ThreadPoolSize = uint(15)
)

type PostDataType int

const (
	_                  PostDataType = iota
	TextPostData                    // text/plain
	UrlencodedPostData              // application/x-www-form-urlencoded
	JsonPostData                    // application/json
)

type HtmlHandler struct {
	Selector string
	UrlReg   *regexp.Regexp
	fun      func(DOM *HTMLElement)
}

type ResponseHandler struct {
	UrlReg *regexp.Regexp
	fun    func(DOM *Response)
}
type TaskQueue struct {
	sync.Mutex
	items []*Request
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		Mutex: sync.Mutex{},
	}
}

func (s *TaskQueue) Push(item *Request) {
	s.Lock()
	s.items = append(s.items, item)
	s.Unlock()
}
func (s *TaskQueue) PushInHead(item *Request) {
	s.Lock()
	s.items = append([]*Request{item}, s.items...)
	s.Unlock()
}
func (s *TaskQueue) Pop() *Request {
	s.Lock()
	item := s.items[0]
	s.items = s.items[1:]
	s.Unlock()
	return item
}

func (s *TaskQueue) IsEmpty() bool {
	return len(s.items) == 0
}

type Spider struct {
	UserAgent      string
	MaxVisit       uint //TODO 实现
	MaxDepth       uint //TODO 实现
	ThreadPoolSize uint
	DepthFirst     bool
	Downloader     func(*Request) (*Response, error)

	pipeline     []PipelineInterface
	taskQueue    *TaskQueue
	taskChan     chan *Request
	taskFinished bool
	wg           sync.WaitGroup

	onHtmlHandlers     []HtmlHandler
	onResponseHandlers []ResponseHandler
	workingThread      int32
}

func NewSpider() *Spider {
	return &Spider{
		taskQueue:      NewTaskQueue(),
		Downloader:     DoRequest,
		UserAgent:      UserAgent,
		ThreadPoolSize: ThreadPoolSize,
	}
}

func (s *Spider) Run() {
	if s.ThreadPoolSize == 0 {
		s.ThreadPoolSize = ThreadPoolSize
	}
	s.taskFinished = false
	s.taskChan = make(chan *Request)
	for i := uint(0); i < s.ThreadPoolSize; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for !s.taskFinished {
				select {
				case req := <-s.taskChan:
					func() {
						defer func() { s.workingThread -= 1 }()
						resp, err := s.Downloader(req)
						if err != nil {
							log.Println("Downloader Error", err, req.Url.String())
							s.handleOnErrorPipeline(err)
							return
						}
						resp = s.handleOnResponsePipeline(resp)
						if resp == nil {
							return
						}
						s.handleResponse(resp)
					}()
					break
				default:
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()
	}
	for {
		if s.taskQueue.IsEmpty() {
			if s.workingThread == 0 { // make sure the queue is empty and no threat is working
				break
			} else {
				time.Sleep(1 * time.Millisecond)
			}
		} else {
			s.taskChan <- s.taskQueue.Pop()
			s.workingThread += 1
		}
	}
	s.taskFinished = true
	s.wg.Wait()
}
func (s *Spider) handleResponse(response *Response) {
	for _, h := range s.onResponseHandlers {
		if h.UrlReg != nil {
			if ok := h.UrlReg.MatchString(response.HttpResponse.Request.URL.String()); !ok {
				continue
			}
		}
		h.fun(response)
	}

	if doc, err := goquery.NewDocumentFromReader(strings.NewReader(response.Text)); err == nil {
		for _, h := range s.onHtmlHandlers {
			if h.UrlReg != nil {
				if ok := h.UrlReg.MatchString(response.HttpResponse.Request.URL.String()); !ok {
					continue
				}
			}
			doc.Find(h.Selector).Each(func(i int, s *goquery.Selection) {
				for _, n := range s.Nodes {
					h.fun(NewHTMLElementFromSelectionNode(response, s, n, i))
				}
			})
		}
	}
}

// Add on-event handler
func (s *Spider) OnResponse(fun func(Resp *Response)) {
	s.onResponseHandlers = append(s.onResponseHandlers, ResponseHandler{
		UrlReg: nil,
		fun:    fun,
	})
}
func (s *Spider) OnUrlResponse(urlreg string, fun func(Resp *Response)) error {
	r, err := regexp.Compile(urlreg)
	if err != nil {
		return err
	}
	s.onResponseHandlers = append(s.onResponseHandlers, ResponseHandler{
		UrlReg: r,
		fun:    fun,
	})
	return nil
}
func (s *Spider) OnHTML(selector string, fun func(DOM *HTMLElement)) {
	s.onHtmlHandlers = append(s.onHtmlHandlers, HtmlHandler{
		Selector: selector,
		UrlReg:   nil,
		fun:      fun,
	})
}
func (s *Spider) OnUrlHTML(urlreg, selector string, fun func(DOM *HTMLElement)) error {
	r, err := regexp.Compile(urlreg)
	if err != nil {
		return err
	}
	s.onHtmlHandlers = append(s.onHtmlHandlers, HtmlHandler{
		Selector: selector,
		UrlReg:   r,
		fun:      fun,
	})
	return nil
}

// Add a new task to the queue
func (s *Spider) Crawl(r *Request) {
	r.Header.Set("User-Agent", s.UserAgent)
	r = s.handleOnRequestPipeline(r)
	if r == nil {
		return
	}

	if s.DepthFirst {
		s.taskQueue.PushInHead(r)
	} else {
		s.taskQueue.Push(r)
	}
}
func (s *Spider) Get(u string) error {
	req, err := NewGetRequest(u)
	if err != nil {
		return err
	}
	s.Crawl(req)
	return nil
}
func (s *Spider) Post(u string, datatype PostDataType, rawdata interface{}) error { //TODO Post func
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
			return err
		}
		data = tdata
		break
	}
	req, err := NewPostRequest(u, data, ct)
	if err != nil {
		return err
	}
	s.Crawl(req)
	return nil
}

// Pipeline handlers and register
func (s *Spider) Use(p PipelineInterface) {
	p.Init(s)
	s.pipeline = append(s.pipeline, p)
}
func (s *Spider) handleInitPipeline() {
	for _, p := range s.pipeline {
		p.Init(s)
	}
}
func (s *Spider) handleOnRequestPipeline(r *Request) *Request {
	for _, p := range s.pipeline {
		r = p.OnRequest(s, r)
		if r == nil {
			return nil
		}
	}
	return r
}
func (s *Spider) handleOnResponsePipeline(r *Response) *Response {
	for _, p := range s.pipeline {
		r = p.OnResponse(s, r)
		if r == nil {
			return nil
		}
	}
	return r
}
func (s *Spider) handleOnErrorPipeline(err error) {
	for _, p := range s.pipeline {
		p.OnError(s, err)
	}
}
func (s *Spider) NewItem(item interface{}) {
	for _, p := range s.pipeline {
		item = p.OnItem(s, item)
		if item == nil {
			return
		}
	}
}
