package service

import (
	"net/http"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type Service interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	RequestParser(r *http.Request) request.Request
}

type BaseService struct {
	Service
	router        map[string]func(w http.ResponseWriter, r *http.Request)
	requestParser func(r *http.Request) request.Request
}

func (s *BaseService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := s.RequestParser(r)
	handler := req.GetRequestHandler()
	if h, ok := s.router[handler]; ok {
		h(w, r)
	}
}

func (s *BaseService) RequestParser(r *http.Request) request.Request {
	if s.requestParser != nil {
		return s.requestParser(r)
	}
	return nil
}

func (s *BaseService) DecorateTile(image tile.Source, service string, layers []string, query_extent *geo.MapExtent) tile.Source {
	return image
}
