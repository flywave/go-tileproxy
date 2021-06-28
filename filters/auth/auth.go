package auth

import (
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/cloudflare/golibs/lrucache"
	"github.com/flywave/go-tileproxy/backend"
	"github.com/flywave/go-tileproxy/filters"
)

const (
	filterName string = "auth"
)

type Config struct {
	CacheSize int
	Basic     []struct {
		Username string
		Password string
	}
	WhiteList []string
}

func (c *Config) fromKVS(conf map[string]backend.KVS) error {
	return nil
}

type Filter struct {
	Config
	AuthCache lrucache.Cache
	Basic     map[string]string
	WhiteList map[string]struct{}
}

func init() {
	filters.Register(filterName, func(conf map[string]backend.KVS) (filters.Filter, error) {
		config := new(Config)
		err := config.fromKVS(conf)
		if err != nil {
			//logger.Fatal(err, "UnmarshallJson failed")
		}
		return NewFilter(config)
	})
}

func NewFilter(config *Config) (filters.Filter, error) {
	f := &Filter{
		Config:    *config,
		AuthCache: lrucache.NewMultiLRUCache(uint(runtime.NumCPU()), uint(config.CacheSize)),
		Basic:     make(map[string]string),
		WhiteList: make(map[string]struct{}),
	}

	for _, v := range config.Basic {
		f.Basic[v.Username] = v.Password
	}

	for _, v := range config.WhiteList {
		f.WhiteList[v] = struct{}{}
	}

	return f, nil
}

func (f *Filter) FilterName() string {
	return filterName
}

func (f *Filter) Request(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
	if auth := req.Header.Get("Proxy-Authorization"); auth != "" {
		req.Header.Del("Proxy-Authorization")
		f.AuthCache.Set(req.RemoteAddr, auth, time.Now().Add(time.Hour))
	}
	return ctx, req, nil
}

func (f *Filter) RoundTrip(ctx context.Context, req *http.Request) (context.Context, *http.Response, error) {

	if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if _, ok := f.WhiteList[ip]; ok {
			return ctx, nil, nil
		}
	}

	if auth, ok := f.AuthCache.GetNotStale(req.RemoteAddr); ok && auth != nil {
		parts := strings.SplitN(auth.(string), " ", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "Basic":
				if userpass, err := base64.StdEncoding.DecodeString(parts[1]); err == nil {
					parts := strings.Split(string(userpass), ":")
					user := parts[0]
					pass := parts[1]
					pass1, ok := f.Basic[user]
					if ok && pass == pass1 {
						return ctx, nil, nil
					}
				}
			default:
				break
			}
		}
	}

	noAuthResponse := &http.Response{
		StatusCode: http.StatusProxyAuthRequired,
		Header: http.Header{
			"Proxy-Authenticate": []string{"Basic realm=\"TileProxy Authentication Required\""},
		},
		Request:       req,
		Close:         true,
		ContentLength: -1,
	}

	return ctx, noAuthResponse, nil
}
