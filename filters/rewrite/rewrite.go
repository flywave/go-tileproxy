package rewrite

import (
	"context"
	"net/http"

	"github.com/flywave/go-tileproxy/backend"
	"github.com/flywave/go-tileproxy/filters"
)

const (
	filterName string = "rewrite"
)

type Config struct {
	UserAgent struct {
		Enabled bool
		Value   string
	}
	Host struct {
		Enabled   bool
		RewriteBy string
	}
}

func (c *Config) fromKVS(conf map[string]backend.KVS) error {
	return nil
}

type Filter struct {
	Config
	UserAgentEnabled bool
	UserAgentValue   string
	HostEnabled      bool
	HostRewriteBy    string
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
	f := &Filter{
		Config:           *config,
		UserAgentEnabled: config.UserAgent.Enabled,
		UserAgentValue:   config.UserAgent.Value,
		HostEnabled:      config.Host.Enabled,
		HostRewriteBy:    config.Host.RewriteBy,
	}

	return f, nil
}

func (f *Filter) FilterName() string {
	return filterName
}

func (f *Filter) Request(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
	if f.UserAgentEnabled {
		req.Header.Set("User-Agent", f.UserAgentValue)
	}

	if f.HostEnabled {
		if host := req.Header.Get(f.HostRewriteBy); host != "" {
			req.Host = host
			req.Header.Set("Host", req.Host)
			req.Header.Del(f.HostRewriteBy)
		}
	}

	return ctx, req, nil
}

func (f *Filter) Response(ctx context.Context, resp *http.Response) (context.Context, *http.Response, error) {
	return ctx, resp, nil
}
