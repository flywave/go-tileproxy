package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockHttpClient 模拟HTTP客户端
type MockHttpClient struct {
	responses map[string]*http.Response
	errors    map[string]error
	mu        sync.RWMutex
}

func NewMockHttpClient() *MockHttpClient {
	return &MockHttpClient{
		responses: make(map[string]*http.Response),
		errors:    make(map[string]error),
	}
}

func (m *MockHttpClient) SetResponse(url string, statusCode int, body string, headers map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	for k, v := range headers {
		resp.Header.Set(k, v)
	}

	m.responses[url] = resp
}

func (m *MockHttpClient) SetError(url string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[url] = err
}

// TestContext 测试Context功能
func TestContext(t *testing.T) {
	t.Run("BasicOperations", func(t *testing.T) {
		ctx := NewContext()

		// 测试Put和Get
		ctx.Put("key1", "value1")
		if got := ctx.Get("key1"); got != "value1" {
			t.Errorf("Expected 'value1', got '%s'", got)
		}

		// 测试GetAny
		ctx.Put("key2", 42)
		if got := ctx.GetAny("key2"); got != 42 {
			t.Errorf("Expected 42, got %v", got)
		}

		// 测试不存在的key
		if got := ctx.Get("nonexistent"); got != "" {
			t.Errorf("Expected empty string, got '%s'", got)
		}

		if got := ctx.GetAny("nonexistent"); got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})

	t.Run("GetWithExists", func(t *testing.T) {
		ctx := NewContext()
		ctx.Put("existing", "value")

		// 测试存在的key
		val, exists := ctx.GetWithExists("existing")
		if !exists {
			t.Error("Expected key to exist")
		}
		if val != "value" {
			t.Errorf("Expected 'value', got %v", val)
		}

		// 测试不存在的key
		val, exists = ctx.GetWithExists("nonexistent")
		if exists {
			t.Error("Expected key to not exist")
		}
		if val != nil {
			t.Errorf("Expected nil, got %v", val)
		}
	})

	t.Run("BatchPut", func(t *testing.T) {
		ctx := NewContext()
		data := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		ctx.BatchPut(data)

		if ctx.Get("key1") != "value1" {
			t.Error("BatchPut failed for key1")
		}
		if ctx.GetAny("key2") != 42 {
			t.Error("BatchPut failed for key2")
		}
		if ctx.GetAny("key3") != true {
			t.Error("BatchPut failed for key3")
		}

		// 测试空数据
		ctx.BatchPut(nil)
		ctx.BatchPut(map[string]interface{}{})
	})

	t.Run("TypeSafety", func(t *testing.T) {
		ctx := NewContext()
		ctx.Put("intValue", 42)

		// Get方法应该只返回字符串，对于非字符串值应该返回空字符串
		if got := ctx.Get("intValue"); got != "" {
			t.Errorf("Expected empty string for non-string value, got '%s'", got)
		}
	})

	t.Run("ObjectPooling", func(t *testing.T) {
		// 测试对象池复用
		ctx1 := NewContext()
		ctx1.Put("test", "value")
		ctx1.Release()

		ctx2 := NewContext()
		// 确保新的context是干净的
		if ctx2.GetAny("test") != nil {
			t.Error("Context pool not properly cleaned")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		ctx := NewContext()
		var wg sync.WaitGroup

		// 并发写入测试
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", i)
				value := fmt.Sprintf("value%d", i)
				ctx.Put(key, value)
			}(i)
		}

		// 并发读取测试
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", i%10)
				ctx.Get(key)
				ctx.GetAny(key)
			}(i)
		}

		wg.Wait()
	})
}

