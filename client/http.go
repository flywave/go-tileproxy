package client

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/crawler/debug"
	"github.com/flywave/go-tileproxy/crawler/extensions"
)

type HttpClient interface {
	Open(url string, data []byte) (statusCode int, body []byte)
}

type CollectorClient struct {
	HttpClient
	Collector *crawler.Collector
}

func NewCollectorClient(config *Config) *CollectorClient {
	return &CollectorClient{Collector: createCollector(config)}
}

func createCollector(config *Config) *crawler.Collector {
	rp, err := CustomProxy(config.Proxys)
	if err != nil {
		log.Fatal(err)
	}
	sc := crawler.NewCollector(
		crawler.Debugger(&debug.LogDebugger{}),
		crawler.Async(true),
	)

	sc.SetProxyFunc(rp)
	sc.Limit(&crawler.LimitRule{DomainGlob: "*", Parallelism: config.Threads, RandomDelay: time.Duration(config.RandomDelay) * time.Second})
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

	return sc
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
		defer c.Collector.ResetOnResponse()
		c.Collector.Wait()
		return
	} else {
		c.Collector.PostRaw(url, data)
		c.Collector.OnResponse(func(resp *crawler.Response) {
			statusCode = resp.StatusCode
			body = resp.Body
		})
		defer c.Collector.ResetOnResponse()
		c.Collector.Wait()
		return
	}
}
