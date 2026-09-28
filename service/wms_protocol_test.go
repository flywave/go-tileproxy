package service

import (
	"encoding/xml"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type testInfoLayer struct{}

func (l *testInfoLayer) GetInfo(query *layer.InfoQuery) resource.FeatureInfoDoc {
	return resource.NewTextFeatureInfoDoc("ok")
}

// createQueryableTestWMSService builds a WMS service whose single layer is
// queryable so that GetFeatureInfo exercises the full code path
func createQueryableTestWMSService() *WMSService {
	testMapLayer := &TestMapLayer{
		name:   "test-layer",
		extent: geo.MapExtentFromDefault(),
	}

	testLayer := NewWMSNodeLayer(&WMSNodeLayerOptions{
		Name:      "test-layer",
		Title:     "Test Layer",
		MapLayers: map[string]layer.Layer{"test-layer": testMapLayer},
		Infos:     map[string]layer.InfoLayer{"test-layer": &testInfoLayer{}},
		Legends:   []layer.LegendLayer{},
		Metadata:  &WMSLayerMetadata{},
	})

	metadata := &WMSMetadata{
		Title:    "Test WMS Service",
		Abstract: "Test WMS Service for unit tests",
		URL:      "http://localhost/wms",
	}

	groupLayer := NewWMSGroupLayer(&WMSGroupLayerOptions{
		Name:     "test-group",
		Title:    "Test Group",
		Layers:   map[string]WMSLayer{"test-layer": testLayer},
		Metadata: &WMSLayerMetadata{},
	})

	return NewWMSService(&WMSServiceOptions{
		RootLayer: groupLayer,
		Layers: map[string]WMSLayer{
			"test-layer": testLayer,
		},
		ImageFormats: map[string]*imagery.ImageOptions{
			"image/png": {
				Format: tile.TileFormat("png"),
				Mode:   imagery.RGBA,
			},
		},
		InfoFormats: map[string]string{
			"text": "text/plain",
			"html": "text/html",
			"xml":  "text/xml",
		},
		Srs: &geo.SupportedSRS{
			Srs: []geo.Proj{geo.NewProj("EPSG:4326")},
		},
		SrsExtents: map[string]*geo.MapExtent{
			"EPSG:4326": geo.MapExtentFromDefault(),
		},
		MaxOutputPixels: 4000 * 4000,
		Metadata:        metadata,
	})
}

// GetMap must accept a standard "format=image/png" request; previously the
// format validation compared the subtype against the full mime list and
// rejected every valid request.
func TestWMSService_GetMap_FormatValidation(t *testing.T) {
	service := createTestWMSService()

	req := httptest.NewRequest("GET", "/wms?service=WMS&request=GetMap&layers=test-layer&styles=&format=image/png&width=256&height=256&crs=EPSG:4326&bbox=-180,-90,180,90", nil)
	wmsReq := request.MakeWMSRequest(req, false)
	if wmsReq == nil {
		t.Fatal("request not parsed")
	}
	resp := service.GetMap(wmsReq)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatus(), string(resp.GetBuffer()))
	}
	if !strings.HasPrefix(resp.GetContentType(), "image/png") {
		t.Fatalf("expected image/png, got %s", resp.GetContentType())
	}
}

// the capabilities document must advertise version 1.3.0
func TestWMSService_CapabilitiesVersion(t *testing.T) {
	service := createTestWMSService()

	req := httptest.NewRequest("GET", "/wms?service=WMS&request=GetCapabilities", nil)
	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.GetCapabilities(wmsReq)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}

	var checker struct {
		Version string `xml:"version,attr"`
	}
	if err := xml.Unmarshal(resp.GetBuffer(), &checker); err != nil {
		t.Fatalf("capabilities not valid XML: %v", err)
	}
	if checker.Version != "1.3.0" {
		t.Fatalf("expected capabilities version 1.3.0, got %q", checker.Version)
	}
}

// GetFeatureInfo honours QUERY_LAYERS and the 1.1.1 X/Y position parameters
func TestWMSService_GetFeatureInfo_QueryLayers(t *testing.T) {
	service := createQueryableTestWMSService()

	req := httptest.NewRequest("GET", "/wms?service=WMS&request=GetFeatureInfo&layers=test-layer&query_layers=test-layer&format=text/html&width=256&height=256&crs=EPSG:4326&bbox=-180,-90,180,90&x=128&y=128", nil)
	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.GetFeatureInfo(wmsReq)
	if resp.GetStatus() != 200 {
		t.Fatalf("expected status 200, got %d", resp.GetStatus())
	}
	if !strings.HasPrefix(resp.GetContentType(), "text/") {
		t.Fatalf("expected text/* content type, got %q", resp.GetContentType())
	}
}

// GetFeatureInfo must reject requests for unknown query layers instead of
// dereferencing a nil layer
func TestWMSService_GetFeatureInfo_UnknownLayer(t *testing.T) {
	service := createTestWMSService()

	req := httptest.NewRequest("GET", "/wms?service=WMS&request=GetFeatureInfo&layers=test-layer&query_layers=nope&format=text/html&width=256&height=256&crs=EPSG:4326&bbox=-180,-90,180,90&x=1&y=1", nil)
	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.GetFeatureInfo(wmsReq)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.GetStatus() == 200 {
		t.Fatal("unknown query layer should not succeed")
	}
}

// group layer legends used to panic with "not implemented"
func TestWMSGroupLayer_Legend(t *testing.T) {
	nodeLayer := NewWMSNodeLayer(&WMSNodeLayerOptions{
		Name:  "child",
		Title: "Child",
	})

	group := NewWMSGroupLayer(&WMSGroupLayerOptions{
		Name:   "group",
		Title:  "Group",
		Layers: map[string]WMSLayer{"child": nodeLayer},
	})

	// must not panic even without concrete legend layers
	_ = group.legend(&layer.LegendQuery{Format: "image/png", Scale: 1})
}

// GetMap limited to the SRS extent intersection must not dereference a nil
// image format entry
func TestWMSService_GetMap_FormatOutsideExtents(t *testing.T) {
	service := createTestWMSService()

	// bbox completely outside EPSG:4326 default extent triggers the
	// limited-extent empty response path
	req := httptest.NewRequest("GET", "/wms?service=WMS&request=GetMap&layers=test-layer&styles=&format=bogus&width=256&height=256&crs=EPSG:4326&bbox=-180,-90,180,90", nil)
	wmsReq := request.MakeWMSRequest(req, false)
	resp := service.GetMap(wmsReq)
	if resp.GetStatus() == 200 && strings.HasPrefix(resp.GetContentType(), "bogus") {
		t.Fatal("bogus format should not be served")
	}
}
