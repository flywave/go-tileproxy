package service

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/utils"
)

func createRoutingTestService() *TileService {
	grid := geo.NewTileGrid(map[string]interface{}{
		"srs":       geo.NewProj("EPSG:3857"),
		"bbox":      []float64{-20037508.34, -20037508.34, 20037508.34, 20037508.34},
		"tile_size": []uint32{256, 256},
		"origin":    geo.ORIGIN_NW,
		"levels":    20,
	})

	mockCache := &MockCacheManager{
		grid:          grid,
		format:        "png",
		requestFormat: "png",
		tileOptions:   &MockTileOptions{},
	}

	provider := NewTileProvider(&TileProviderOptions{
		Name:        "test-layer",
		Title:       "Test Layer",
		Metadata:    &TileProviderMetadata{Name: "test-layer", Title: "Test Layer"},
		TileManager: mockCache,
		Dimensions:  make(utils.Dimensions),
	})

	return NewTileService(&TileServiceOptions{
		Layers: map[string]Provider{
			"test-layer": provider,
		},
		Metadata: &TileMetadata{
			Title:    "Test Tile Service",
			Abstract: "routing tests",
			URL:      "http://localhost/tms/",
		},
	})
}

// the TMS root resource document must be reachable through HTTP routing at
// the bare service mount point; previously the route was unreachable
func TestTileService_RootResourceRouting(t *testing.T) {
	s := createRoutingTestService()

	for _, path := range []string{"/tms", "/tms/", "/tiles", "/tiles/"} {
		req := httptest.NewRequest("GET", path, nil)
		r := s.RequestParser(req)
		if r == nil {
			t.Fatalf("%s: request not parsed", path)
		}
		if r.GetRequestHandler() != "tms_root_resource" {
			t.Fatalf("%s: expected tms_root_resource handler, got %q", path, r.GetRequestHandler())
		}
		resp := s.RootResource(r)
		if resp.GetStatus() != 200 {
			t.Fatalf("%s: expected status 200, got %d", path, resp.GetStatus())
		}
		if !strings.HasPrefix(resp.GetContentType(), "text/xml") {
			t.Fatalf("%s: expected text/xml, got %q", path, resp.GetContentType())
		}
		if !strings.Contains(string(resp.GetBuffer()), "TileMapService") {
			t.Fatalf("%s: expected root resource XML, got %q", path, string(resp.GetBuffer()))
		}
	}
}

// tile requests must resolve regardless of the service mount point
func TestTileService_CustomMountPoint(t *testing.T) {
	s := createRoutingTestService()

	req := httptest.NewRequest("GET", "/mytiles/1.0.0/test-layer/0/0/0.png", nil)
	r := s.RequestParser(req)
	if r == nil {
		t.Fatal("request not parsed")
	}
	if r.GetRequestHandler() != "map" {
		t.Fatalf("expected map handler, got %q", r.GetRequestHandler())
	}

	resp := s.GetMap(r)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}
	if !strings.HasPrefix(resp.GetContentType(), "image/png") {
		t.Fatalf("expected image/png, got %q", resp.GetContentType())
	}
}

// tile responses must carry the tile format content type
func TestTileService_GetMap_ContentType(t *testing.T) {
	s := createRoutingTestService()

	req := httptest.NewRequest("GET", "/tiles/1.0.0/test-layer/0/0/0.png", nil)
	r := s.RequestParser(req)
	resp := s.GetMap(r)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}
	if !strings.HasPrefix(resp.GetContentType(), "image/png") {
		t.Fatalf("expected image/png, got %q", resp.GetContentType())
	}
}
