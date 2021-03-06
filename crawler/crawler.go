package crawler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kennygrant/sanitize"
)

type CollectorOption func(*Collector)

type Collector struct {
	UserAgent                string
	AllowedDomains           []string
	DisallowedDomains        []string
	DisallowedURLFilters     []*regexp.Regexp
	URLFilters               []*regexp.Regexp
	MaxBodySize              int
	CacheDir                 string
	Async                    bool
	ParseHTTPErrorResponse   bool
	ID                       uint32
	DetectCharset            bool
	redirectHandler          func(req *http.Request, via []*http.Request) error
	CheckHead                bool
	TraceHTTP                bool
	Context                  context.Context
	requestCallbacks         []RequestCallback
	responseCallbacks        []ResponseCallback
	responseHeadersCallbacks []ResponseHeadersCallback
	errorCallbacks           []ErrorCallback
	requestCount             uint32
	responseCount            uint32
	backend                  *httpBackend
	wg                       *sync.WaitGroup
	lock                     *sync.RWMutex
}

type RequestCallback func(*Request)

type ResponseHeadersCallback func(*Response)

type ResponseCallback func(*Response)

type ErrorCallback func(*Response, error)

type ProxyFunc func(*http.Request) (*url.URL, error)

var collectorCounter uint32

type key int

const ProxyURLKey key = iota

var (
	ErrForbiddenDomain     = errors.New("forbidden domain")
	ErrMissingURL          = errors.New("missing URL")
	ErrForbiddenURL        = errors.New("forbiddenURL")
	ErrNoURLFiltersMatch   = errors.New("no URLFilters match")
	ErrAlreadyVisited      = errors.New("URL already visited")
	ErrRobotsTxtBlocked    = errors.New("URL blocked by robots.txt")
	ErrNoCookieJar         = errors.New("cookie jar is not available")
	ErrNoPattern           = errors.New("no pattern defined in LimitRule")
	ErrEmptyProxyURL       = errors.New("proxy URL list is empty")
	ErrAbortedAfterHeaders = errors.New("aborted after receiving response headers")
	ErrQueueFull           = errors.New("queue MaxSize reached")
)

