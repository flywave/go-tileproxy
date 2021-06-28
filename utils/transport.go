package utils

import (
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
)

var (
	ReqWriteExcludeHeader = map[string]bool{
		"Vary":                true,
		"Via":                 true,
		"X-Forwarded-For":     true,
		"Proxy-Authorization": true,
		"Proxy-Connection":    true,
		"Upgrade":             true,
		"X-Chrome-Variations": true,
		"Connection":          true,
		"Cache-Control":       true,
	}
)

func FixRequestURL(req *http.Request) {
	if req.URL.Host == "" {
		switch {
		case req.Host != "":
			req.URL.Host = req.Host
		case req.TLS != nil:
			req.URL.Host = req.TLS.ServerName
		}
	}
}

func FixRequestHeader(req *http.Request) {
	if req.ContentLength > 0 {
		if req.Header.Get("Content-Length") == "" {
			req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
		}
	}
}

func CloneRequest(r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}

func GetHostName(req *http.Request) string {
	if host, _, err := net.SplitHostPort(req.Host); err == nil {
		return host
	} else {
		return req.Host
	}
}

func IsStaticRequest(req *http.Request) bool {
	switch path.Ext(req.URL.Path) {
	case "bmp", "gif", "ico", "jpeg", "jpg", "png", "tif", "tiff",
		"3gp", "3gpp", "avi", "f4v", "flv", "m4p", "mkv", "mp4",
		"mp4v", "mpv4", "rmvb", ".webp", ".js", ".css":
		return true
	case "":
		name := path.Base(req.URL.Path)
		if strings.Contains(name, "play") ||
			strings.Contains(name, "video") {
			return true
		}
	default:
		if req.Header.Get("Range") != "" ||
			strings.Contains(req.Host, "img.") ||
			strings.Contains(req.Host, "cache.") ||
			strings.Contains(req.Host, "video.") ||
			strings.Contains(req.Host, "static.") ||
			strings.HasPrefix(req.Host, "img") ||
			strings.HasPrefix(req.URL.Path, "/static") ||
			strings.HasPrefix(req.URL.Path, "/asset") ||
			strings.Contains(req.URL.Path, "static") ||
			strings.Contains(req.URL.Path, "asset") ||
			strings.Contains(req.URL.Path, "/cache/") {
			return true
		}
	}
	return false
}
