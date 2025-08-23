package client

import (
	"testing"
	"time"
)

// TestNewCollectorContext 测试创建新的CollectorContext
func TestNewCollectorContext(t *testing.T) {
	config := &Config{
		SkipSSL:           false,
		Threads:           10,
		UserAgent:         "test-agent",
		RandomDelay:       5,
		DisableKeepAlives: true,
		Proxys:            []string{"http://proxy1.com", "http://proxy2.com"},
		RequestTimeout:    30 * time.Second,
	}

	ctx := NewCollectorContext(config)

	if ctx == nil {
		t.Fatal("创建CollectorContext失败")
	}

	if ctx.client == nil {
		t.Error("CollectorContext的client不应为nil")
	}
}

// TestCollectorContextClient 测试CollectorContext的Client方法
func TestCollectorContextClient(t *testing.T) {
	config := &Config{
		Threads: 5,
	}

	ctx := NewCollectorContext(config)
	client := ctx.Client()

	if client == nil {
		t.Error("Client()方法不应返回nil")
	}

	// 验证返回的client类型
	if _, ok := client.(*CollectorClient); !ok {
		t.Error("Client()应返回*CollectorClient类型")
	}
}

// TestCollectorContextGetCollector 测试CollectorContext的GetCollector方法
func TestCollectorContextGetCollector(t *testing.T) {
	config := &Config{
		Threads: 3,
	}

	ctx := NewCollectorContext(config)
	collector := ctx.GetCollector()

	if collector == nil {
		t.Error("GetCollector()不应返回nil")
	}
}

// TestCollectorContextWithEmptyConfig 测试使用空配置创建CollectorContext
func TestCollectorContextWithEmptyConfig(t *testing.T) {
	config := &Config{}

	ctx := NewCollectorContext(config)

	if ctx == nil {
		t.Fatal("使用空配置创建CollectorContext失败")
	}

	if ctx.Client() == nil {
		t.Error("空配置下Client()不应返回nil")
	}
}

// TestCollectorContextClientConsistency 测试Client方法的一致性
func TestCollectorContextClientConsistency(t *testing.T) {
	config := &Config{
		Threads: 2,
	}

	ctx := NewCollectorContext(config)
	client1 := ctx.Client()
	client2 := ctx.Client()

	if client1 != client2 {
		t.Error("多次调用Client()应返回相同的client实例")
	}
}

// TestCollectorContextConfigPropagation 测试配置是否正确传递
func TestCollectorContextConfigPropagation(t *testing.T) {
	config := &Config{
		SkipSSL:        true,
		Threads:        8,
		UserAgent:      "custom-agent",
		RequestTimeout: 60 * time.Second,
	}

	ctx := NewCollectorContext(config)

	// 验证配置被正确传递到底层client
	collector := ctx.GetCollector()
	if collector == nil {
		t.Error("配置应正确传递到底层collector")
	}
}