var envMap = map[string]func(*Collector, string){
	"ALLOWED_DOMAINS": func(c *Collector, val string) {
		c.AllowedDomains = strings.Split(val, ",")
	},
	"CACHE_DIR": func(c *Collector, val string) {
		c.CacheDir = val
	},
	"DETECT_CHARSET": func(c *Collector, val string) {
		c.DetectCharset = isYesString(val)
	},
	"DISABLE_COOKIES": func(c *Collector, _ string) {
		c.backend.Client.Jar = nil
	},
	"DISALLOWED_DOMAINS": func(c *Collector, val string) {
		c.DisallowedDomains = strings.Split(val, ",")
	},
	"FOLLOW_REDIRECTS": func(c *Collector, val string) {
		if !isYesString(val) {
			c.redirectHandler = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
	},
	"MAX_BODY_SIZE": func(c *Collector, val string) {
		size, err := strconv.Atoi(val)
		if err == nil {
			c.MaxBodySize = size
		}
	},
	"PARSE_HTTP_ERROR_RESPONSE": func(c *Collector, val string) {
		c.ParseHTTPErrorResponse = isYesString(val)
	},
	"TRACE_HTTP": func(c *Collector, val string) {
		c.TraceHTTP = isYesString(val)
	},
	"USER_AGENT": func(c *Collector, val string) {
		c.UserAgent = val
	},
}

func NewCollector(options ...CollectorOption) *Collector {
	c := &Collector{}
	c.Init()

	for _, f := range options {
		f(c)
	}

	c.parseSettingsFromEnv()

	return c
}

func UserAgent(ua string) CollectorOption {
	return func(c *Collector) {
		c.UserAgent = ua
	}
}

func AllowedDomains(domains ...string) CollectorOption {
	return func(c *Collector) {
		c.AllowedDomains = domains
	}
}

func ParseHTTPErrorResponse() CollectorOption {
	return func(c *Collector) {
		c.ParseHTTPErrorResponse = true
	}
}

func DisallowedDomains(domains ...string) CollectorOption {
	return func(c *Collector) {
		c.DisallowedDomains = domains
	}
}

func DisallowedURLFilters(filters ...*regexp.Regexp) CollectorOption {
	return func(c *Collector) {
		c.DisallowedURLFilters = filters
	}
}

func URLFilters(filters ...*regexp.Regexp) CollectorOption {
	return func(c *Collector) {
		c.URLFilters = filters
	}
}

func MaxBodySize(sizeInBytes int) CollectorOption {
	return func(c *Collector) {
		c.MaxBodySize = sizeInBytes
	}
}

func CacheDir(path string) CollectorOption {
	return func(c *Collector) {
		c.CacheDir = path
	}
}

func TraceHTTP() CollectorOption {
	return func(c *Collector) {
		c.TraceHTTP = true
	}
}

func StdlibContext(ctx context.Context) CollectorOption {
	return func(c *Collector) {
		c.Context = ctx
	}
}

func ID(id uint32) CollectorOption {
	return func(c *Collector) {
		c.ID = id
	}
}

func Async(a ...bool) CollectorOption {
	return func(c *Collector) {
		c.Async = true
	}
}

func DetectCharset() CollectorOption {
	return func(c *Collector) {
		c.DetectCharset = true
	}
}

func CheckHead() CollectorOption {
	return func(c *Collector) {
		c.CheckHead = true
	}
}

func (c *Collector) Init() {
	c.UserAgent = "crawler"
	c.MaxBodySize = 10 * 1024 * 1024
	c.backend = &httpBackend{}
	jar, _ := cookiejar.New(nil)
	c.backend.Init(jar)
	c.backend.Client.CheckRedirect = c.checkRedirectFunc()
	c.wg = &sync.WaitGroup{}
	c.lock = &sync.RWMutex{}
	c.ID = atomic.AddUint32(&collectorCounter, 1)
	c.TraceHTTP = false
	c.Context = context.Background()
}

func (c *Collector) Visit(URL string, userData interface{}, hdr http.Header) error {
	if c.CheckHead {
		if check := c.scrape(URL, "HEAD", 1, nil, nil, userData, nil); check != nil {
			return check
		}
	}
	return c.scrape(URL, "GET", 1, nil, nil, userData, hdr)
}

func (c *Collector) Head(URL string) error {
	return c.scrape(URL, "HEAD", 1, nil, nil, nil, nil)
}

func (c *Collector) Post(URL string, requestData map[string]string, userData interface{}) error {
	return c.scrape(URL, "POST", 1, createFormReader(requestData), nil, userData, nil)
}

func (c *Collector) PostRaw(URL string, requestData []byte, userData interface{}) error {
	return c.scrape(URL, "POST", 1, bytes.NewReader(requestData), nil, userData, nil)
}

func (c *Collector) PostMultipart(URL string, requestData map[string][]byte, userData interface{}) error {
	boundary := randomBoundary()
	hdr := http.Header{}
	hdr.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	hdr.Set("User-Agent", c.UserAgent)
	return c.scrape(URL, "POST", 1, createMultipartReader(boundary, requestData), nil, userData, hdr)
}

func (c *Collector) Request(method, URL string, requestData io.Reader, ctx *Context, userData interface{}, hdr http.Header) error {
	return c.scrape(URL, method, 1, requestData, ctx, userData, hdr)
}

func (c *Collector) UnmarshalRequest(r []byte) (*Request, error) {
	req := &serializableRequest{}
	err := json.Unmarshal(r, req)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, err
	}

	ctx := NewContext()
	for k, v := range req.Ctx {
		ctx.Put(k, v)
	}

	return &Request{
		Method:    req.Method,
		URL:       u,
		Depth:     req.Depth,
		Body:      bytes.NewReader(req.Body),
		Ctx:       ctx,
		ID:        atomic.AddUint32(&c.requestCount, 1),
		Headers:   &req.Headers,
		collector: c,
	}, nil
}

