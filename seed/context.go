package seed

import (
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/queue"
)

type Context struct {
	Client client.MapClient
	Queue  *queue.Queue
}

func (c *Context) GetCollector() *crawler.Collector {
	return c.GetCollector()
}

func (c *Context) Run() error {
	return c.Queue.Run(c.GetCollector())
}

func (c *Context) Stop() {
	c.Queue.Stop()
}

func (c *Context) IsEmpty() bool {
	return c.Queue.IsEmpty()
}

func (c *Context) Size() int {
	r, _ := c.Queue.Size()
	return r
}

func (c *Context) Sync() {
	c.GetCollector().Wait()
}

func (c *Context) addRequest(r *crawler.Request) error {
	return c.Queue.AddRequest(r)
}
