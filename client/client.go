package client

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
	"github.com/flywave/go-tileproxy/debug"
	"github.com/flywave/go-tileproxy/extensions"
	"github.com/flywave/go-tileproxy/queue"
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
}

type BaseClient struct {
	Client
	Collector *crawler.Collector
	Queue     *queue.Queue
}

func (c *BaseClient) GetCollector() *crawler.Collector {
	return c.Collector
}

func (c *BaseClient) Sync() {
	c.Collector.Wait()
}
