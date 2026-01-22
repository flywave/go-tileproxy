package crawler

import (
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"compress/gzip"

	"github.com/gobwas/glob"
)

// 对象池优化，复用buffer减少内存分配
var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 32*1024) // 32KB初始容量
		},
	}
	gzipReaderPool = sync.Pool{
		New: func() interface{} {
			return &gzip.Reader{}
		},
	}
)

type httpBackend struct {
	LimitRules  []*LimitRule
	Client      *http.Client
	lock        *sync.RWMutex
	maxBodySize int
	cacheStats  struct {
		hits   uint64
		misses uint64
		mu     sync.RWMutex
	}
}

type checkHeadersFunc func(req *http.Request, statusCode int, header http.Header) bool

type LimitRule struct {
	DomainRegexp   string
	DomainGlob     string
	Delay          time.Duration
	RandomDelay    time.Duration
	Parallelism    int
	waitChan       chan bool
	compiledRegexp *regexp.Regexp
	compiledGlob   glob.Glob
}

func (r *LimitRule) Init() error {
	waitChanSize := 1
	if r.Parallelism > 1 {
		waitChanSize = r.Parallelism
	}
	r.waitChan = make(chan bool, waitChanSize)
	hasPattern := false
	if r.DomainRegexp != "" {
		c, err := regexp.Compile(r.DomainRegexp)
		if err != nil {
			return err
		}
		r.compiledRegexp = c
		hasPattern = true
	}
	if r.DomainGlob != "" {
		c, err := glob.Compile(r.DomainGlob)
		if err != nil {
			return err
		}
		r.compiledGlob = c
		hasPattern = true
	}
	if !hasPattern {
		return ErrNoPattern
	}
	return nil
}

func (h *httpBackend) Init(jar http.CookieJar) {
	h.Client = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}
	h.lock = &sync.RWMutex{}
	h.maxBodySize = 10 * 1024 * 1024 // 10MB default max body size
}

func (r *LimitRule) Match(domain string) bool {
	match := false
	if r.compiledRegexp != nil && r.compiledRegexp.MatchString(domain) {
		match = true
	}
	if r.compiledGlob != nil && r.compiledGlob.Match(domain) {
		match = true
	}
	return match
}

func (h *httpBackend) GetMatchingRule(domain string) *LimitRule {
	if h.LimitRules == nil {
		return nil
	}
	h.lock.RLock()
	defer h.lock.RUnlock()
	for _, r := range h.LimitRules {
		if r.Match(domain) {
			return r
		}
	}
	return nil
}

func (h *httpBackend) Cache(request *http.Request, bodySize int, checkHeadersFunc checkHeadersFunc, cacheDir string) (*Response, error) {
	if cacheDir == "" || request.Method != "GET" || request.Header.Get("Cache-Control") == "no-cache" {
		return h.Do(request, bodySize, checkHeadersFunc)
	}

	// 优化哈希计算，使用对象池
	sum := sha1.Sum([]byte(request.URL.String()))
	hash := hex.EncodeToString(sum[:])
	dir := path.Join(cacheDir, hash[:2])
	filename := path.Join(dir, hash)

	// 先尝试读取缓存
	if resp, err := h.loadFromCache(filename, request, checkHeadersFunc); err == nil {
		h.cacheStats.mu.Lock()
		h.cacheStats.hits++
		h.cacheStats.mu.Unlock()
		return resp, nil
	}

	// 缓存未命中，发起请求
	resp, err := h.Do(request, bodySize, checkHeadersFunc)
	if err != nil || resp.StatusCode >= 500 {
		return resp, err
	}

	h.cacheStats.mu.Lock()
	h.cacheStats.misses++
	h.cacheStats.mu.Unlock()

	// 异步保存缓存，不阻塞响应
	go h.saveToCache(resp, dir, filename)

	return resp, nil
}

// loadFromCache 从缓存加载响应
func (h *httpBackend) loadFromCache(filename string, request *http.Request, checkHeadersFunc checkHeadersFunc) (*Response, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	resp := new(Response)
	if err := gob.NewDecoder(file).Decode(resp); err != nil {
		return nil, err
	}

	if !checkHeadersFunc(request, resp.StatusCode, *resp.Headers) {
		return nil, ErrAbortedAfterHeaders
	}

	if resp.StatusCode >= 500 {
		return nil, errors.New("cached error response")
	}

	return resp, nil
}

