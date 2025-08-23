package client

import (
	"testing"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
)

// TestNewCollectorClient 测试创建新的CollectorClient
func TestNewCollectorClient(t *testing.T) {
	config := &Config{
		SkipSSL:           false,
		Threads:           2,
		UserAgent:         "test-agent",
		RandomDelay:       1,
		DisableKeepAlives: false,
		RequestTimeout:    5 * time.Second,
	}

	ctx := &crawler.Context{}
	client := NewCollectorClient(config, ctx)

	if client == nil {
		t.Fatal("创建CollectorClient失败")
	}

	if client.BaseRequest == nil {
		t.Error("BaseRequest不应为nil")
	}

	if client.BaseRequest.Ctx != ctx {
		t.Error("BaseRequest的Ctx应正确设置")
	}
}

// TestCollectorClientGetCollector 测试获取Collector
func TestCollectorClientGetCollector(t *testing.T) {
	config := &Config{Threads: 1}
	client := NewCollectorClient(config, nil)

	collector := client.GetCollector()
	if collector == nil {
		t.Error("GetCollector不应返回nil")
	}
}

// TestCollectorClientConfigValidation 测试配置验证
func TestCollectorClientConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "有效配置",
			config: &Config{Threads: 1, RequestTimeout: 5 * time.Second},
		},
		{
			name:   "零线程配置",
			config: &Config{Threads: 0, RequestTimeout: 5 * time.Second},
		},
		{
			name: "完整配置",
			config: &Config{
				SkipSSL:           true,
				Threads:           5,
				UserAgent:         "test-user-agent",
				RandomDelay:       2,
				DisableKeepAlives: true,
				RequestTimeout:    10 * time.Second,
				Proxys:            []string{"http://proxy:8080"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCollectorClient(tt.config, nil)
			if client == nil {
				t.Error("期望创建成功")
			}

			if client != nil && client.Collector == nil {
				t.Error("Collector不应为nil")
			}

			if client != nil && client.BaseRequest == nil {
				t.Error("BaseRequest不应为nil")
			}
		})
	}
}

// TestCollectorClientDefaultValues 测试默认值
func TestCollectorClientDefaultValues(t *testing.T) {
	config := &Config{} // 空配置
	client := NewCollectorClient(config, nil)

	if client == nil {
		t.Fatal("空配置应创建客户端")
	}

	if client.Collector == nil {
		t.Error("Collector不应为nil")
	}

	if client.BaseRequest == nil {
		t.Error("BaseRequest不应为nil")
	}
}

// TestCollectorClientWithContext 测试上下文设置
func TestCollectorClientWithContext(t *testing.T) {
	ctx := &crawler.Context{}
	config := &Config{Threads: 1}

	client := NewCollectorClient(config, ctx)

	if client.BaseRequest == nil {
		t.Fatal("BaseRequest不应为nil")
	}

	if client.BaseRequest.Ctx != ctx {
		t.Error("上下文应正确设置")
	}
}

// TestCollectorClientNilConfig 测试nil配置处理
func TestCollectorClientNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("nil配置导致panic: %v", r)
		}
	}()

	// 测试nil配置的处理
	client := NewCollectorClient(nil, nil)
	if client != nil {
		t.Log("nil配置被处理")
	}
}
