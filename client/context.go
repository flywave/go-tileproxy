package client

import (
	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/crawler/queue"
)

type Context interface {
	Run() error
	Stop()
	Empty() bool
	Size() int
	Sync()
	GetHttpClient() HttpClient
}

type CollectorContext struct {
	Context
	Client *CollectorClient
	Queue  *queue.Queue
}

func NewCollectorContext(client *CollectorClient, queue *queue.Queue) *CollectorContext {
	return &CollectorContext{Client: client, Queue: queue}
}

func (c *CollectorContext) GetHttpClient() HttpClient {
	return c.Client
}

func (c *CollectorContext) GetCollector() *crawler.Collector {
	return c.Client.GetCollector()
}

func (c *CollectorContext) Run() error {
	return c.Queue.Run(c.GetCollector())
}

func (c *CollectorContext) Stop() {
	c.Queue.Stop()
}

func (c *CollectorContext) Empty() bool {
	return c.Queue.IsEmpty()
}

func (c *CollectorContext) Size() int {
	r, _ := c.Queue.Size()
	return r
}

func (c *CollectorContext) Sync() {
	c.GetCollector().Wait()
}

func (c *CollectorContext) addRequest(r *crawler.Request) error {
	return c.Queue.AddRequest(r)
}
