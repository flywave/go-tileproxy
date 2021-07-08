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

func createCollector(proxys []string) *crawler.Collector {
	rp, err := CustomProxy(proxys)
	if err != nil {
		log.Fatal(err)
	}

	sc := crawler.NewCollector(
		crawler.Debugger(&debug.LogDebugger{}),
		crawler.AllowedDomains(allowed1, allowed2),
		crawler.MaxDepth(2),
		crawler.Async(true),
	)
	sc.SetProxyFunc(rp)
	sc.Limit(&crawler.LimitRule{DomainGlob: "*", Parallelism: 10, RandomDelay: 20 * time.Second})
	sc.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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
		DisableKeepAlives:     true,
	})

	min5, _ := time.ParseDuration("10m")
	sc.SetRequestTimeout(min5)

	extensions.RandomUserAgent(sc)
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
