package scheduler

import "github.com/flywave/go-tileproxy/request"

type SimpleScheduler struct {
	queue chan *request.Request
}

func NewSimpleScheduler() *SimpleScheduler {
	ch := make(chan *request.Request, 1024)
	return &SimpleScheduler{ch}
}

func (this *SimpleScheduler) Push(requ *request.Request) {
	this.queue <- requ
}

func (this *SimpleScheduler) Poll() *request.Request {
	if len(this.queue) == 0 {
		return nil
	} else {
		return <-this.queue
	}
}

func (this *SimpleScheduler) Count() int {
	return len(this.queue)
}
