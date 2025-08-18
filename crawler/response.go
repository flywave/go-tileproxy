package crawler

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"
)

type Response struct {
	StatusCode int
	Body       []byte
	Ctx        *Context
	Request    *Request
	Headers    *http.Header
	Trace      *HTTPTrace
	UserData   interface{}
}

func (r *Response) Save(fileName string) error {
	return os.WriteFile(fileName, r.Body, 0644)
}

func (r *Response) FileName() string {
	_, params, err := mime.ParseMediaType(r.Headers.Get("Content-Disposition"))
	if fName, ok := params["filename"]; ok && err == nil {
		return SanitizeFileName(fName)
	}
	if r.Request.URL.RawQuery != "" {
		return SanitizeFileName(fmt.Sprintf("%s_%s", r.Request.URL.Path, r.Request.URL.RawQuery))
	}
	return SanitizeFileName(strings.TrimPrefix(r.Request.URL.Path, "/"))
}
