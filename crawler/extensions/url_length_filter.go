package extensions

import "github.com/flywave/go-tileproxy/crawler"

func URLLengthFilter(c *crawler.Collector, URLLengthLimit int) {
	c.OnRequest(func(r *crawler.Request) {
		if len(r.URL.String()) > URLLengthLimit {
			r.Abort()
		}
	})
}
