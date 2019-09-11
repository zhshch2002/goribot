package goribotExts

import (
	"github.com/zhshch2002/goribot"
	"regexp"
)

type urlFilterPipeline struct {
	goribot.Pipeline
	rex *regexp.Regexp
}

func NewUrlFilterPipeline(urlrex string) *urlFilterPipeline {
	r, err := regexp.Compile(urlrex)

	if err != nil {
		panic(err)
	}
	return &urlFilterPipeline{rex: r}
}

func (s *urlFilterPipeline) OnRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	if ok := s.rex.MatchString(request.Url.String()); !ok {
		return nil
	}
	return request
}
