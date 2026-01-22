package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
)

var defaultFutureTimeout = 30 * time.Second

func SetDefaultFutureTimeout(timeout time.Duration) {
	defaultFutureTimeout = timeout
}

type Future struct {
	finished   bool
	result     *crawler.Response
	resultchan chan *crawler.Response
	error_     error
	once       sync.Once
	mu         sync.RWMutex
	timeout    time.Duration
}

func (f *Future) GetResult() *crawler.Response {
	f.mu.RLock()
	if f.finished {
		result := f.result
		f.mu.RUnlock()
		return result
	}
	f.mu.RUnlock()

	f.once.Do(func() {
		f.mu.Lock()
		if !f.finished && f.resultchan == nil {
			f.resultchan = make(chan *crawler.Response, 1)
		}
		f.mu.Unlock()
	})

	timeout := defaultFutureTimeout
	if f.timeout > 0 {
		timeout = f.timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		f.mu.Lock()
		if !f.finished {
			f.finished = true
			f.result = nil
			f.error_ = errors.New("timeout")
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

	if f.finished {
		return
	}

	if f.resultchan != nil {
		select {
		case f.resultchan <- result:
		default:
			f.result = result
			f.finished = true
		}
	} else {
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
		timeout:    defaultFutureTimeout,
	}
}

func newFutureWithTimeout(timeout time.Duration) *Future {
	return &Future{
		finished:   false,
		result:     nil,
		resultchan: nil,
		error_:     nil,
		timeout:    timeout,
	}
}
