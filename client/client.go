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

type Config struct {
	URL               string
	SkipSSL           bool
	Threads           int
	UserAgent         string
	RandomDelay       int
	DisableKeepAlives bool
	Proxys            []string
	RequestTimeout    time.Duration
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

type Client interface {
	GetCollector() *crawler.Collector
	Sync()
	Open(url string, data []byte) *crawler.Response
	Get(url string) *crawler.Response
}

type BaseClient struct {
	Client
	Collector *crawler.Collector
}

func (c *BaseClient) GetCollector() *crawler.Collector {
	return c.Collector
}

func (c *BaseClient) Sync() {
	c.Collector.Wait()
}

func (c *BaseClient) Get(url string) *crawler.Response {
	return nil
}
