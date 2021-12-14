package client

import (
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/crawler/extensions"
)

type HttpClient interface {
	Open(url string, data []byte) (statusCode int, body []byte)
	Start() error
	Stop()
}

type CollectorClient struct {
	HttpClient
	Collector      *crawler.Collector
	CollectorQueue *Queue
	BaseRequest    *crawler.Request
}

func NewCollectorClient(config *Config, ctx *crawler.Context) *CollectorClient {
	q, _ := NewQueue(config.Threads, config.MaxQueueSize)
	c := createCollector(config)

	cli := &CollectorClient{Collector: c, CollectorQueue: q, BaseRequest: &crawler.Request{Ctx: ctx}}

	go cli.Start()

	return cli
}

func createCollector(config *Config) *crawler.Collector {
	rp, err := CustomProxy(config.Proxys)
	if err != nil {
		log.Fatal(err)
	}
	sc := crawler.NewCollector(
		crawler.Async(true),
	)

	sc.SetProxyFunc(rp)
	sc.Limit(&crawler.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Threads,
		RandomDelay: time.Duration(config.RandomDelay) * time.Second,
	})
	sc.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SkipSSL},
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   120 * time.Second,
			KeepAlive: 120 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		DisableKeepAlives:     config.DisableKeepAlives,
	})

	sc.SetRequestTimeout(config.RequestTimeout)
	if config.UserAgent != "" {
		crawler.UserAgent(config.UserAgent)(sc)
	} else {
		extensions.RandomUserAgent(sc)
	}
	extensions.Referer(sc)

	sc.OnResponse(func(resp *crawler.Response) {
		if resp.UserData != nil {
			fut := resp.UserData.(*Future)
			fut.setResult(resp)
		}
	})

	return sc
}

func (c *CollectorClient) Start() error {
	return c.CollectorQueue.Run(c.Collector)
}

func (c *CollectorClient) Stop() {
	c.CollectorQueue.Stop()
}

func (c *CollectorClient) GetCollector() *crawler.Collector {
	return c.Collector
}

func (c *CollectorClient) Open(u string, data []byte) (statusCode int, body []byte) {
	if data == nil {
		u, err := url.Parse(u)
		if err != nil {
			return 500, nil
		}
		req := &crawler.Request{
			URL:    u,
			Method: "GET",
		}
		fut, err := c.CollectorQueue.AddRequest(req)
		if err != nil {
			return 500, nil
		}
		reqult := fut.GetResult()
		return reqult.StatusCode, reqult.Body
	} else {
		u, err := url.Parse(u)
		if err != nil {
			return 500, nil
		}
		req := &crawler.Request{
			URL:    u,
			Method: "POST",
			Body:   bytes.NewBuffer(data),
		}
		fut, err := c.CollectorQueue.AddRequest(req)
		if err != nil {
			return 500, nil
		}
		reqult := fut.GetResult()
		return reqult.StatusCode, reqult.Body
	}
}
