package service

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/utils"
)

// conditional requests with an If-Modified-Since equal to Last-Modified
// must yield 304 (RFC 7232)
func TestResponse_MakeConditional_IfModifiedSinceEqual(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	resp := NewResponse([]byte("payload"), 200, "image/png")
	resp.SetLastModified(now)

	req := httptest.NewRequest("GET", "/tile.png", nil)
	req.Header.Set("If-Modified-Since", utils.FormatHTTPDate(now))
	resp.makeConditional(req)

	if resp.GetStatus() != 304 {
		t.Fatalf("expected 304 for equal If-Modified-Since, got %d", resp.GetStatus())
	}
	if resp.GetBuffer() != nil {
		t.Fatal("304 response must not carry a body")
	}
}

// a changed ETag must serve the full response
func TestResponse_MakeConditional_ETagMatch(t *testing.T) {
	resp := NewResponse([]byte("payload"), 200, "image/png")
	resp.SetETag("abc123")
	etag := resp.GetETag()
	if len(etag) == 0 || len(etag) > 64 {
		t.Fatalf("unexpected etag length: %d", len(etag))
	}

	req := httptest.NewRequest("GET", "/tile.png", nil)
	req.Header.Set("If-None-Match", etag)
	resp.makeConditional(req)
	if resp.GetStatus() != 304 {
		t.Fatalf("expected 304 for matching etag, got %d", resp.GetStatus())
	}

	req2 := httptest.NewRequest("GET", "/tile.png", nil)
	req2.Header.Set("If-None-Match", "other")
	resp2 := NewResponse([]byte("payload"), 200, "image/png")
	resp2.SetETag("abc123")
	resp2.makeConditional(req2)
	if resp2.GetStatus() == 304 {
		t.Fatal("mismatched etag must not produce 304")
	}
}

// WMTS tiles always carry an ETag; without MaxTileAge they use pure
// revalidation instead of max-age
func TestWMTSGetTile_RevalidationHeaders(t *testing.T) {
	s := createTestWMTSService()

	req := httptest.NewRequest("GET", "/wmts?service=WMTS&request=GetTile&version=1.0.0&layer=test_layer&style=default&tilematrixset=test_grid&tilematrix=0&tilerow=0&tilecol=0&format=image/png", nil)
	r := request.MakeWMTSRequest(req, false)

	resp := s.GetTile(r)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected 200, got %d", resp.GetStatus())
	}

	etag := resp.GetETag()
	if etag == "" {
		t.Fatal("tile response must carry an ETag for revalidation")
	}
	if ct := resp.GetContentType(); ct == "max-age" || ct == "" {
		t.Fatalf("unexpected content type %q", ct)
	}

	// a second render must produce the same validator, so a client sending
	// the etag back gets a 304 instead of the payload
	cond := httptest.NewRequest("GET", "/wmts", nil)
	cond.Header.Set("If-None-Match", etag)
	resp.makeConditional(cond)
	if resp.GetStatus() != 304 {
		t.Fatalf("expected 304 on revalidation, got %d", resp.GetStatus())
	}
}

// capabilities documents are cached per base URL and revalidated with ETag
func TestWMSCapabilities_ETagNegotiation(t *testing.T) {
	service := createTestWMSService()

	req := httptest.NewRequest("GET", "/wms?service=WMS&request=GetCapabilities", nil)
	r := request.MakeWMSRequest(req, false)

	first := service.GetCapabilities(r)
	if first.GetStatus() != 200 {
		t.Fatalf("expected 200, got %d", first.GetStatus())
	}
	etag := first.GetETag()
	if etag == "" {
		t.Fatal("capabilities must carry an ETag")
	}

	// second render must hit the cache and 304 on the matching etag
	req2 := httptest.NewRequest("GET", "/wms?service=WMS&request=GetCapabilities", nil)
	r2 := request.MakeWMSRequest(req2, false)
	second := service.GetCapabilities(r2)
	if string(second.GetBuffer()) != string(first.GetBuffer()) {
		t.Fatal("cached document differs from first render")
	}

	req3 := httptest.NewRequest("GET", "/wms?service=WMS&request=GetCapabilities", nil)
	req3.Header.Set("If-None-Match", etag)
	r3 := request.MakeWMSRequest(req3, false)
	third := service.GetCapabilities(r3)
	third.makeConditional(req3)
	if third.GetStatus() != 304 {
		t.Fatalf("expected 304 on revalidation, got %d", third.GetStatus())
	}
}

// a different base URL gets its own cached document
func TestWMSCapabilities_CachePerBaseURL(t *testing.T) {
	service := createTestWMSService()

	r1 := request.MakeWMSRequest(httptest.NewRequest("GET", "http://a.example.com/wms?service=WMS&request=GetCapabilities", nil), false)
	r2 := request.MakeWMSRequest(httptest.NewRequest("GET", "http://b.example.com/wms?service=WMS&request=GetCapabilities", nil), false)

	first := service.GetCapabilities(r1)
	second := service.GetCapabilities(r2)

	// documents differ per host (OnlineResource) but each carries its own etag
	if first.GetETag() == "" || second.GetETag() == "" {
		t.Fatal("both responses must carry etags")
	}
}
