package extensions

import "github.com/flywave/go-tileproxy/crawler"

func Referer(c *crawler.Collector) {
	c.OnResponse(func(r *crawler.Response) {
		r.Ctx.Put("_referer", r.Request.URL.String())
	})
	c.OnRequest(func(r *crawler.Request) {
		if ref := r.Ctx.Get("_referer"); ref != "" {
			r.Headers.Set("Referer", ref)
		}
	})
}