func (c *Collector) scrape(u, method string, depth int, requestData io.Reader, ctx *Context, userData interface{}, hdr http.Header) error {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return err
	}
	if err := c.requestCheck(u, parsedURL, method, requestData, depth); err != nil {
		return err
	}

	if hdr == nil {
		hdr = http.Header{}
	}
	if _, ok := hdr["User-Agent"]; !ok {
		hdr.Set("User-Agent", c.UserAgent)
	}
	rc, ok := requestData.(io.ReadCloser)
	if !ok && requestData != nil {
		rc = ioutil.NopCloser(requestData)
	}

	host := parsedURL.Host
	if hostHeader := hdr.Get("Host"); hostHeader != "" {
		host = hostHeader
	}
	req := &http.Request{
		Method:     method,
		URL:        parsedURL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     hdr,
		Body:       rc,
		Host:       host,
	}

	req = req.WithContext(c.Context)
	setRequestBody(req, requestData)
	u = parsedURL.String()
	c.wg.Add(1)
	if c.Async {
		go c.fetch(u, method, depth, requestData, ctx, userData, hdr, req)
		return nil
	}
	return c.fetch(u, method, depth, requestData, ctx, userData, hdr, req)
}

func setRequestBody(req *http.Request, body io.Reader) {
	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.ContentLength = int64(v.Len())
			buf := v.Bytes()
			req.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return ioutil.NopCloser(r), nil
			}
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		}
		if req.GetBody != nil && req.ContentLength == 0 {
			req.Body = http.NoBody
			req.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}
}

func (c *Collector) fetch(u, method string, depth int, requestData io.Reader, ctx *Context, userData interface{}, hdr http.Header, req *http.Request) error {
	defer c.wg.Done()
	if ctx == nil {
		ctx = NewContext()
	}
	request := &Request{
		URL:       req.URL,
		Headers:   &req.Header,
		Ctx:       ctx,
		Depth:     depth,
		Method:    method,
		Body:      requestData,
		collector: c,
		ID:        atomic.AddUint32(&c.requestCount, 1),
	}

	c.handleOnRequest(request)

	if request.abort {
		return nil
	}

	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}

	var hTrace *HTTPTrace
	if c.TraceHTTP {
		hTrace = &HTTPTrace{}
		req = hTrace.WithTrace(req)
	}
	origURL := req.URL
	checkHeadersFunc := func(req *http.Request, statusCode int, headers http.Header) bool {
		if req.URL != origURL {
			request.URL = req.URL
			request.Headers = &req.Header
		}
		c.handleOnResponseHeaders(&Response{Ctx: ctx, Request: request, StatusCode: statusCode, Headers: &headers})
		return !request.abort
	}
	response, err := c.backend.Cache(req, c.MaxBodySize, checkHeadersFunc, c.CacheDir)
	if proxyURL, ok := req.Context().Value(ProxyURLKey).(string); ok {
		request.ProxyURL = proxyURL
	}
	if err := c.handleOnError(response, err, request, ctx); err != nil {
		return err
	}
	atomic.AddUint32(&c.responseCount, 1)
	response.Ctx = ctx
	response.Request = request
	response.Trace = hTrace
	response.UserData = userData

	c.handleOnResponse(response)

	if err != nil {
		c.handleOnError(response, err, request, ctx)
	}

	return err
}

func (c *Collector) requestCheck(u string, parsedURL *url.URL, method string, requestData io.Reader, depth int) error {
	if u == "" {
		return ErrMissingURL
	}
	if len(c.DisallowedURLFilters) > 0 {
		if isMatchingFilter(c.DisallowedURLFilters, []byte(u)) {
			return ErrForbiddenURL
		}
	}
	if len(c.URLFilters) > 0 {
		if !isMatchingFilter(c.URLFilters, []byte(u)) {
			return ErrNoURLFiltersMatch
		}
	}
	if !c.isDomainAllowed(parsedURL.Hostname()) {
		return ErrForbiddenDomain
	}
	return nil
}

func (c *Collector) isDomainAllowed(domain string) bool {
	for _, d2 := range c.DisallowedDomains {
		if d2 == domain {
			return false
		}
	}
	if c.AllowedDomains == nil || len(c.AllowedDomains) == 0 {
		return true
	}
	for _, d2 := range c.AllowedDomains {
		if d2 == domain {
			return true
		}
	}
	return false
}

