package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
)

type Future struct {
	finished   bool
	result     *crawler.Response
	resultchan chan *crawler.Response
	error_     error
	once       sync.Once
	mu         sync.RWMutex
}

func (f *Future) GetResult() *crawler.Response {
	// 快速检查是否已完成（使用读锁）
	f.mu.RLock()
	if f.finished {
		result := f.result
		f.mu.RUnlock()
		return result
	}
	f.mu.RUnlock()

	// 使用sync.Once确保channel只初始化一次
	f.once.Do(func() {
		f.mu.Lock()
		if !f.finished && f.resultchan == nil {
			f.resultchan = make(chan *crawler.Response, 1)
		}
		f.mu.Unlock()
	})

	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		f.mu.Lock()
		if !f.finished {
			f.finished = true
			f.result = nil
			f.error_ = errors.New("timeout")
			// 安全关闭channel
			if f.resultchan != nil {
				close(f.resultchan)
				f.resultchan = nil
			}
		}
		f.mu.Unlock()
		return nil
	case result := <-f.resultchan:
		f.mu.Lock()
		if !f.finished {
			f.finished = true
			f.result = result
			f.error_ = nil
			// 安全关闭channel
			if f.resultchan != nil {
				close(f.resultchan)
				f.resultchan = nil
			}
		}
		f.mu.Unlock()
		return result
	}
}

func (f *Future) setResult(result *crawler.Response) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 检查是否已完成
	if f.finished {
		return
	}

	// 如果channel存在且未关闭，尝试发送结果
	if f.resultchan != nil {
		select {
		case f.resultchan <- result:
			// 成功发送
		default:
			// channel可能已满或已关闭，直接设置结果
			f.result = result
			f.finished = true
		}
	} else {
		// 直接设置结果
		f.result = result
		f.finished = true
	}
}

func newFuture() *Future {
	return &Future{
		finished:   false,
		result:     nil,
		resultchan: nil,
		error_:     nil,
	}
}
