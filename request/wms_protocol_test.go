package request

import (
	"net/http/httptest"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

// WMS 1.3.0 clients pass the CRS via "crs" and send EPSG:4326 bboxes in
// northing/easting order; the parsed request must normalize to lon/lat.
func TestWMSRequestAxisOrder130(t *testing.T) {
	req := httptest.NewRequest("GET", "/wms?service=WMS&version=1.3.0&request=GetMap&layers=foo&format=image/png&width=256&height=256&crs=EPSG:4326&bbox=-90,-180,90,180", nil)
	r := MakeWMSRequest(req, false)
	if r == nil {
		t.Fatal("request not parsed")
	}
	params := NewWMSMapRequestParams(r.GetParams())
	bbox := params.GetBBox()
	if bbox.Min[0] != -180 || bbox.Min[1] != -90 || bbox.Max[0] != 180 || bbox.Max[1] != 90 {
		t.Fatalf("bbox not normalized to lon/lat: %v", bbox)
	}
	if params.GetCrs() != "EPSG:4326" {
		t.Fatalf("unexpected crs %q", params.GetCrs())
	}
}

// WMS 1.1.1 clients pass the SRS via "srs" and send lon/lat bboxes that
// must not be touched.
func TestWMSRequestAxisOrder111(t *testing.T) {
	req := httptest.NewRequest("GET", "/wms?service=WMS&version=1.1.1&request=GetMap&layers=foo&format=image/png&width=256&height=256&srs=EPSG:4326&bbox=-180,-90,180,90", nil)
	r := MakeWMSRequest(req, false)
	if r == nil {
		t.Fatal("request not parsed")
	}
	params := NewWMSMapRequestParams(r.GetParams())
	bbox := params.GetBBox()
	if bbox.Min[0] != -180 || bbox.Min[1] != -90 || bbox.Max[0] != 180 || bbox.Max[1] != 90 {
		t.Fatalf("1.1.1 bbox corrupted: %v", bbox)
	}
	if params.GetCrs() != "EPSG:4326" {
		t.Fatalf("GetCrs should fall back to srs, got %q", params.GetCrs())
	}
}

// comma separated LAYERS lists must be split into individual layer names
func TestWMSRequestCommaLayers(t *testing.T) {
	req := httptest.NewRequest("GET", "/wms?request=GetMap&layers=foo,bar,baz&format=image/png", nil)
	r := MakeWMSRequest(req, false)
	params := NewWMSMapRequestParams(r.GetParams())
	layers := params.GetLayers()
	if len(layers) != 3 || layers[0] != "foo" || layers[1] != "bar" || layers[2] != "baz" {
		t.Fatalf("layers not split: %v", layers)
	}
}

// GetFeatureInfo must read the position from I/J and fall back to X/Y
func TestWMSFeatureInfoPos(t *testing.T) {
	req := httptest.NewRequest("GET", "/wms?request=GetFeatureInfo&layers=foo&query_layers=foo&i=10&j=20&format=text/html", nil)
	r := MakeWMSRequest(req, false)
	params := NewWMSFeatureInfoRequestParams(r.GetParams())
	if pos := params.GetPos(); pos != [2]float64{10, 20} {
		t.Fatalf("I/J position not read: %v", pos)
	}

	req = httptest.NewRequest("GET", "/wms?request=GetFeatureInfo&layers=foo&query_layers=foo&x=30&y=40&format=text/html", nil)
	r = MakeWMSRequest(req, false)
	params = NewWMSFeatureInfoRequestParams(r.GetParams())
	if pos := params.GetPos(); pos != [2]float64{30, 40} {
		t.Fatalf("X/Y fallback not used: %v", pos)
	}

	ql := params.GetQueryLayers()
	if len(ql) != 1 || ql[0] != "foo" {
		t.Fatalf("query_layers not read: %v", ql)
	}
}

// format validation accepts full mime types and bare subtypes
func TestWMSValidateFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/wms?request=GetMap&layers=foo&format=image/png", nil)
	r := MakeWMSRequest(req, false)
	if err := r.(*WMSMapRequest).ValidateFormat([]string{"image/png", "image/jpeg"}); err != nil {
		t.Fatalf("full mime type rejected: %v", err)
	}

	req = httptest.NewRequest("GET", "/wms?request=GetMap&layers=foo&format=png", nil)
	r = MakeWMSRequest(req, false)
	if err := r.(*WMSMapRequest).ValidateFormat([]string{"image/png", "image/jpeg"}); err != nil {
		t.Fatalf("bare subtype rejected: %v", err)
	}

	req = httptest.NewRequest("GET", "/wms?request=GetMap&layers=foo&format=image/gif", nil)
	r = MakeWMSRequest(req, false)
	if err := r.(*WMSMapRequest).ValidateFormat([]string{"image/png", "image/jpeg"}); err == nil {
		t.Fatal("unsupported format not rejected")
	}
}

// SwitchBBoxEpsgAxisOrder is an involution for NE-axis SRS and a no-op
// for EN-axis SRS
func TestSwitchBBoxEpsgAxisOrder(t *testing.T) {
	rect := vec2d.Rect{Min: vec2d.T{-90, -180}, Max: vec2d.T{90, 180}}
	swapped := SwitchBBoxEpsgAxisOrder(rect, "EPSG:4326")
	if swapped.Min[0] != -180 || swapped.Min[1] != -90 {
		t.Fatalf("4326 bbox not swapped: %v", swapped)
	}
	if twice := SwitchBBoxEpsgAxisOrder(swapped, "EPSG:4326"); twice != rect {
		t.Fatalf("swap not invertible: %v", twice)
	}
	en := SwitchBBoxEpsgAxisOrder(rect, "EPSG:3857")
	if en != rect {
		t.Fatalf("3857 bbox should be untouched: %v", en)
	}
}