func (c *Collector) String() string {
	return fmt.Sprintf(
		"Requests made: %d (%d responses) | Callbacks: OnRequest: %d, OnResponse: %d, OnError: %d",
		atomic.LoadUint32(&c.requestCount),
		atomic.LoadUint32(&c.responseCount),
		len(c.requestCallbacks),
		len(c.responseCallbacks),
		len(c.errorCallbacks),
	)
}

func (c *Collector) Wait() {
	c.wg.Wait()
}

func (c *Collector) OnRequest(f RequestCallback) {
	c.lock.Lock()
	if c.requestCallbacks == nil {
		c.requestCallbacks = make([]RequestCallback, 0, 4)
	}
	c.requestCallbacks = append(c.requestCallbacks, f)
	c.lock.Unlock()
}

func (c *Collector) OnResponseHeaders(f ResponseHeadersCallback) {
	c.lock.Lock()
	c.responseHeadersCallbacks = append(c.responseHeadersCallbacks, f)
	c.lock.Unlock()
}

func (c *Collector) ResetOnResponse() {
	c.lock.Lock()
	c.responseCallbacks = make([]ResponseCallback, 0, 4)
	c.lock.Unlock()
}

func (c *Collector) OnResponse(f ResponseCallback) {
	c.lock.Lock()
	if c.responseCallbacks == nil {
		c.responseCallbacks = make([]ResponseCallback, 0, 4)
	}
	c.responseCallbacks = append(c.responseCallbacks, f)
	c.lock.Unlock()
}

func (c *Collector) OnError(f ErrorCallback) {
	c.lock.Lock()
	if c.errorCallbacks == nil {
		c.errorCallbacks = make([]ErrorCallback, 0, 4)
	}
	c.errorCallbacks = append(c.errorCallbacks, f)
	c.lock.Unlock()
}

func (c *Collector) SetClient(client *http.Client) {
	c.backend.Client = client
}

func (c *Collector) WithTransport(transport http.RoundTripper) {
	c.backend.Client.Transport = transport
}

func (c *Collector) DisableCookies() {
	c.backend.Client.Jar = nil
}

func (c *Collector) SetCookieJar(j http.CookieJar) {
	c.backend.Client.Jar = j
}

func (c *Collector) SetRequestTimeout(timeout time.Duration) {
	c.backend.Client.Timeout = timeout
}

func (c *Collector) SetProxy(proxyURL string) error {
	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}

	c.SetProxyFunc(http.ProxyURL(proxyParsed))

	return nil
}

func (c *Collector) SetProxyFunc(p ProxyFunc) {
	t, ok := c.backend.Client.Transport.(*http.Transport)
	if c.backend.Client.Transport != nil && ok {
		t.Proxy = p
		t.DisableKeepAlives = true
	} else {
		c.backend.Client.Transport = &http.Transport{
			Proxy:             p,
			DisableKeepAlives: true,
		}
	}
}

func (c *Collector) handleOnRequest(r *Request) {
	for _, f := range c.requestCallbacks {
		f(r)
	}
}

func (c *Collector) handleOnResponse(r *Response) {
	for _, f := range c.responseCallbacks {
		f(r)
	}
}

func (c *Collector) handleOnResponseHeaders(r *Response) {
	for _, f := range c.responseHeadersCallbacks {
		f(r)
	}
}

func (c *Collector) handleOnError(response *Response, err error, request *Request, ctx *Context) error {
	if err == nil && (c.ParseHTTPErrorResponse || response.StatusCode < 203) {
		return nil
	}
	if err == nil && response.StatusCode >= 203 {
		err = errors.New(http.StatusText(response.StatusCode))
	}
	if response == nil {
		response = &Response{
			Request: request,
			Ctx:     ctx,
		}
	}
	if response.Request == nil {
		response.Request = request
	}
	if response.Ctx == nil {
		response.Ctx = request.Ctx
	}
	for _, f := range c.errorCallbacks {
		f(response, err)
	}
	return err
}

func (c *Collector) Limit(rule *LimitRule) error {
	return c.backend.Limit(rule)
}

func (c *Collector) Limits(rules []*LimitRule) error {
	return c.backend.Limits(rules)
}

func (c *Collector) SetRedirectHandler(f func(req *http.Request, via []*http.Request) error) {
	c.redirectHandler = f
	c.backend.Client.CheckRedirect = c.checkRedirectFunc()
}

