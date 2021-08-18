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
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
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

type MapClient interface {
	Retrieve(query *layer.MapQuery, format *tile.TileFormat) []byte
	CombinedClient(other MapClient, query *layer.MapQuery) MapClient
}

type BaseClient struct {
	ctx Context
}

func (c *BaseClient) GetHttpClient() HttpClient {
	return c.ctx.GetHttpClient()
}
