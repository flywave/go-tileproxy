package proxy

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
	index := atomic.AddUint32(&r.index, 1) - 1
	u := r.proxyURLs[index%uint32(len(r.proxyURLs))]
	ctx := context.WithValue(pr.Context(), crawler.ProxyURLKey, u.String())
	*pr = *pr.WithContext(ctx)
	return u, nil
}

func RoundRobinProxySwitcher(ProxyURLs ...string) (crawler.ProxyFunc, error) {
	if len(ProxyURLs) < 1 {
		return nil, crawler.ErrEmptyProxyURL
	}
	urls := make([]*url.URL, len(ProxyURLs))
	for i, u := range ProxyURLs {
		parsedU, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		urls[i] = parsedU
	}
	return (&roundRobinSwitcher{urls, 0}).GetProxy, nil
}
