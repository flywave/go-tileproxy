package crawler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

// 字符串池用于复用常用字符串buffer
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

type Request struct {
	URL                       *url.URL
	Headers                   *http.Header
	Ctx                       *Context
	Depth                     int
	Method                    string
	Body                      io.Reader
	ResponseCharacterEncoding string
	ID                        uint32
	collector                 *Collector
	abort                     bool
	baseURL                   *url.URL
	ProxyURL                  string
	UserData                  interface{}
}

type serializableRequest struct {
	URL     string
	Method  string
	Depth   int
	Body    []byte
	ID      uint32
	Ctx     map[string]interface{}
	Headers http.Header
}

func (r *Request) New(method, URL string, body io.Reader) (*Request, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	return &Request{
		Method:    method,
		URL:       u,
		Body:      body,
		Ctx:       r.Ctx,
		UserData:  r.UserData,
		Headers:   &http.Header{},
		ID:        atomic.AddUint32(&r.collector.requestCount, 1),
		collector: r.collector,
	}, nil
}

func (r *Request) Abort() {
	r.abort = true
}

func (r *Request) AbsoluteURL(u string) string {
	if strings.HasPrefix(u, "#") {
		return ""
	}

	// 优化：如果已经是绝对URL，直接返回
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}

	var base *url.URL
	if r.baseURL != nil {
		base = r.baseURL
	} else {
		base = r.URL
	}

	if base == nil {
		return u
	}

	absURL, err := base.Parse(u)
	if err != nil {
		return ""
	}

	absURL.Fragment = ""
	if absURL.Scheme == "//" {
		absURL.Scheme = r.URL.Scheme
	}
	return absURL.String()
}

func (r *Request) Visit(URL string) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "GET", r.Depth+1, nil, r.Ctx, r.UserData, nil)
}

func (r *Request) Post(URL string, requestData map[string]string) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, createFormReader(requestData), r.Ctx, r.UserData, nil)
}

func (r *Request) PostRaw(URL string, requestData []byte) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, bytes.NewReader(requestData), r.Ctx, r.UserData, nil)
}

func (r *Request) PostMultipart(URL string, requestData map[string][]byte) error {
	boundary := randomBoundary()
	hdr := http.Header{}
	hdr.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	hdr.Set("User-Agent", r.collector.UserAgent)
	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, createMultipartReader(boundary, requestData), r.Ctx, r.UserData, hdr)
}

func (r *Request) Retry() error {
	r.Headers.Del("Cookie")
	return r.collector.scrape(r.URL.String(), r.Method, r.Depth, r.Body, r.Ctx, r.UserData, *r.Headers)
}

func (r *Request) Do() error {
	return r.collector.scrape(r.URL.String(), r.Method, r.Depth, r.Body, r.Ctx, r.UserData, *r.Headers)
}

func (r *Request) Wait() {
	r.collector.Wait()
}

func (r *Request) Marshal() ([]byte, error) {
	ctx := make(map[string]interface{})
	if r.Ctx != nil {
		r.Ctx.ForEach(func(k string, v interface{}) interface{} {
			ctx[k] = v
			return nil
		})
	}
	var err error
	var body []byte
	if r.Body != nil {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
	}
	sr := &serializableRequest{
		URL:    r.URL.String(),
		Method: r.Method,
		Depth:  r.Depth,
		Body:   body,
		ID:     r.ID,
		Ctx:    ctx,
	}
	if r.Headers != nil {
		sr.Headers = *r.Headers
	}
	return json.Marshal(sr)
}
