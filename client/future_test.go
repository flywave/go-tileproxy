package client

import (
	"testing"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
)

// TestNewFuture 测试创建新的Future实例
func TestNewFuture(t *testing.T) {
	future := newFuture()

	if future == nil {
		t.Fatal("创建Future失败")
	}

	if future.finished {
		t.Error("新创建的Future不应标记为完成")
	}

	if future.result != nil {
		t.Error("新创建的Future的result应为nil")
	}

	if future.resultchan != nil {
		t.Error("新创建的Future的resultchan应为nil")
	}
}

// TestFutureSetResultBasic 测试基本设置结果功能
func TestFutureSetResultBasic(t *testing.T) {
	future := newFuture()
	mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("test data")}

	// 验证初始状态
	if future.finished {
		t.Error("初始状态不应完成")
	}

	// 设置结果
	future.setResult(mockResponse)

	// setResult会直接设置finished=true当没有channel时

	// 验证结果已保存
	future.mu.RLock()
	if future.result != mockResponse {
		t.Error("Future应保存正确的结果")
	}
	if !future.finished {
		t.Error("Future应标记为完成")
	}
	future.mu.RUnlock()
}

// TestFutureSetResultMultipleTimes 测试多次设置结果
func TestFutureSetResultMultipleTimes(t *testing.T) {
	future := newFuture()
	mockResponse1 := &crawler.Response{StatusCode: 200, Body: []byte("data1")}
	mockResponse2 := &crawler.Response{StatusCode: 404, Body: []byte("data2")}

	// 第一次设置结果
	future.setResult(mockResponse1)

	// 第二次设置结果（应该被忽略，因为第一次调用后finished=true）
	future.setResult(mockResponse2)

	// 验证保留了第一次的结果
	future.mu.RLock()
	if future.result != mockResponse1 {
		t.Error("应保留第一次设置的结果")
	}
	future.mu.RUnlock()
}

// TestFutureGetResultWithPreSetResult 测试有预设结果时的GetResult
func TestFutureGetResultWithPreSetResult(t *testing.T) {
	future := newFuture()
	mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("preset data")}

	// 先设置结果
	future.setResult(mockResponse)

	// GetResult应该立即返回预设的结果
	result := future.GetResult()

	if result != mockResponse {
		t.Error("应返回预设的结果")
	}

	// 验证状态
	future.mu.RLock()
	if !future.finished {
		t.Error("Future应标记为完成")
	}
	future.mu.RUnlock()
}

// TestFutureNilResponse 测试设置nil响应
func TestFutureNilResponse(t *testing.T) {
	future := newFuture()

	// 设置nil结果
	future.setResult(nil)

	// 验证结果已保存
	future.mu.RLock()
	if future.result != nil {
		t.Error("应保存nil结果")
	}
	if !future.finished {
		t.Error("Future应标记为完成")
	}
	future.mu.RUnlock()
}

// TestFutureImmediateState 测试立即状态检查
func TestFutureImmediateState(t *testing.T) {
	future := newFuture()
	mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("immediate")}

	// 手动设置状态
	future.mu.Lock()
	future.finished = true
	future.result = mockResponse
	future.mu.Unlock()

	// 测试GetResult应该立即返回
	result := future.GetResult()
	if result != mockResponse {
		t.Error("应返回立即完成的结果")
	}

	// 验证状态
	future.mu.RLock()
	if !future.finished {
		t.Error("Future应标记为完成")
	}
	future.mu.RUnlock()
}

// TestFutureSetResultAfterCompletion 测试完成后再设置结果
func TestFutureSetResultAfterCompletion(t *testing.T) {
	future := newFuture()
	mockResponse1 := &crawler.Response{StatusCode: 200, Body: []byte("first")}
	mockResponse2 := &crawler.Response{StatusCode: 404, Body: []byte("second")}

	// 手动设置为完成状态
	future.mu.Lock()
	future.finished = true
	future.result = mockResponse1
	future.mu.Unlock()

	// 尝试再设置结果（应该被忽略）
	future.setResult(mockResponse2)

	// 验证结果没有改变
	future.mu.RLock()
	if future.result != mockResponse1 {
		t.Error("完成后不应再改变结果")
	}
	future.mu.RUnlock()
}

