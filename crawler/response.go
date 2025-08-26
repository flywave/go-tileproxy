package crawler

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Response struct {
	StatusCode int
	Body       []byte
	Ctx        *Context
	Request    *Request
	Headers    *http.Header
	Trace      *HTTPTrace
	UserData   interface{}
	// 性能优化：缓存文件名
	fileName string
	once     sync.Once
}

func (r *Response) Save(fileName string) error {
	return os.WriteFile(fileName, r.Body, 0644)
}

func (r *Response) FileName() string {
	r.once.Do(func() {
		r.fileName = r.calculateFileName()
	})
	return r.fileName
}

// calculateFileName 计算文件名，只调用一次
func (r *Response) calculateFileName() string {
	if r.Headers != nil {
		_, params, err := mime.ParseMediaType(r.Headers.Get("Content-Disposition"))
		if fName, ok := params["filename"]; ok && err == nil {
			return SanitizeFileName(fName)
		}
	}

	if r.Request != nil && r.Request.URL != nil {
		if r.Request.URL.RawQuery != "" {
			return SanitizeFileName(fmt.Sprintf("%s_%s", r.Request.URL.Path, r.Request.URL.RawQuery))
		}
		return SanitizeFileName(strings.TrimPrefix(r.Request.URL.Path, "/"))
	}

	return "unknown"
}
