package service

import (
	"crypto/sha512"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flywave/go-tileproxy/utils"
)

const (
	Charset            = "utf-8"
	DefaultContentType = "text/plain"
	BlockSize          = 1024 * 32
)

var (
	_status_codes = map[int]string{
		100: "Continue",
		101: "Switching Protocols",
		200: "OK",
		201: "Created",
		202: "Accepted",
		203: "Non-Authoritative Information",
		204: "No Content",
		205: "Reset Content",
		206: "Partial Content",
		300: "Multiple Choices",
		301: "Moved Permanently",
		302: "Found",
		303: "See Other",
		304: "Not Modified",
		305: "Use Proxy",
		307: "Temporary Redirect",
		400: "Bad Request",
		401: "Unauthorized",
		402: "Payment Required",
		403: "Forbidden",
		404: "Not Found",
		405: "Method Not Allowed",
		406: "Not Acceptable",
		407: "Proxy Authentication Required",
		408: "Request Time-out",
		409: "Conflict",
		410: "Gone",
		411: "Length Required",
		412: "Precondition Failed",
		413: "Request Entity Too Large",
		414: "Request-URI Too Large",
		415: "Unsupported Media Type",
		416: "Requested range not satisfiable",
		417: "Expectation Failed",
		500: "Internal Server Error",
		501: "Not Implemented",
		502: "Bad Gateway",
		503: "Service Unavailable",
		504: "Gateway Time-out",
		505: "HTTP Version not supported",
	}
)

func StatusCode(code int) string {
	return strconv.Itoa(code) + " " + _status_codes[code]
}

type Response struct {
	response      []byte
	status        int
	timestamp     *time.Time
	last_modified time.Time
	headers       http.Header
}

func NewResponse(response []byte, status int, mimetype string) *Response {
	if status == -1 {
		status = 200
	}
	r := &Response{response: response, status: status, timestamp: nil, headers: make(http.Header)}
	var content_type string
	if mimetype != "" {
		if strings.HasPrefix(mimetype, "text/") {
			content_type = mimetype + "; charset=" + Charset
		} else {
			content_type = mimetype
		}
	}
	if content_type == "" {
		content_type = DefaultContentType
	}
	r.headers["Content-type"] = []string{content_type}
	return r
}

func (r *Response) SetStatus(status int) {
	r.status = status
}

func (r *Response) GetStatus() int {
	return r.status
}

func (r *Response) SetLastModified(date time.Time) {
	r.timestamp = &date
	r.headers["Last-modified"] = []string{utils.FormatHTTPDate(*r.timestamp)}
}

func (r *Response) GetLastModified() *time.Time {
	if vs, ok := r.headers["Last-modified"]; ok {
		t, err := utils.ParseHTTPDate(vs[0])
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
}

func (r *Response) SetETag(value string) {
	r.headers["ETag"] = []string{value}
}

func (r *Response) GetETag() string {
	if vs, ok := r.headers["ETag"]; ok {
		return vs[0]
	}
	return ""
}

func etagFor(data []byte) string {
	return fmt.Sprintf("%X", sha512.Sum512(data))
}

func (r *Response) noCacheHeaders() {
	r.headers["Cache-Control"] = []string{"no-cache, no-store"}
	r.headers["Pragma"] = []string{"no-cache"}
	r.headers["Expires"] = []string{"-1"}
}

func (r *Response) cacheHeaders(timestamp *time.Time, etag_data []string, max_age int) {
	if etag_data != nil {
		hash_src := strings.Join(etag_data, "")
		r.SetETag(etagFor([]byte(hash_src)))
	}

	r.last_modified = *timestamp
	if (timestamp != nil || etag_data != nil) && max_age != -1 {
		r.headers["Cache-Control"] = []string{fmt.Sprintf("public, max-age=%d, s-maxage=%d", max_age, max_age)}
	}
}

func (r *Response) makeConditional(req *http.Request) {
	not_modified := false
	if v := req.Header.Get("If-none-match"); v == r.GetETag() {
		not_modified = true
	} else if r.timestamp != nil {
		if date := req.Header.Get("If-modified-since"); date != "" {
			timestamp, _ := utils.ParseDateTime(date)
			if r.timestamp.Before(timestamp) {
				not_modified = true
			}
		}
	}

	if not_modified {
		r.status = 304
		r.response = nil
		if _, ok := r.headers["Content-type"]; ok {
			delete(r.headers, "Content-type")
		}
	}
}

func (r *Response) GetContentLength() int {
	if vs, ok := r.headers["Content-length"]; ok {
		l, err := strconv.Atoi(vs[0])
		if err != nil {
			return 0
		}
		return l
	}
	return 0
}

func (r *Response) GetContentType() string {
	if vs, ok := r.headers["Content-type"]; ok {
		return vs[0]
	}
	return ""
}