func (c *Collector) SetCookies(URL string, cookies []*http.Cookie) error {
	if c.backend.Client.Jar == nil {
		return ErrNoCookieJar
	}
	u, err := url.Parse(URL)
	if err != nil {
		return err
	}
	c.backend.Client.Jar.SetCookies(u, cookies)
	return nil
}

func (c *Collector) Cookies(URL string) []*http.Cookie {
	if c.backend.Client.Jar == nil {
		return nil
	}
	u, err := url.Parse(URL)
	if err != nil {
		return nil
	}
	return c.backend.Client.Jar.Cookies(u)
}

func (c *Collector) Clone() *Collector {
	return &Collector{
		AllowedDomains:         c.AllowedDomains,
		CacheDir:               c.CacheDir,
		DetectCharset:          c.DetectCharset,
		DisallowedDomains:      c.DisallowedDomains,
		ID:                     atomic.AddUint32(&collectorCounter, 1),
		MaxBodySize:            c.MaxBodySize,
		DisallowedURLFilters:   c.DisallowedURLFilters,
		URLFilters:             c.URLFilters,
		CheckHead:              c.CheckHead,
		ParseHTTPErrorResponse: c.ParseHTTPErrorResponse,
		UserAgent:              c.UserAgent,
		TraceHTTP:              c.TraceHTTP,
		Context:                c.Context,
		backend:                c.backend,
		Async:                  c.Async,
		redirectHandler:        c.redirectHandler,
		errorCallbacks:         make([]ErrorCallback, 0, 8),
		lock:                   c.lock,
		requestCallbacks:       make([]RequestCallback, 0, 8),
		responseCallbacks:      make([]ResponseCallback, 0, 8),
		wg:                     &sync.WaitGroup{},
	}
}

func (c *Collector) checkRedirectFunc() func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if !c.isDomainAllowed(req.URL.Hostname()) {
			return fmt.Errorf("not following redirect to %s because its not in AllowedDomains", req.URL.Host)
		}

		if c.redirectHandler != nil {
			return c.redirectHandler(req, via)
		}

		if len(via) >= 10 {
			return http.ErrUseLastResponse
		}

		lastRequest := via[len(via)-1]

		if req.URL.Host != lastRequest.URL.Host {
			req.Header.Del("Authorization")
		}

		return nil
	}
}

func (c *Collector) parseSettingsFromEnv() {
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "TILEPROXY_") {
			continue
		}
		pair := strings.SplitN(e[6:], "=", 2)
		if f, ok := envMap[pair[0]]; ok {
			f(c, pair[1])
		} else {
			log.Println("Unknown environment variable:", pair[0])
		}
	}
}

func SanitizeFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	cleanExt := sanitize.BaseName(ext)
	if cleanExt == "" {
		cleanExt = ".unknown"
	}
	return strings.Replace(fmt.Sprintf(
		"%s.%s",
		sanitize.BaseName(fileName[:len(fileName)-len(ext)]),
		cleanExt[1:],
	), "-", "_", -1)
}

func createFormReader(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}
	return strings.NewReader(form.Encode())
}

func createMultipartReader(boundary string, data map[string][]byte) io.Reader {
	dashBoundary := "--" + boundary

	body := []byte{}
	buffer := bytes.NewBuffer(body)

	buffer.WriteString("Content-type: multipart/form-data; boundary=" + boundary + "\n\n")
	for contentType, content := range data {
		buffer.WriteString(dashBoundary + "\n")
		buffer.WriteString("Content-Disposition: form-data; name=" + contentType + "\n")
		buffer.WriteString(fmt.Sprintf("Content-Length: %d \n\n", len(content)))
		buffer.Write(content)
		buffer.WriteString("\n")
	}
	buffer.WriteString(dashBoundary + "--\n\n")
	return buffer
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func isYesString(s string) bool {
	switch strings.ToLower(s) {
	case "1", "yes", "true", "y":
		return true
	}
	return false
}

func isMatchingFilter(fs []*regexp.Regexp, d []byte) bool {
	for _, r := range fs {
		if r.Match(d) {
			return true
		}
	}
	return false
}
