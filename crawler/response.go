package crawler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"golang.org/x/net/html/charset"
)

type Response struct {
	StatusCode int
	Body       []byte
	Ctx        *Context
	Request    *Request
	Headers    *http.Header
	Trace      *HTTPTrace
}

func (r *Response) Save(fileName string) error {
	return ioutil.WriteFile(fileName, r.Body, 0644)
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

func (r *Response) fixCharset(detectCharset bool, defaultEncoding string) error {
	if len(r.Body) == 0 {
		return nil
	}
	if defaultEncoding != "" {
		tmpBody, err := encodeBytes(r.Body, "text/plain; charset="+defaultEncoding)
		if err != nil {
			return err
		}
		r.Body = tmpBody
		return nil
	}
	contentType := strings.ToLower(r.Headers.Get("Content-Type"))

	if strings.Contains(contentType, "utf-8") || strings.Contains(contentType, "utf8") {
		return nil
	}
	tmpBody, err := encodeBytes(r.Body, contentType)
	if err != nil {
		return err
	}
	r.Body = tmpBody
	return nil
}

func encodeBytes(b []byte, contentType string) ([]byte, error) {
	r, err := charset.NewReader(bytes.NewReader(b), contentType)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}
