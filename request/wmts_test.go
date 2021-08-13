package request

import (
	"net/http"
	"testing"
)

func TestWMTS100TileRequest(t *testing.T) {
	param := http.Header{
		"layer": []string{"foo"},
	}
	req := NewWMTS100TileRequest(param, "/service?map=foo", false, nil)

	rp := req.GetRequestParams()

	if rp.GetLayer() != "foo" {
		t.FailNow()
	}

	rp.SetCoord([3]int{0, 0, 1})

	if rp.GetCoord() != [3]int{0, 0, 1} {
		t.FailNow()
	}

	rp.SetFormat("png")

	if rp.GetFormatMimeType() != "image/png" {
		t.FailNow()
	}

	rp.SetTileMatrixSet("EPSG:3857")

	if rp.GetTileMatrixSet() != "EPSG:3857" {
		t.FailNow()
	}

	uri := req.CompleteUrl()

	if uri == "" {
		t.FailNow()
	}
}

func TestWMTS100FeatureInfoRequest(t *testing.T) {
	param := http.Header{
		"layer": []string{"foo"},
	}
	req := NewWMTS100FeatureInfoRequest(param, "/service?map=foo", false, nil)
	rp := req.GetRequestParams()

	if rp.GetLayer() != "foo" {
		t.FailNow()
	}

	rp.SetCoord([3]int{0, 0, 1})

	if rp.GetCoord() != [3]int{0, 0, 1} {
		t.FailNow()
	}

	rp.SetFormat("png")

	if rp.GetFormatMimeType() != "image/png" {
		t.FailNow()
	}

	rp.SetTileMatrixSet("EPSG:3857")

	if rp.GetTileMatrixSet() != "EPSG:3857" {
		t.FailNow()
	}

	rp.SetPos([2]float64{512, 512})

	if rp.GetPos() != [2]float64{512, 512} {
		t.FailNow()
	}

	rp.SetInfoformat("application/json")

	if rp.GetInfoformat() != "application/json" {
		t.FailNow()
	}

	uri := req.CompleteUrl()

	if uri == "" {
		t.FailNow()
	}

}

func TestWMTS100CapabilitiesRequest(t *testing.T) {
	param := http.Header{}
	req := NewWMTS100CapabilitiesRequest(param, "/service?map=foo", false, nil)
	rp := req.GetRequestParams()

	rp.SetFormat("application/json")

	if rp.GetFormatMimeType() != "application/json" {
		t.FailNow()
	}

	rp.SetUpdateSequence(UpdateSequenceNone)

	if rp.GetUpdateSequence() != UpdateSequenceNone {
		t.FailNow()
	}

	uri := req.CompleteUrl()

	if uri == "" {
		t.FailNow()
	}

}
