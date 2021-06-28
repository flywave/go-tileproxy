package filters

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/flywave/go-tileproxy/backend"
)

const (
	FILTER_CONFIG_PREFIX = "TILEPROXY_FILTER_"
)

var (
	DummyRequest  *http.Request  = &http.Request{}
	DummyResponse *http.Response = &http.Response{}
)

type Filter interface {
	FilterName() string
}

type RequestFilter interface {
	Filter
	Request(context.Context, *http.Request) (context.Context, *http.Request, error)
}

type RoundTripFilter interface {
	Filter
	RoundTrip(context.Context, *http.Request) (context.Context, *http.Response, error)
}

type ResponseFilter interface {
	Filter
	Response(context.Context, *http.Response) (context.Context, *http.Response, error)
}

var (
	mu  = new(sync.Mutex)
	mm  = make(map[string]*sync.Mutex)
	fnm = make(map[string]func(map[string]backend.KVS) (Filter, error))
	fm  = make(map[string]Filter)
)

func Register(name string, New func(map[string]backend.KVS) (Filter, error)) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := fnm[name]; !ok {
		fnm[name] = New
		fm[name] = nil
		mm[name] = new(sync.Mutex)
	}
}

func GetFilter(name string, conf backend.Config) (Filter, error) {
	if f, ok := fm[name]; ok && f != nil {
		return f, nil
	}

	mu := mm[name]
	mu.Lock()
	defer mu.Unlock()

	if f, ok := fm[name]; ok && f != nil {
		return f, nil
	}

	confname := FILTER_CONFIG_PREFIX + strings.ToUpper(name)
	f, err := fnm[name](conf[confname])
	if err != nil {
		return nil, err
	}

	fm[name] = f

	return f, nil
}
