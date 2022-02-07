package client

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/crawler/extensions"
)

type HttpClient interface {
	Open(url string, data []byte, hdr http.Header) (statusCode int, body []byte)
}

type CollectorClient struct {
	HttpClient
	Collector   *crawler.Collector
	BaseRequest *crawler.Request
}

func NewCollectorClient(config *Config, ctx *crawler.Context) *CollectorClient {
	c := createCollector(config)
	cli := &CollectorClient{Collector: c, BaseRequest: &crawler.Request{Ctx: ctx}}
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

func (c *CollectorClient) GetCollector() *crawler.Collector {
	return c.Collector
}

func (c *CollectorClient) Open(u string, data []byte, hdr http.Header) (statusCode int, body []byte) {
	if data == nil {
		fut := newFuture()
		err := c.Collector.Visit(u, fut, hdr)
		if err != nil {
			return 500, nil
		}
		reqult := fut.GetResult()
		return reqult.StatusCode, reqult.Body
	} else {
		fut := newFuture()
		err := c.Collector.PostRaw(u, data, fut)
		if err != nil {
			return 500, nil
		}
		reqult := fut.GetResult()
		return reqult.StatusCode, reqult.Body
	}
}
