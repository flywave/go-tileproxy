package direct

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudflare/golibs/lrucache"

	"github.com/flywave/go-tileproxy/backend"
	"github.com/flywave/go-tileproxy/filters"
	"github.com/flywave/go-tileproxy/proxy"
	"github.com/flywave/go-tileproxy/utils"
)

const (
	filterName string = "direct"
)

type Config struct {
	Transport struct {
		Dialer struct {
			Timeout        int
			KeepAlive      int
			DualStack      bool
			DNSCacheExpiry int
			DNSCacheSize   uint
		}
		Proxy struct {
			Enabled bool
			URL     string
		}
		TLSClientConfig struct {
			InsecureSkipVerify     bool
			ClientSessionCacheSize int
		}
		DisableKeepAlives   bool
		DisableCompression  bool
		TLSHandshakeTimeout int
		MaxIdleConnsPerHost int
	}
}

func (c *Config) fromKVS(conf map[string]backend.KVS) error {
	return nil
}

type Filter struct {
	Config
	filters.RoundTripFilter
	transport *http.Transport
}

func init() {
	filters.Register(filterName, func(conf map[string]backend.KVS) (filters.Filter, error) {
		config := new(Config)
		err := config.fromKVS(conf)
		if err != nil {
			//logger.Fatal(err, "storage.ReadJsonConfig failed")
		}
		return NewFilter(config)
	})
}

func NewFilter(config *Config) (filters.Filter, error) {
	d := &utils.Dialer{
		Dialer: &net.Dialer{
			KeepAlive: time.Duration(config.Transport.Dialer.KeepAlive) * time.Second,
			Timeout:   time.Duration(config.Transport.Dialer.Timeout) * time.Second,
			DualStack: config.Transport.Dialer.DualStack,
		},
		Resolver: &utils.Resolver{
			LRUCache:  lrucache.NewLRUCache(config.Transport.Dialer.DNSCacheSize),
			DNSExpiry: time.Duration(config.Transport.Dialer.DNSCacheExpiry) * time.Second,
			BlackList: lrucache.NewLRUCache(1024),
		},
	}

	if ips, err := utils.LocalIPv4s(); err == nil {
		for _, ip := range ips {
			d.Resolver.BlackList.Set(ip.String(), struct{}{}, time.Time{})
		}
		for _, s := range []string{"127.0.0.1", "::1"} {
			d.Resolver.BlackList.Set(s, struct{}{}, time.Time{})
		}
	}

	tr := &http.Transport{
		Dial: d.Dial,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Transport.TLSClientConfig.InsecureSkipVerify,
			ClientSessionCache: tls.NewLRUClientSessionCache(config.Transport.TLSClientConfig.ClientSessionCacheSize),
		},
		TLSHandshakeTimeout: time.Duration(config.Transport.TLSHandshakeTimeout) * time.Second,
		MaxIdleConnsPerHost: config.Transport.MaxIdleConnsPerHost,
		DisableCompression:  config.Transport.DisableCompression,
	}

	if config.Transport.Proxy.Enabled {
		fixedURL, err := url.Parse(config.Transport.Proxy.URL)
		if err != nil {
			//logger.Fatal(err, "url.Parse(%#v) error", config.Transport.Proxy.URL)
		}

		switch fixedURL.Scheme {
		case "http", "https":
			tr.Proxy = http.ProxyURL(fixedURL)
			tr.Dial = nil
			tr.DialTLS = nil
		default:
			dialer, err := proxy.FromURL(fixedURL, d, nil)
			if err != nil {
				//logger.Fatal(err, "proxy.FromURL(%#v) error", fixedURL.String())
			}

			tr.Dial = dialer.Dial
			tr.DialTLS = nil
			tr.Proxy = nil
		}
	}

	return &Filter{
		Config:    *config,
		transport: tr,
	}, nil
}

func (f *Filter) FilterName() string {
	return filterName
}

func (f *Filter) RoundTrip(ctx context.Context, req *http.Request) (context.Context, *http.Response, error) {
	switch req.Method {
	case "CONNECT":
		rconn, err := f.transport.Dial("tcp", req.Host)
		if err != nil {
			return ctx, nil, err
		}

		rw := filters.GetResponseWriter(ctx)

		hijacker, ok := rw.(http.Hijacker)
		if !ok {
			return ctx, nil, fmt.Errorf("http.ResponseWriter(%#v) does not implments http.Hijacker", rw)
		}

		flusher, ok := rw.(http.Flusher)
		if !ok {
			return ctx, nil, fmt.Errorf("http.ResponseWriter(%#v) does not implments http.Flusher", rw)
		}

		rw.WriteHeader(http.StatusOK)
		flusher.Flush()

		lconn, _, err := hijacker.Hijack()
		if err != nil {
			return ctx, nil, fmt.Errorf("%#v.Hijack() error: %v", hijacker, err)
		}
		defer lconn.Close()

		go utils.IOCopy(rconn, lconn)
		utils.IOCopy(lconn, rconn)

		return ctx, filters.DummyResponse, nil
	default:
		utils.FixRequestURL(req)
		utils.FixRequestHeader(req)
		resp, err := f.transport.RoundTrip(req)

		if err != nil {
			return ctx, nil, err
		}

		if req.RemoteAddr != "" {
			//logger.Info("%s \"DIRECT %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, resp.StatusCode, resp.Header.Get("Content-Length"))
		}

		return ctx, resp, err
	}
}
