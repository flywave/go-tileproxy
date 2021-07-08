package client

import (
	"context"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/flywave/go-tileproxy/crawler"
)

type roundRobinSwitcher struct {
	proxyURLs []*url.URL
	index     uint32
}

func (r *roundRobinSwitcher) GetProxy(pr *http.Request) (*url.URL, error) {
	u := r.proxyURLs[r.index%uint32(len(r.proxyURLs))]
	atomic.AddUint32(&r.index, 1)
	ctx := context.WithValue(pr.Context(), crawler.ProxyURLKey, u.String())
	*pr = *pr.WithContext(ctx)
	return u, nil
}

func CustomProxy(urls []string) (crawler.ProxyFunc, error) {
	proxyURLS := make([]*url.URL, len(urls))
	for i, u := range urls {
		parsedU, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		proxyURLS[i] = parsedU
	}
	return (&roundRobinSwitcher{proxyURLS, 0}).GetProxy, nil
}
