package request

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bitly/go-simplejson"
)

type Request struct {
	Url string

	RespType string

	Method string

	Postdata string

	Urltag string

	Header http.Header

	Cookies []*http.Cookie

	ProxyHost string

	checkRedirect func(req *http.Request, via []*http.Request) error

	Meta interface{}
}

func NewRequest(url string, respType string, urltag string, method string,
	postdata string, header http.Header, cookies []*http.Cookie,
	checkRedirect func(req *http.Request, via []*http.Request) error,
	meta interface{}) *Request {
	return &Request{url, respType, method, postdata, urltag, header, cookies, "", checkRedirect, meta}
}

func NewRequestWithProxy(url string, respType string, urltag string, method string,
	postdata string, header http.Header, cookies []*http.Cookie, proxyHost string,
	checkRedirect func(req *http.Request, via []*http.Request) error,
	meta interface{}) *Request {
	return &Request{url, respType, method, postdata, urltag, header, cookies, proxyHost, checkRedirect, meta}
}

func NewRequestWithHeaderFile(url string, respType string, headerFile string) *Request {
	_, err := os.Stat(headerFile)
	if err != nil {
		return NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
	}

	h := readHeaderFromFile(headerFile)

	return NewRequest(url, respType, "", "GET", "", h, nil, nil, nil)
}

func readHeaderFromFile(headerFile string) http.Header {
	b, err := ioutil.ReadFile(headerFile)
	if err != nil {
		return nil
	}
	js, _ := simplejson.NewJson(b)

	h := make(http.Header)
	h.Add("User-Agent", js.Get("User-Agent").MustString())
	h.Add("Referer", js.Get("Referer").MustString())
	h.Add("Cookie", js.Get("Cookie").MustString())
	h.Add("Cache-Control", "max-age=0")
	h.Add("Connection", "keep-alive")
	return h
}

func (this *Request) AddHeaderFile(headerFile string) *Request {
	_, err := os.Stat(headerFile)
	if err != nil {
		return this
	}
	h := readHeaderFromFile(headerFile)
	this.Header = h
	return this
}

func (this *Request) AddProxyHost(host string) *Request {
	this.ProxyHost = host
	return this
}

func (this *Request) GetUrl() string {
	return this.Url
}

func (this *Request) GetUrlTag() string {
	return this.Urltag
}

func (this *Request) GetMethod() string {
	return this.Method
}

func (this *Request) GetPostdata() string {
	return this.Postdata
}

func (this *Request) GetHeader() http.Header {
	return this.Header
}

func (this *Request) GetCookies() []*http.Cookie {
	return this.Cookies
}

func (this *Request) GetProxyHost() string {
	return this.ProxyHost
}

func (this *Request) GetResponceType() string {
	return this.RespType
}

func (this *Request) GetRedirectFunc() func(req *http.Request, via []*http.Request) error {
	return this.checkRedirect
}

func (this *Request) GetMeta() interface{} {
	return this.Meta
}
