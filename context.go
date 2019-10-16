package goribot

import "github.com/PuerkitoBio/goquery"

// Context is a wrap of response,origin request,new task,etc
type Context struct {
	Text string                 // the response text
	Html *goquery.Document      // spider will try to parse the response as html
	Json map[string]interface{} // spider will try to parse the response as json

	Request  *Request  // origin request
	Response *Response // a response object

	Tasks []*Task                // the new request task which will send to the spider
	Items []interface{}          // the new result data which will send to the spider，use to store
	Meta  map[string]interface{} // the request task created by NewTaskWithMeta func will have a k-y pair

	drop bool // in handlers chain,you can use ctx.Drop() to break the handler chain and stop handling
}

// Drop this context to break the handler chain and stop handling
func (c *Context) Drop() {
	c.drop = true
}

// IsDrop return was the context dropped
func (c *Context) IsDrop() bool {
	return c.drop
}

// AddItem add an item to new item list. After every handler func return,
// spider will collect these items and call OnItem handler func
func (c *Context) AddItem(i interface{}) {
	c.Items = append(c.Items, i)
}

// AddTask add a task to new task list. After every handler func return,spider will collect these tasks
func (c *Context) AddTask(r *Task) {
	c.Tasks = append(c.Tasks, r)
}

// NewTask create a task and add it to new task list After every handler func return,spider will collect these tasks
func (c *Context) NewTask(req *Request, RespHandler ...func(ctx *Context)) {
	c.AddTask(NewTask(req, RespHandler...))
}

// NewTaskWithMeta create a task with meta data and add it to new task list After every handler func return,
// spider will collect these tasks
func (c *Context) NewTaskWithMeta(req *Request, meta map[string]interface{}, RespHandler ...func(ctx *Context)) {
	t := NewTask(req, RespHandler...)
	t.Meta = meta
	c.Tasks = append(c.Tasks, t)

}
