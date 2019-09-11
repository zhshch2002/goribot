package goribotExts

import (
	"github.com/zhshch2002/goribot"
	"strings"
)

type allowHostPipeline struct {
	goribot.Pipeline
	Hosts map[string]struct{}
}

func NewAllowHostPipeline(hosts ...string) *allowHostPipeline {
	p := &allowHostPipeline{}
	p.Hosts = make(map[string]struct{})
	for _, h := range hosts {
		p.Hosts[strings.ToLower(h)] = struct{}{}
	}
	return p
}

func (s *allowHostPipeline) OnRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	if _, ok := s.Hosts[strings.ToLower(request.Url.Host)]; !ok {
		return nil
	}
	return request
}
