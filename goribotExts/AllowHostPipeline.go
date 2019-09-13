package goribotExts

import (
	"github.com/zhshch2002/goribot"
	"strings"
)

type AllowHostPipeline struct {
	goribot.Pipeline
	Hosts map[string]struct{}
}

func NewAllowHostPipeline(hosts ...string) *AllowHostPipeline {
	p := &AllowHostPipeline{}
	p.Hosts = make(map[string]struct{})
	for _, h := range hosts {
		p.Hosts[strings.ToLower(h)] = struct{}{}
	}
	return p
}

func (s *AllowHostPipeline) OnRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	if _, ok := s.Hosts[strings.ToLower(request.Url.Host)]; !ok {
		return nil
	}
	return request
}

type DisallowHostPipeline struct {
	goribot.Pipeline
	Hosts map[string]struct{}
}

func NewDisallowHostPipeline(hosts ...string) *AllowHostPipeline {
	p := &AllowHostPipeline{}
	p.Hosts = make(map[string]struct{})
	for _, h := range hosts {
		p.Hosts[strings.ToLower(h)] = struct{}{}
	}
	return p
}

func (s *DisallowHostPipeline) OnRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	if _, ok := s.Hosts[strings.ToLower(request.Url.Host)]; ok {
		return nil
	}
	return request
}
