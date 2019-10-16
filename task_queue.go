package goribot

import "sync"

// Task is a wrap of request and its handler funcs
type Task struct {
	Request        *Request
	onRespHandlers []func(ctx *Context)
	Meta           map[string]interface{}
}

// NewTask create a new task
func NewTask(req *Request, RespHandler ...func(ctx *Context)) *Task {
	return &Task{Request: req, onRespHandlers: RespHandler}
}

// TaskQueue is a queue of task
type TaskQueue struct {
	sync.Mutex
	items []*Task
}

// NewTaskQueue create a new queue
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		Mutex: sync.Mutex{},
	}
}

// Push a task to the queue
func (s *TaskQueue) Push(item *Task) {
	s.Lock()
	s.items = append(s.items, item)
	s.Unlock()
}

// PushInHead push a task to the queue head
func (s *TaskQueue) PushInHead(item *Task) {
	s.Lock()
	s.items = append([]*Task{item}, s.items...)
	s.Unlock()
}

// Pop a task from the queue
func (s *TaskQueue) Pop() *Task {
	s.Lock()
	item := s.items[0]
	s.items = s.items[1:]
	s.Unlock()
	return item
}

// IsEmpty return true if the queue is empty
func (s *TaskQueue) IsEmpty() bool {
	return len(s.items) == 0
}
