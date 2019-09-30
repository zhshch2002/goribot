package goribot

import "sync"

type TaskQueue struct {
	sync.Mutex
	items []*Task
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		Mutex: sync.Mutex{},
	}
}

func (s *TaskQueue) Push(item *Task) {
	s.Lock()
	s.items = append(s.items, item)
	s.Unlock()
}
func (s *TaskQueue) PushInHead(item *Task) {
	s.Lock()
	s.items = append([]*Task{item}, s.items...)
	s.Unlock()
}
func (s *TaskQueue) Pop() *Task {
	s.Lock()
	item := s.items[0]
	s.items = s.items[1:]
	s.Unlock()
	return item
}

func (s *TaskQueue) IsEmpty() bool {
	return len(s.items) == 0
}
