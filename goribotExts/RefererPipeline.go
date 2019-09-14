package goribotExts

import (
	"github.com/zhshch2002/goribot"
)

type RefererPipeline struct {
	goribot.Pipeline
}

func NewRefererPipeline() *RefererPipeline {
	return &RefererPipeline{}
}

func (s *RefererPipeline) OnNewRequest(spider *goribot.Spider, preResp *goribot.Response, request *goribot.Request) *goribot.Request {
	if preResp != nil {
		request.Header.Set("Referer", preResp.Request.Url.String())
	}
	return request
}
