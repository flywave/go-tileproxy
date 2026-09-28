package service

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/request"
)

func createTestWMTSGrid() *geo.TileGrid {
	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_BBOX] = []float64{-180, -90, 180, 90}
	conf[geo.TILEGRID_BBOX_SRS] = geo.NewProj("EPSG:4326")
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_NUM_LEVELS] = 3

	grid := geo.NewTileGrid(conf)
	grid.Name = "test_grid"
	return grid
}

func createTestWMTSService() *WMTSService {
	grid := createTestWMTSGrid()
	mockProvider := createTestTileProvider("test_layer", grid)

	return NewWMTSService(&WMTSServiceOptions{
		Layers: map[string]Provider{
			"test_layer": mockProvider,
		},
		Metadata: &WMTSMetadata{
			Title: "Test WMTS Service",
			URL:   "http://localhost/wmts",
		},
		MaxTileAge: &[]time.Duration{time.Hour}[0],
		InfoFormats: map[string]string{
			"xml": "text/xml",
		},
	})
}

func createTestWMTSRestService() *WMTSRestService {
	grid := createTestWMTSGrid()
	mockProvider := createTestTileProvider("test_layer", grid)

	return NewWMTSRestService(&WMTSRestServiceOptions{
		Layers: map[string]Provider{
			"test_layer": mockProvider,
		},
		Metadata: &WMTSMetadata{
			Title: "Test WMTS Rest Service",
			URL:   "http://localhost/wmts",
		},
		MaxTileAge: &[]time.Duration{time.Hour}[0],
		InfoFormats: map[string]string{
			"xml": "text/xml",
		},
	})
}

// RESTful tile URLs advertised in the capabilities document must actually
// resolve to tiles
func TestWMTSRestService_ParseTileURL(t *testing.T) {
	s := createTestWMTSRestService()

	req := httptest.NewRequest("GET", "/wmts/test_layer/test_grid/0/0/0.png", nil)
	r := s.RequestParser(req)
	if r == nil {
		t.Fatal("REST tile URL not parsed")
	}
	if r.GetRequestHandler() != "tile" {
		t.Fatalf("expected tile handler, got %q", r.GetRequestHandler())
	}

	resp := s.GetTile(r)
	if resp == nil {
		t.Fatal("GetTile returned nil")
	}
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}
	if !strings.HasPrefix(resp.GetContentType(), "image/png") {
		t.Fatalf("expected image/png content type, got %q", resp.GetContentType())
	}
}

// RESTful GetCapabilities URL must resolve
func TestWMTSRestService_ParseCapabilitiesURL(t *testing.T) {
	s := createTestWMTSRestService()

	req := httptest.NewRequest("GET", "/wmts/1.0.0/WMTSCapabilities.xml", nil)
	r := s.RequestParser(req)
	if r == nil {
		t.Fatal("REST capabilities URL not parsed")
	}
	if r.GetRequestHandler() != "capabilities" {
		t.Fatalf("expected capabilities handler, got %q", r.GetRequestHandler())
	}

	resp := s.GetCapabilities(r)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}
	body := string(resp.GetBuffer())
	if !strings.Contains(body, "ResourceURL") {
		t.Fatal("REST capabilities should advertise ResourceURL templates")
	}
	if !strings.Contains(body, "/test_layer/{TileMatrixSet}/{TileMatrix}/{TileCol}/{TileRow}.png") {
		t.Fatalf("capabilities missing tile template, got: %s", body)
	}
}

// KVP requests keep working on a REST service and must respond with the
// image content type instead of text/plain
func TestWMTSRestService_KVPFallback(t *testing.T) {
	s := createTestWMTSRestService()

	req := httptest.NewRequest("GET", "/wmts?service=WMTS&request=GetTile&version=1.0.0&layer=test_layer&style=default&tilematrixset=test_grid&tilematrix=0&tilerow=0&tilecol=0&format=image/png", nil)
	r := s.RequestParser(req)
	if r == nil {
		t.Fatal("KVP request not parsed")
	}

	resp := s.GetTile(r)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}
	if !strings.HasPrefix(resp.GetContentType(), "image/png") {
		t.Fatalf("expected image/png content type, got %q", resp.GetContentType())
	}
}

// the KVP-only WMTS service must return the image content type for tiles
func TestWMTSService_GetTile_ContentType(t *testing.T) {
	s := createTestWMTSService()

	req := httptest.NewRequest("GET", "/wmts?service=WMTS&request=GetTile&version=1.0.0&layer=test_layer&style=default&tilematrixset=test_grid&tilematrix=0&tilerow=0&tilecol=0&format=image/png", nil)
	r := request.MakeWMTSRequest(req, false)

	resp := s.GetTile(r)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}
	if !strings.HasPrefix(resp.GetContentType(), "image/png") {
		t.Fatalf("expected image/png content type, got %q", resp.GetContentType())
	}
}

// the template converter must match REST URLs built from the default
// template and extract all variables
func TestURLTemplateConverter_Match(t *testing.T) {
	conv := request.NewURLTemplateConverter(DEFAULT_WMTS_TEMPLATE)

	m := conv.Match("/wmts/my_layer/my_tms/3/2/1.png")
	if m == nil {
		t.Fatal("template did not match tile URL")
	}
	if m["layer"] != "my_layer" || m["tilematrixset"] != "my_tms" ||
		m["tilematrix"] != "3" || m["tilecol"] != "2" || m["tilerow"] != "1" || m["format"] != "png" {
		t.Fatalf("unexpected template variables: %v", m)
	}

	if conv.Match("/wmts/1.0.0/WMTSCapabilities.xml") != nil {
		t.Fatal("capabilities URL must not match the tile template")
	}

	infoConv := request.NewFeatureInfoURLTemplateConverter(DEFAULT_WMTS_INFO_TEMPLATE)
	mi := infoConv.Match("/wmts/my_layer/my_tms/3/2/1/10/20.xml")
	if mi == nil {
		t.Fatal("info template did not match feature info URL")
	}
	if mi["i"] != "10" || mi["j"] != "20" || mi["infoformat"] != "xml" {
		t.Fatalf("unexpected info template variables: %v", mi)
	}
}
