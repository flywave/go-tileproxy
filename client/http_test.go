package client

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	whatwgUrl "github.com/nlnwa/whatwg-url/url"

	"github.com/flywave/go-tileproxy/crawler"
)

var (
	httpConf = Config{
		SkipSSL:           false,
		Threads:           2,
		UserAgent:         "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0",
		RandomDelay:       2,
		DisableKeepAlives: false,
		Proxys:            nil,
		RequestTimeout:    time.Duration(20 * time.Second),
	}
)

func TestHttpFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(serverHandler))
	defer server.Close()
	rng := rand.New(rand.NewSource(12387123712321232))
	var (
		requests uint32
		success  uint32
		failure  uint32
	)
	client := NewCollectorClient(&httpConf, nil)
	client.Collector.OnResponse(func(resp *crawler.Response) {
		if resp.StatusCode == http.StatusOK {
			atomic.AddUint32(&success, 1)
		} else {
			atomic.AddUint32(&failure, 1)
		}

		if resp.UserData != nil {
			fut := resp.UserData.(*Future)
			fut.setResult(resp)
		}
	})

	futs := []*Future{}

	for i := 0; i < 30; i++ {
		ti := time.Duration(rng.Intn(50)) * time.Microsecond
		uri := server.URL + "/delay?t=" + ti.String()

		u, _ := whatwgUrl.Parse(uri)
		u2, _ := url.Parse(u.Href(false))

		atomic.AddUint32(&requests, 1)

		fut := newFuture()

		futs = append(futs, fut)

		client.Collector.Visit(u2.String(), fut)
	}

	client.Collector.Wait()

	if success+failure != requests || failure > 0 || len(futs) == 0 {
		t.Fatalf("wrong Queue implementation: "+
			" requests = %d, success = %d, failure = %d",
			requests, success, failure)
	}
}

func serverHandler(w http.ResponseWriter, req *http.Request) {
	if !serverRoute(w, req) {
		shutdown(w)
	}
}

func serverRoute(w http.ResponseWriter, req *http.Request) bool {
	if req.URL.Path == "/delay" {
		return serveDelay(w, req) == nil
	}
	return false
}

func serveDelay(w http.ResponseWriter, req *http.Request) error {
	q := req.URL.Query()
	t, err := time.ParseDuration(q.Get("t"))
	if err != nil {
		return err
	}
	time.Sleep(t)
	w.WriteHeader(http.StatusOK)
	return nil
}

func shutdown(w http.ResponseWriter) {
	taker, ok := w.(http.Hijacker)
	if !ok {
		return
	}
	raw, _, err := taker.Hijack()
	if err != nil {
		return
	}
	raw.Close()
}
