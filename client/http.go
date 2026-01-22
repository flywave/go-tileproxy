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

	tlsConfig := &tls.Config{InsecureSkipVerify: config.SkipSSL}
	if config.SkipSSL {
		log.Println("WARNING: SSL verification is disabled. This is insecure and should only be used in development/testing environments. Never disable SSL verification in production!")
	}

	sc.WithTransport(&http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   120 * time.Second,
			KeepAlive: 120 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
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

// executeRequest 执行请求并处理结果的公共逻辑
func (c *CollectorClient) executeRequest(fut *Future) (statusCode int, body []byte) {
	result := fut.GetResult()
	if result == nil {
		return 500, nil
	}
	return result.StatusCode, result.Body
}

func (c *CollectorClient) Open(u string, data []byte, hdr http.Header) (statusCode int, body []byte) {
	fut := newFuture()
	var err error

	// 根据是否有数据选择不同的请求方法
	if data == nil {
		err = c.Collector.Visit(u, fut, hdr)
	} else {
		err = c.Collector.PostRaw(u, data, fut)
	}

	// 统一的错误处理
	if err != nil {
		return 500, nil
	}

	// 统一的结果处理
	return c.executeRequest(fut)
}
