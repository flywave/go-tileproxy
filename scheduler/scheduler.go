package scheduler

import "github.com/flywave/go-tileproxy/request"

type Scheduler interface {
	Push(requ *request.Request)
	Poll() *request.Request
	Count() int
}
