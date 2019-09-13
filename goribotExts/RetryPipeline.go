package goribotExts

import (
	"github.com/zhshch2002/goribot"
	"log"
)

type RetryPipeline struct {
	goribot.Pipeline
	MaxRetryTimes int
	ErrCode       map[int]struct{}
}

func NewRetryPipeline(maxRetryTimes int) *RetryPipeline {
	if maxRetryTimes == 0 {
		maxRetryTimes = 2
	}
	p := &RetryPipeline{MaxRetryTimes: maxRetryTimes}
	p.ErrCode = make(map[int]struct{})
	for _, c := range []int{500, 503, 504, 400, 408} {
		p.ErrCode[c] = struct{}{}
	}
	return p
}

func NewRetryPipelineWithErrorCode(maxRetryTimes int, errCode ...int) *RetryPipeline {
	p := NewRetryPipeline(maxRetryTimes)
	for _, c := range []int{500, 503, 504, 400, 408} {
		p.ErrCode[c] = struct{}{}
	}
	for _, c := range errCode {
		p.ErrCode[c] = struct{}{}
	}
	return p
}

func (s *RetryPipeline) retry(spider *goribot.Spider, r *goribot.Request) {
	if _, ok := r.Meta["RetryTimes"]; ok {
		if r.Meta["RetryTimes"].(int) > s.MaxRetryTimes {
			log.Println(r.Url.String(), "retry times out")
			return
		}
		r.Meta["RetryTimes"] = r.Meta["RetryTimes"].(int) + 1
	} else {
		r.Meta["RetryTimes"] = 1
	}
	spider.Crawl(r)
}

func (s *RetryPipeline) OnError(spider *goribot.Spider, err error) {
	if e, ok := err.(goribot.HttpErr); ok {
		s.retry(spider, e.Request)
	}
}

func (s *RetryPipeline) OnResponse(spider *goribot.Spider, response *goribot.Response) *goribot.Response {
	if _, ok := s.ErrCode[response.HttpResponse.StatusCode]; ok {
		s.retry(spider, response.Request)
		return nil
	}
	return response
}