// TestCollector 测试Collector功能
func TestCollector(t *testing.T) {
	t.Run("NewCollector", func(t *testing.T) {
		c := NewCollector()

		if c.UserAgent != "crawler" {
			t.Errorf("Expected UserAgent 'crawler', got '%s'", c.UserAgent)
		}

		if c.MaxBodySize != 10*1024*1024 {
			t.Errorf("Expected MaxBodySize %d, got %d", 10*1024*1024, c.MaxBodySize)
		}

		if c.backend == nil {
			t.Error("Backend should not be nil")
		}
	})

	t.Run("CollectorOptions", func(t *testing.T) {
		userAgent := "test-agent"
		domains := []string{"example.com", "test.com"}
		maxBodySize := 5 * 1024 * 1024

		c := NewCollector(
			UserAgent(userAgent),
			AllowedDomains(domains...),
			MaxBodySize(maxBodySize),
			Async(true),
			DetectCharset(),
			TraceHTTP(),
		)

		if c.UserAgent != userAgent {
			t.Errorf("Expected UserAgent '%s', got '%s'", userAgent, c.UserAgent)
		}

		if len(c.AllowedDomains) != len(domains) {
			t.Errorf("Expected %d domains, got %d", len(domains), len(c.AllowedDomains))
		}

		if c.MaxBodySize != maxBodySize {
			t.Errorf("Expected MaxBodySize %d, got %d", maxBodySize, c.MaxBodySize)
		}

		if !c.Async {
			t.Error("Expected Async to be true")
		}

		if !c.DetectCharset {
			t.Error("Expected DetectCharset to be true")
		}

		if !c.TraceHTTP {
			t.Error("Expected TraceHTTP to be true")
		}
	})

	t.Run("DomainFiltering", func(t *testing.T) {
		c := NewCollector(AllowedDomains("example.com", "test.com"))

		if !c.isDomainAllowed("example.com") {
			t.Error("example.com should be allowed")
		}

		if !c.isDomainAllowed("test.com") {
			t.Error("test.com should be allowed")
		}

		if c.isDomainAllowed("forbidden.com") {
			t.Error("forbidden.com should not be allowed")
		}

		// 测试禁止域名
		c2 := NewCollector(DisallowedDomains("forbidden.com"))

		if c2.isDomainAllowed("forbidden.com") {
			t.Error("forbidden.com should not be allowed")
		}

		if !c2.isDomainAllowed("example.com") {
			t.Error("example.com should be allowed when not in disallowed list")
		}
	})

	t.Run("URLFiltering", func(t *testing.T) {
		allowedPattern := regexp.MustCompile(`^https://example\.com/`)
		disallowedPattern := regexp.MustCompile(`/admin/`)

		c := NewCollector(
			URLFilters(allowedPattern),
			DisallowedURLFilters(disallowedPattern),
		)

		testCases := []struct {
			url      string
			expected error
		}{
			{"https://example.com/page", nil},
			{"https://example.com/admin/secret", ErrForbiddenURL},
			{"https://other.com/page", ErrNoURLFiltersMatch},
			{"", ErrMissingURL},
		}

		for _, tc := range testCases {
			parsedURL, _ := url.Parse(tc.url)
			err := c.requestCheck(tc.url, parsedURL, "GET", nil, 0)
			if err != tc.expected {
				t.Errorf("URL %s: expected error %v, got %v", tc.url, tc.expected, err)
			}
		}
	})
}

// TestHTTPBackend 测试HTTP后端功能
func TestHTTPBackend(t *testing.T) {
	t.Run("LimitRule", func(t *testing.T) {
		rule := &LimitRule{
			DomainGlob:  "*.example.com",
			Delay:       100 * time.Millisecond,
			Parallelism: 2,
		}

		err := rule.Init()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !rule.Match("test.example.com") {
			t.Error("Rule should match test.example.com")
		}

		if rule.Match("other.com") {
			t.Error("Rule should not match other.com")
		}

		// 测试无模式的规则
		emptyRule := &LimitRule{}
		if err := emptyRule.Init(); err != ErrNoPattern {
			t.Errorf("Expected ErrNoPattern, got %v", err)
		}
	})

	t.Run("Cache", func(t *testing.T) {
		tempDir := t.TempDir()

		// 创建测试服务器
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("test response"))
		}))
		defer ts.Close()

		backend := &httpBackend{}
		backend.Init(nil)

		req, _ := http.NewRequest("GET", ts.URL, nil)

		checkFunc := func(req *http.Request, statusCode int, headers http.Header) bool {
			return true
		}

		// 第一次请求 - 应该从服务器获取
		resp1, err := backend.Cache(req, 0, checkFunc, tempDir)
		if err != nil {
			t.Errorf("First request failed: %v", err)
		}

		if resp1.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp1.StatusCode)
		}

		if string(resp1.Body) != "test response" {
			t.Errorf("Expected 'test response', got '%s'", string(resp1.Body))
		}

		// 等待缓存文件写入
		time.Sleep(100 * time.Millisecond)

		// 第二次请求 - 应该从缓存获取
		// 创建新的请求对象，避免重用可能导致的context问题
		req2, _ := http.NewRequest("GET", ts.URL, nil)
		resp2, err := backend.Cache(req2, 0, checkFunc, tempDir)
		if err != nil {
			t.Errorf("Second request failed: %v", err)
			return
		}

		if resp2 == nil {
			t.Error("Second response is nil")
			return
		}

		if resp2.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp2.StatusCode)
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		backend := &httpBackend{}
		backend.Init(nil)

		// 测试无效URL格式
		req, _ := http.NewRequest("GET", "not-a-valid-url", nil)

		checkFunc := func(req *http.Request, statusCode int, headers http.Header) bool {
			return true
		}

		_, err := backend.Do(req, 0, checkFunc)
		if err == nil {
			t.Error("Expected error for invalid URL format")
		}
	})
}

