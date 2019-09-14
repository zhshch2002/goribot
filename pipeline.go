package goribot

import "log"

type PipelineInterface interface {
	Init(spider *Spider)
	OnDoRequest(spider *Spider, request *Request) *Request
	OnNewRequest(spider *Spider, preResp *Response, request *Request) *Request
	OnResponse(spider *Spider, response *Response) *Response
	OnItem(spider *Spider, item interface{}) interface{}
	OnError(spider *Spider, err error)
	Finish(spider *Spider)
}

type Pipeline struct {
	PipelineInterface
}

func (s *Pipeline) Init(spider *Spider) {}
func (s *Pipeline) OnDoRequest(spider *Spider, request *Request) *Request {
	return request
}
func (s *Pipeline) OnNewRequest(spider *Spider, preResp *Response, request *Request) *Request {
	return request
}
func (s *Pipeline) OnResponse(spider *Spider, response *Response) *Response {
	return response
}
func (s *Pipeline) OnItem(spider *Spider, item interface{}) interface{} {
	return item
}
func (s *Pipeline) OnError(spider *Spider, err error) {}

func (s *Pipeline) Finish(spider *Spider) {}

type PrintLogPipeline struct {
	Pipeline
}

func (s *PrintLogPipeline) Init(spider *Spider) {
	log.Println("Pipe Init")
	return
}
func (s *PrintLogPipeline) OnDoRequest(spider *Spider, request *Request) *Request {
	log.Println("Pipe OnRequest")
	return request
}
func (s *PrintLogPipeline) OnNewRequest(spider *Spider, preResp *Response, request *Request) *Request {
	log.Println("Pipe OnRequest")
	return request
}
func (s *PrintLogPipeline) OnResponse(spider *Spider, response *Response) *Response {
	log.Println("Pipe OnResponse")
	return response
}
func (s *PrintLogPipeline) OnItem(spider *Spider, item interface{}) interface{} {
	log.Println("Pipe OnItem", item)
	return item
}
func (s *PrintLogPipeline) OnError(spider *Spider, err error) {
	log.Println("Pipe OnError", err)
}
