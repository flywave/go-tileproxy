package service

import "net/http"

type Service interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type BaseService struct {
	Service
}