// TestRequest 测试Request功能
func TestRequest(t *testing.T) {
	t.Run("AbsoluteURL", func(t *testing.T) {
		c := NewCollector()
		baseURL, _ := url.Parse("https://example.com/path/")

		req := &Request{
			URL:       baseURL,
			collector: c,
		}

		testCases := []struct {
			input    string
			expected string
		}{
			{"relative.html", "https://example.com/path/relative.html"},
			{"/absolute.html", "https://example.com/absolute.html"},
			{"https://other.com/external.html", "https://other.com/external.html"},
			{"#fragment", ""},
			{"", "https://example.com/path/"},
		}

		for _, tc := range testCases {
			result := req.AbsoluteURL(tc.input)
			if result != tc.expected {
				t.Errorf("AbsoluteURL(%s): expected '%s', got '%s'", tc.input, tc.expected, result)
			}
		}
	})

	t.Run("Serialization", func(t *testing.T) {
		c := NewCollector()
		originalURL, _ := url.Parse("https://example.com/test")
		ctx := NewContext()
		ctx.Put("key", "value")

		req := &Request{
			URL:       originalURL,
			Method:    "POST",
			Depth:     2,
			Body:      bytes.NewReader([]byte("test body")),
			Ctx:       ctx,
			Headers:   &http.Header{"Content-Type": []string{"application/json"}},
			collector: c,
		}

		// 测试序列化
		data, err := req.Marshal()
		if err != nil {
			t.Errorf("Marshal failed: %v", err)
		}

		// 测试反序列化
		deserializedReq, err := c.UnmarshalRequest(data)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}

		if deserializedReq.URL.String() != originalURL.String() {
			t.Errorf("URL mismatch: expected %s, got %s", originalURL.String(), deserializedReq.URL.String())
		}

		if deserializedReq.Method != "POST" {
			t.Errorf("Method mismatch: expected POST, got %s", deserializedReq.Method)
		}

		if deserializedReq.Depth != 2 {
			t.Errorf("Depth mismatch: expected 2, got %d", deserializedReq.Depth)
		}
	})
}

// TestResponse 测试Response功能
func TestResponse(t *testing.T) {
	t.Run("FileName", func(t *testing.T) {
		// 测试从Content-Disposition获取文件名
		headers := http.Header{
			"Content-Disposition": []string{`attachment; filename="test.pdf"`},
		}

		req := &Request{
			URL: &url.URL{Path: "/download"},
		}

		resp := &Response{
			Headers: &headers,
			Request: req,
		}

		fileName := resp.FileName()
		// SanitizeFileName处理后，.pdf应该变成test.pdf，然后把-替换为_
		if fileName != "test.pdf" {
			t.Errorf("Expected 'test.pdf', got '%s'", fileName)
		}

		// 测试从URL路径获取文件名
		resp2 := &Response{
			Headers: &http.Header{},
			Request: &Request{
				URL: &url.URL{Path: "/path/document.html"},
			},
		}

		fileName2 := resp2.FileName()
		if !strings.Contains(fileName2, "document") {
			t.Errorf("Expected filename to contain 'document', got '%s'", fileName2)
		}

		// 测试缓存 - 多次调用应该返回相同结果
		fileName3 := resp2.FileName()
		if fileName2 != fileName3 {
			t.Error("FileName should be cached and return same result")
		}
	})

	t.Run("Save", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")

		resp := &Response{
			Body: []byte("test content"),
		}

		err := resp.Save(filePath)
		if err != nil {
			t.Errorf("Save failed: %v", err)
		}

		// 验证文件内容
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read saved file: %v", err)
		}

		if string(content) != "test content" {
			t.Errorf("File content mismatch: expected 'test content', got '%s'", string(content))
		}
	})
}

