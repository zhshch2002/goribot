package goribot

import "sync"

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
