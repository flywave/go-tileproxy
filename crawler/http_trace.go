package crawler

import (
	"net/http"
	"net/http/httptrace"
	"time"
)

type HTTPTrace struct {
	start, connect    time.Time
	ConnectDuration   time.Duration
	FirstByteDuration time.Duration
}

func (ht *HTTPTrace) trace() *httptrace.ClientTrace {
	trace := &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) { ht.connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			ht.ConnectDuration = time.Since(ht.connect)
		},

		GetConn: func(hostPort string) { ht.start = time.Now() },
		GotFirstResponseByte: func() {
			ht.FirstByteDuration = time.Since(ht.start)
		},
	}
	return trace
}

func (ht *HTTPTrace) WithTrace(req *http.Request) *http.Request {
	return req.WithContext(httptrace.WithClientTrace(req.Context(), ht.trace()))
}
