package goribot

import (
	"sync"
)

// Scheduler is a queue of tasks and items
type Scheduler interface {
	// GetTask pops a task
	GetTask() *Task
	// GetItem pops a item
	GetItem() interface{}

	// AddTask push a task
	AddTask(t *Task)
	// AddItem push a item
	AddItem(i interface{})

	// IsTaskEmpty returns is tasks queue empty
	IsTaskEmpty() bool
	// IsItemEmpty returns is items queue empty
	IsItemEmpty() bool
}

// Scheduler is default scheduler of goribot
type BaseScheduler struct {
	tasksLock sync.Mutex
	tasks     []*Task
	itemsLock sync.Mutex
	items     []interface{}
	// DepthFirst sets push new tasks to the top of the queue
	DepthFirst bool
}

func NewBaseScheduler(depthFirst bool) *BaseScheduler {
	return &BaseScheduler{DepthFirst: depthFirst, tasksLock: sync.Mutex{}, itemsLock: sync.Mutex{}}
}

func (s *BaseScheduler) GetTask() *Task {
	if len(s.tasks) == 0 {
		return nil
	}
	s.tasksLock.Lock()
	task := s.tasks[0]
	s.tasks = s.tasks[1:]
	s.tasksLock.Unlock()
	return task

}
func (s *BaseScheduler) GetItem() interface{} {
	if len(s.items) == 0 {
		return nil
	}
	s.itemsLock.Lock()
	item := s.items[0]
	s.items = s.items[1:]
	s.itemsLock.Unlock()
	return item
}
func (s *BaseScheduler) AddTask(t *Task) {
	s.tasksLock.Lock()
	if s.DepthFirst {
		s.tasks = append([]*Task{t}, s.tasks...)
	} else {
		s.tasks = append(s.tasks, t)
	}
	s.tasksLock.Unlock()
}
func (s *BaseScheduler) AddItem(i interface{}) {
	s.itemsLock.Lock()
	s.items = append(s.items, i)
	s.itemsLock.Unlock()

}
func (s *BaseScheduler) IsTaskEmpty() bool {
	return len(s.tasks) == 0
}
func (s *BaseScheduler) IsItemEmpty() bool {
	return len(s.items) == 0
}
