package goribotExts

import (
	"github.com/zhshch2002/goribot"
	"sync/atomic"
)

type MaxRequestPipeline struct {
	goribot.Pipeline
	RequestCount, MaxRequestLimit uint64
}

func NewMaxRequestPipeline(maxRequestLimit uint64) *MaxRequestPipeline {
	return &MaxRequestPipeline{MaxRequestLimit: maxRequestLimit}
}
func (s *MaxRequestPipeline) OnDoRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	atomic.AddUint64(&s.RequestCount, 1)
	if atomic.LoadUint64(&s.RequestCount) > s.MaxRequestLimit {
		return nil
	}
	return request
}
