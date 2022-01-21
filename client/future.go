package client

import (
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
	l          sync.Mutex
}

func (f *Future) GetResult() *crawler.Response {
	f.l.Lock()
	if f.finished {
		return f.result
	}
	f.l.Unlock()

	f.resultchan = make(chan *crawler.Response, 1)

	ticker := time.NewTicker(time.Second * 60)
	select {
	case <-ticker.C:
		f.finished = true
		f.result = nil
		f.error_ = errors.New("timeout")
		return nil
	case f.result = <-f.resultchan:
		f.finished = true
		f.error_ = nil
		close(f.resultchan)
		f.resultchan = nil
		return f.result
	}
}

func (f *Future) setResult(result *crawler.Response) {
	f.l.Lock()
	defer f.l.Unlock()
	if f.finished {
		return
	}
	if f.resultchan != nil {
		f.resultchan <- result
	} else {
		f.result = result
	}
}

func newFuture() *Future {
	return &Future{finished: false, result: nil, resultchan: nil}
}
