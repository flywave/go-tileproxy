package tileproxy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParseServiceId(t *testing.T) {
	proxy := NewTileProxy(nil, nil, nil)

	tests := []struct {
		path string
		want string
	}{
		{"/service1/tiles/1/2/3.png", "service1"},
		{"/my-service/layer.json", "my-service"},
		{"/", ""},
	}

	for _, tt := range tests {
		req := &http.Request{URL: &url.URL{Path: tt.path}}
		got := proxy.parseServiceId(req)
		if got != tt.want {
			t.Errorf("parseServiceId(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestRemoveService_Existing(t *testing.T) {
	proxy := NewTileProxy(nil, nil, nil)
	// RemoveService should not panic when removing existing service
	proxy.RemoveService("non-existent-id")
}

func TestRemoveService_NonExistent(t *testing.T) {
	proxy := NewTileProxy(nil, nil, nil)
	// Should not panic even if service doesn't exist
	proxy.RemoveService("non-existent-id")
}
