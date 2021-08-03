package client

import "github.com/flywave/go-tileproxy/crawler"

type HttpClient interface {
	Open(url string, data []byte) (statusCode int, body []byte)
}

type CollectorClient struct {
	HttpClient
	Collector *crawler.Collector
}

func (c *CollectorClient) GetCollector() *crawler.Collector {
	return c.Collector
}

func (c *CollectorClient) Open(url string, data []byte) (statusCode int, body []byte) {
	if data == nil {
		c.Collector.Visit(url)
		c.Collector.OnResponse(func(resp *crawler.Response) {
			statusCode = resp.StatusCode
			body = resp.Body
		})
		c.Collector.Wait()
		return
	} else {
		c.Collector.PostRaw(url, data)
		c.Collector.OnResponse(func(resp *crawler.Response) {
			statusCode = resp.StatusCode
			body = resp.Body
		})
		c.Collector.Wait()
		return
	}
}
