package goribotExts

import (
	"github.com/zhshch2002/goribot"
	"regexp"
)

type UrlFilterPipeline struct {
	goribot.Pipeline
	rex *regexp.Regexp
}

func NewUrlFilterPipeline(urlrex string) *UrlFilterPipeline {
	r, err := regexp.Compile(urlrex)

	if err != nil {
		panic(err)
	}
	return &UrlFilterPipeline{rex: r}
}

func (s *UrlFilterPipeline) OnRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	if ok := s.rex.MatchString(request.Url.String()); !ok {
		return nil
	}
	return request
}