// TestCallbacks 测试回调函数
func TestCallbacks(t *testing.T) {
	t.Run("OnRequest", func(t *testing.T) {
		c := NewCollector()
		var requestCount int

		c.OnRequest(func(r *Request) {
			requestCount++
		})

		c.handleOnRequest(&Request{})

		if requestCount != 1 {
			t.Errorf("Expected 1 request callback, got %d", requestCount)
		}
	})

	t.Run("OnResponse", func(t *testing.T) {
		c := NewCollector()
		var responseCount int

		c.OnResponse(func(r *Response) {
			responseCount++
		})

		c.handleOnResponse(&Response{})

		if responseCount != 1 {
			t.Errorf("Expected 1 response callback, got %d", responseCount)
		}
	})

	t.Run("OnError", func(t *testing.T) {
		c := NewCollector()
		var errorCount int

		c.OnError(func(r *Response, err error) {
			errorCount++
		})

		err := c.handleOnError(&Response{}, fmt.Errorf("test error"), &Request{}, &Context{})

		if errorCount != 1 {
			t.Errorf("Expected 1 error callback, got %d", errorCount)
		}

		if err == nil {
			t.Error("Expected error to be returned")
		}
	})
}

// TestAsyncOperations 测试异步操作
func TestAsyncOperations(t *testing.T) {
	t.Run("AsyncRequests", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("async response"))
		}))
		defer ts.Close()

		c := NewCollector(Async(true))
		var responseReceived bool
		var mu sync.Mutex

		c.OnResponse(func(r *Response) {
			mu.Lock()
			responseReceived = true
			mu.Unlock()
		})

		err := c.Visit(ts.URL, nil, nil)
		if err != nil {
			t.Errorf("Visit failed: %v", err)
		}

		// 等待异步操作完成
		c.Wait()

		mu.Lock()
		if !responseReceived {
			t.Error("Expected response to be received")
		}
		mu.Unlock()
	})
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	t.Run("NilParameters", func(t *testing.T) {
		c := NewCollector()

		// 测试nil context
		err := c.scrape("https://example.com", "GET", 0, nil, nil, nil, nil)
		// 这应该不会panic，因为我们在代码中处理了nil context
		if err == nil {
			t.Log("scrape with nil context handled gracefully")
		}
	})

	t.Run("EmptyValues", func(t *testing.T) {
		c := NewCollector()

		// 测试空URL
		err := c.Visit("", nil, nil)
		if err != ErrMissingURL {
			t.Errorf("Expected ErrMissingURL, got %v", err)
		}
	})

	t.Run("LargeBody", func(t *testing.T) {
		c := NewCollector(MaxBodySize(100)) // 限制为100字节

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 返回超过限制的响应
			w.Write(bytes.Repeat([]byte("a"), 200))
		}))
		defer ts.Close()

		var receivedBody []byte
		c.OnResponse(func(r *Response) {
			receivedBody = r.Body
		})

		err := c.Visit(ts.URL, nil, nil)
		if err != nil {
			t.Errorf("Visit failed: %v", err)
		}

		// 应该只接收到100字节
		if len(receivedBody) > 100 {
			t.Errorf("Expected body size <= 100, got %d", len(receivedBody))
		}
	})
}

// TestStats 测试统计功能
func TestStats(t *testing.T) {
	c := NewCollector()

	// 测试String方法的输出
	statsStr := c.String()
	if !strings.Contains(statsStr, "Requests made:") {
		t.Error("Stats string should contain request information")
	}

	if !strings.Contains(statsStr, "Callbacks:") {
		t.Error("Stats string should contain callback information")
	}
}

// BenchmarkContext 性能测试
func BenchmarkContext(b *testing.B) {
	b.Run("Put", func(b *testing.B) {
		ctx := NewContext()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i%100)
			ctx.Put(key, i)
		}
	})

	b.Run("Get", func(b *testing.B) {
		ctx := NewContext()
		// 预填充数据
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%d", i)
			ctx.Put(key, fmt.Sprintf("value%d", i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i%100)
			ctx.Get(key)
		}
	})

	b.Run("ConcurrentAccess", func(b *testing.B) {
		ctx := NewContext()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					key := fmt.Sprintf("key%d", i%100)
					ctx.Put(key, i)
				} else {
					key := fmt.Sprintf("key%d", i%100)
					ctx.Get(key)
				}
				i++
			}
		})
	})
}

// BenchmarkCollector Collector性能测试
func BenchmarkCollector(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("benchmark response"))
	}))
	defer ts.Close()

	b.Run("SyncRequests", func(b *testing.B) {
		c := NewCollector()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Visit(ts.URL, nil, nil)
		}
	})

	b.Run("AsyncRequests", func(b *testing.B) {
		c := NewCollector(Async(true))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Visit(ts.URL, nil, nil)
		}
		c.Wait()
	})
}
