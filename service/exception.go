package service

import (
	"fmt"

	"github.com/flywave/go-tileproxy/request"
)

type RequestError struct {
	Message  string
	Code     string
	Request  request.Request
	Internal bool
	Status   *int
}

func NewRequestError(message string, code string, request request.Request, internal bool, status *int) *RequestError {
	return &RequestError{Message: message, Code: code, Request: request, Internal: internal, Status: status}
}

func (e *RequestError) Render() *Response {
	var resp *Response
	if e.Request != nil {
		handler := e.Request.GetExceptionHandler().(ExceptionHandler)
		resp = handler.Render(e)
	} else if e.Status != nil {
		resp = NewResponse([]byte(e.Message), *e.Status, "text/plain")
	} else {
		resp = NewResponse([]byte(fmt.Sprintf("internal error: %s", e.Message)), 500, "text/plain")
	}
	resp.noCacheHeaders()
	return resp
}

func (e *RequestError) ToString() string {
	return fmt.Sprintf("RequestError(\"%s\", code=%s, request=%s)", e.Message, e.Code, e.Request.ToString())
}

type ExceptionHandler interface {
	Render(err *RequestError) *Response
}

type PlainExceptionHandler struct {
	ExceptionHandler
}

func (h *PlainExceptionHandler) Render(request_error *RequestError) *Response {
	var status_code int
	if request_error.Internal {
		status_code = 500
	}
	return NewResponse([]byte(request_error.Message), status_code, "text/plain")
}
