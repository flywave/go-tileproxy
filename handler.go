package tileproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/flywave/go-tileproxy/filters"
	"github.com/flywave/go-tileproxy/utils"
)

type Handler struct {
	Listener         utils.Listener
	RequestFilters   []filters.RequestFilter
	RoundTripFilters []filters.RoundTripFilter
	ResponseFilters  []filters.ResponseFilter
	Branding         string
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var err error

	ctx := filters.NewContext(req.Context(), h, h.Listener, rw, h.Branding)
	req = req.WithContext(ctx)

	if req.Method != "CONNECT" && !req.URL.IsAbs() {
		if req.URL.Scheme == "" {
			if req.TLS != nil && req.ProtoMajor == 1 {
				req.URL.Scheme = "https"
			} else {
				req.URL.Scheme = "http"
			}
		}

		if req.TLS != nil {
			if req.Host == "" {
				if req.URL.Host != "" {
					req.Host = req.URL.Host
				} else {
					req.Host = req.TLS.ServerName
				}
			}
			if req.URL.Host == "" {
				if req.Host != "" {
					req.URL.Host = req.Host
				} else {
					req.URL.Host = req.TLS.ServerName
				}
			}
		}
	}

	for _, f := range h.RequestFilters {
		ctx, req, err = f.Request(ctx, req)
		if req == filters.DummyRequest {
			return
		}
		if err != nil {
			if err != io.EOF {
				//glog.Errorf("%s Filter Request %T error: %+v", remoteAddr, f, err)
			}
			return
		}
		req = req.WithContext(ctx)
	}

	if req.Body != nil {
		defer req.Body.Close()
	}

	var resp *http.Response
	for _, f := range h.RoundTripFilters {
		ctx, resp, err = f.RoundTrip(ctx, req)
		if resp == filters.DummyResponse {
			return
		}
		if err != nil {
			filters.SetRoundTripFilter(ctx, f)
			http.Error(rw, h.FormatError(ctx, err), http.StatusBadGateway)
			return
		}
		req = req.WithContext(ctx)
		if resp != nil {
			resp.Request = req
			filters.SetRoundTripFilter(ctx, f)
			break
		}
	}

	for _, f := range h.ResponseFilters {
		if resp == nil || resp == filters.DummyResponse {
			return
		}
		ctx, resp, err = f.Response(ctx, resp)
		if err != nil {
			http.Error(rw, h.FormatError(ctx, err), http.StatusBadGateway)
			return
		}
		req = req.WithContext(ctx)
	}

	if resp == nil {
		http.Error(rw, h.FormatError(ctx, fmt.Errorf("empty response")), http.StatusBadGateway)
		return
	}

	if resp.Header.Get("Content-Length") == "" && resp.ContentLength >= 0 {
		resp.Header.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	for key, values := range resp.Header {
		for _, value := range values {
			rw.Header().Add(key, value)
		}
	}
	rw.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		defer resp.Body.Close()
		_, err := utils.IOCopy(rw, resp.Body)
		if err != nil {
			isClosedConnError(err)
		}
	}
}

func (h Handler) FormatError(ctx context.Context, err error) string {
	return fmt.Sprintf(`{
    "type": "localproxy",
    "host": "%s",
    "software": "%s (go/%s %s/%s)",
    "filter": "%T",
    "error": "%s"
}
`, filters.GetListener(ctx).Addr().String(),
		h.Branding, runtime.Version(), runtime.GOOS, runtime.GOARCH,
		filters.GetRoundTripFilter(ctx),
		err.Error())
}

func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}

	str := err.Error()
	if strings.Contains(str, "use of closed network connection") {
		return true
	}

	if runtime.GOOS == "windows" {
		const WSAECONNABORTED = 10053
		const WSAECONNRESET = 10054
		if oe, ok := err.(*net.OpError); ok && (oe.Op == "read" || oe.Op == "write") {
			if se, ok := oe.Err.(*os.SyscallError); ok && (se.Syscall == "wsarecv" || se.Syscall == "wsasend") {
				if n, ok := se.Err.(syscall.Errno); ok {
					if n == WSAECONNRESET || n == WSAECONNABORTED {
						return true
					}
				}
			}
		}
	}
	return false
}
