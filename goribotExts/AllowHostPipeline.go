package goribotExts

import (
	"github.com/zhshch2002/goribot"
	"strings"
)

/*
type MyPipeline struct {
	goribot.Pipeline
}

func NewMyPipeline() *MyPipeline {
	return &MyPipeline{}
}

func (s *MyPipeline) Init(spider *goribot.Spider) {}
func (s *MyPipeline) OnNewRequest(spider *goribot.Spider, preResp *goribot.Response, request *goribot.Request) *goribot.Request {
	return request
}
func (s *MyPipeline) OnDoRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	return request
}
func (s *MyPipeline) OnResponse(spider *goribot.Spider, response *goribot.Response) *goribot.Response {
	return response
}
func (s *MyPipeline) OnItem(spider *goribot.Spider, item interface{}) interface{} {
	return item
}
func (s *MyPipeline) OnError(spider *goribot.Spider, err error) {}

func (s *MyPipeline) Finish(spider *goribot.Spider) {}
*/

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

func (s *AllowHostPipeline) OnNewRequest(spider *goribot.Spider, preResp *goribot.Response, request *goribot.Request) *goribot.Request {
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

func (s *DisallowHostPipeline) OnNewRequest(spider *goribot.Spider, preResp *goribot.Response, request *goribot.Request) *goribot.Request {
	if _, ok := s.Hosts[strings.ToLower(request.Url.Host)]; ok {
		return nil
	}
	return request
}