// saveToCache 异步保存响应到缓存
func (h *httpBackend) saveToCache(resp *Response, dir, filename string) {
	// 创建目录
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return // 忽略缓存错误
		}
	}

	// 使用临时文件写入，再原子性重命名
	tempFile := filename + "~"
	file, err := os.Create(tempFile)
	if err != nil {
		return
	}
	defer file.Close()

	if err := gob.NewEncoder(file).Encode(resp); err != nil {
		os.Remove(tempFile) // 清理失败的临时文件
		return
	}

	file.Close()
	os.Rename(tempFile, filename)
}

func (h *httpBackend) Do(request *http.Request, bodySize int, checkHeadersFunc checkHeadersFunc) (*Response, error) {
	r := h.GetMatchingRule(request.URL.Host)
	if r != nil {
		r.waitChan <- true
		defer func(r *LimitRule) {
			randomDelay := time.Duration(0)
			if r.RandomDelay != 0 {
				randomDelay = time.Duration(rand.Int63n(int64(r.RandomDelay)))
			}
			time.Sleep(r.Delay + randomDelay)
			<-r.waitChan
		}(r)
	}

	res, err := h.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.Request != nil {
		*request = *res.Request
	}
	if !checkHeadersFunc(request, res.StatusCode, res.Header) {
		return nil, ErrAbortedAfterHeaders
	}

	// 使用对象池获取buffer
	buf := bufferPool.Get().([]byte)
	buf = buf[:0] // 重置长度但保留容量
	defer bufferPool.Put(buf)

	var bodyReader io.Reader = res.Body
	if bodySize > 0 {
		bodyReader = io.LimitReader(bodyReader, int64(bodySize))
	}

	// 优化gzip处理
	contentEncoding := strings.ToLower(res.Header.Get("Content-Encoding"))
	if !res.Uncompressed && (strings.Contains(contentEncoding, "gzip") ||
		(contentEncoding == "" && strings.Contains(strings.ToLower(res.Header.Get("Content-Type")), "gzip")) ||
		strings.HasSuffix(strings.ToLower(request.URL.Path), ".xml.gz")) {

		gzReader := gzipReaderPool.Get().(*gzip.Reader)

		if err := gzReader.Reset(bodyReader); err != nil {
			gzipReaderPool.Put(gzReader)
			return nil, err
		}
		bodyReader = gzReader
		defer func() {
			gzReader.Close()
			gzipReaderPool.Put(gzReader)
		}()
	}

	// 使用预分配buffer读取数据
	body, err := h.readBody(bodyReader, buf)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: res.StatusCode,
		Body:       body,
		Headers:    &res.Header,
	}, nil
}

// readBody 优化的读取方法
func (h *httpBackend) readBody(reader io.Reader, buf []byte) ([]byte, error) {
	for {
		// 如果已经达到最大大小，停止读取
		if h.maxBodySize > 0 && len(buf) >= h.maxBodySize {
			return nil, fmt.Errorf("body size exceeds maximum limit of %d bytes", h.maxBodySize)
		}

		n, err := reader.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// 如果buffer满了，需要扩容
		if len(buf) == cap(buf) {
			newCap := cap(buf) * 2
			if h.maxBodySize > 0 && newCap > h.maxBodySize {
				newCap = h.maxBodySize
			}
			newBuf := make([]byte, len(buf), newCap)
			copy(newBuf, buf)
			buf = newBuf
		}
	}

	// 复制数据到新的slice，避免引用原始的大buffer
	result := make([]byte, len(buf))
	copy(result, buf)
	return result, nil
}

func (h *httpBackend) Limit(rule *LimitRule) error {
	h.lock.Lock()
	if h.LimitRules == nil {
		h.LimitRules = make([]*LimitRule, 0, 8)
	}
	h.LimitRules = append(h.LimitRules, rule)
	h.lock.Unlock()
	return rule.Init()
}

func (h *httpBackend) Limits(rules []*LimitRule) error {
	for _, r := range rules {
		if err := h.Limit(r); err != nil {
			return err
		}
	}
	return nil
}
