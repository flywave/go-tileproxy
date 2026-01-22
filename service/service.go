package service

import (
	"encoding/json"
	"net/http"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/tile"
)

type Service interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	RequestParser(r *http.Request) request.Request
}

type BaseService struct {
	Service
	router        map[string]func(r request.Request) *Response
	requestParser func(r *http.Request) request.Request
}

func (s *BaseService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/health" || path == "/health/" {
		s.healthCheckHandler(w, r)
		return
	}

	req := s.RequestParser(r)
	if req == nil {
		w.WriteHeader(404)
		return
	}
	handler := req.GetRequestHandler()
	if h, ok := s.router[handler]; ok {
		resp := h(req)
		if resp != nil {
			resp.Write(w)
		} else {
			w.WriteHeader(404)
		}
	} else {
		w.WriteHeader(404)
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

func (s *BaseService) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	healthStatus := map[string]string{
		"status": "healthy",
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{"health": healthStatus})
}
