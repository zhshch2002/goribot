package goribot

// Context is a wrap of response,origin request,new task,etc
type Context struct {
	// Req is the origin request
	Req *Request
	// Resp is the response object
	Resp *Response

	// tasks is the new request task which will send to the spider
	tasks []*Task
	// items is the new result data which will send to the spider，use to store
	items []interface{}
	// Meta the request task created by NewTaskWithMeta func will have a k-y pair
	Meta map[string]interface{}

	Handlers []CtxHandlerFun

	abort bool
}

// Abort this context to break the handler chain and stop handling
func (c *Context) Abort() {
	c.abort = true
}

// IsAborted return was the context dropped
func (c *Context) IsAborted() bool {
	return c.abort
}

// AddItem add an item to new item list. After every handler func return,
// spider will collect these items and call OnItem handler func
func (c *Context) AddItem(i interface{}) {
	c.items = append(c.items, i)
}

// AddTask add a task to new task list. After every handler func return,spider will collect these tasks
func (c *Context) AddTask(request *Request, handlers ...CtxHandlerFun) {
	t := NewTask(request, handlers...)
	if t != nil {
		c.tasks = append(c.tasks, t)
	}
}
