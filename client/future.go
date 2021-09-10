package client

import (
	"errors"
	"sync"
	"time"

	"github.com/flywave/go-tileproxy/crawler"
)

type Future struct {
	finished   bool
	req        []byte
	result     *crawler.Response
	resultchan chan *crawler.Response
	c          *crawler.Collector
	error_     error
	l          sync.Mutex
}

func (f *Future) setCollector(c *crawler.Collector) {
	f.c = c
}

func (f *Future) loadRequest() (*crawler.Request, error) {
	buf := f.req
	if buf == nil {
		return nil, errors.New("req nil")
	}
	copied := make([]byte, len(buf))
	copy(copied, buf)
	req, err := f.c.UnmarshalRequest(copied)
	if err == nil {
		req.UserData = f
	}
	return req, err
}

func (f *Future) Do() error {
	req, err := f.loadRequest()
	if err == nil {
		err = req.Do()
		if err != nil {
			return err
		}
	}
	return err
}

func (f *Future) GetResult() *crawler.Response {
	f.l.Lock()
	defer f.l.Unlock()
	if f.finished {
		return f.result
	}
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
		return f.result
	}
}

func (f *Future) setResult(result *crawler.Response) {
	if f.finished {
		return
	}
	f.resultchan <- result
	close(f.resultchan)
}

func newFuture(req []byte) *Future {
	return &Future{finished: false, result: nil, resultchan: make(chan *crawler.Response, 1), req: req}
}