// TestFutureChannelBehavior 测试channel行为
func TestFutureChannelBehavior(t *testing.T) {
	future := newFuture()
	mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("channel test")}

	// 创建channel
	future.resultchan = make(chan *crawler.Response, 1)

	// 设置结果
	future.setResult(mockResponse)

	// 验证结果通过channel发送
	select {
	case result := <-future.resultchan:
		if result != mockResponse {
			t.Error("channel应返回正确的结果")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("channel应该收到结果")
	}
}

// TestFutureSetResultWithoutChannel 测试没有channel时的行为
func TestFutureSetResultWithoutChannel(t *testing.T) {
	future := newFuture()
	mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("no channel")}

	// 确保没有channel
	if future.resultchan != nil {
		t.Error("初始状态不应有channel")
	}

	// 设置结果
	future.setResult(mockResponse)

	// 验证结果直接保存并设置为完成
	future.mu.RLock()
	if future.result != mockResponse {
		t.Error("应直接保存结果")
	}
	if !future.finished {
		t.Error("Future应标记为完成")
	}
	future.mu.RUnlock()
}

// TestFutureGetResultTimeout 测试超时功能
func TestFutureGetResultTimeout(t *testing.T) {
	future := newFuture()

	// 创建一个短超时的context来测试超时功能
	// 我们不能直接修改GetResult的超时时间，所以这个测试会很慢
	// 但我们可以测试基本的超时逻辑
	start := time.Now()

	// 启动一个goroutine，在1秒后设置结果，这样可以避免60秒的超时
	go func() {
		time.Sleep(100 * time.Millisecond)
		mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("delayed result")}
		future.setResult(mockResponse)
	}()

	result := future.GetResult()
	elapsed := time.Since(start)

	if result == nil {
		t.Error("应该在超时前收到结果")
	}

	if elapsed > time.Second {
		t.Error("不应该等待太长时间")
	}
}

// TestFutureConcurrentAccess 测试并发访问
func TestFutureConcurrentAccess(t *testing.T) {
	future := newFuture()
	mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("concurrent test")}

	// 启动多个goroutine并发访问
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 一半设置结果，一半读取状态
			if id%2 == 0 {
				future.setResult(mockResponse)
			} else {
				// 读取状态
				future.mu.RLock()
				_ = future.finished
				_ = future.result
				future.mu.RUnlock()
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// 继续
		case <-time.After(time.Second):
			t.Fatal("并发测试超时")
		}
	}

	// 验证最终状态
	future.mu.RLock()
	if !future.finished {
		t.Error("Future应该被标记为完成")
	}
	if future.result != mockResponse {
		t.Error("结果应该被正确设置")
	}
	future.mu.RUnlock()
}

// TestFutureChannelCommunication 测试channel通信
func TestFutureChannelCommunication(t *testing.T) {
	future := newFuture()
	mockResponse := &crawler.Response{StatusCode: 200, Body: []byte("channel comm")}

	// 手动创建channel模拟GetResult的初始化
	future.mu.Lock()
	future.resultchan = make(chan *crawler.Response, 1)
	future.mu.Unlock()

	// 在另一个goroutine中设置结果
	go func() {
		time.Sleep(50 * time.Millisecond)
		future.setResult(mockResponse)
	}()

	// 等待channel接收结果
	select {
	case result := <-future.resultchan:
		if result != mockResponse {
			t.Error("channel应该传递正确的结果")
		}
	case <-time.After(time.Second):
		t.Error("应该通过channel接收到结果")
	}
}
