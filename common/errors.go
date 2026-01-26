package common

import (
	"fmt"
)

var (
	ErrServiceNotFound      = NewServiceError("SERVICE_NOT_FOUND", "service not found", nil)
	ErrCacheNotFound        = NewServiceError("CACHE_NOT_FOUND", "cache not found", nil)
	ErrSourceNotFound       = NewServiceError("SOURCE_NOT_FOUND", "source not found", nil)
	ErrInvalidConfig        = NewServiceError("INVALID_CONFIG", "invalid configuration", nil)
	ErrGridNotFound         = NewServiceError("GRID_NOT_FOUND", "grid not found", nil)
	ErrTileNotFound         = NewServiceError("TILE_NOT_FOUND", "tile not found", nil)
	ErrServiceAlreadyExists = NewServiceError("SERVICE_ALREADY_EXISTS", "service already exists", nil)
	ErrCoverageNotFound     = NewServiceError("COVERAGE_NOT_FOUND", "coverage not found", nil)
	ErrLayerNotFound        = NewServiceError("LAYER_NOT_FOUND", "layer not found", nil)
	ErrInvalidTileFormat    = NewServiceError("INVALID_TILE_FORMAT", "invalid tile format", nil)
	ErrInvalidSRS           = NewServiceError("INVALID_SRS", "invalid spatial reference system", nil)
	ErrCacheReadFailed      = NewServiceError("CACHE_READ_FAILED", "failed to read from cache", nil)
	ErrCacheWriteFailed     = NewServiceError("CACHE_WRITE_FAILED", "failed to write to cache", nil)
	ErrSourceRequestFailed  = NewServiceError("SOURCE_REQUEST_FAILED", "failed to request from source", nil)
)

type ServiceError struct {
	Code    string
	Message string
	Cause   error
}

func NewServiceError(code, message string, cause error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func (e *ServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *ServiceError) Unwrap() error {
	return e.Cause
}

func (e *ServiceError) GetCode() string {
	return e.Code
}

func WrapServiceError(code, message string, cause error) error {
	if cause == nil {
		return NewServiceError(code, message, nil)
	}
	return NewServiceError(code, message, cause)
}

func IsServiceError(err error, code string) bool {
	if se, ok := err.(*ServiceError); ok {
		return se.Code == code
	}
	return false
}

func GetErrorCode(err error) string {
	if se, ok := err.(*ServiceError); ok {
		return se.Code
	}
	return "UNKNOWN_ERROR"
}
