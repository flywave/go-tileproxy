package request

import (
	"net/http"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestArcGISRequest(t *testing.T) {
	param := http.Header{
		"layers": []string{"foo"},
	}
	req := NewArcGISRequest(param, "/MapServer/export?map=foo")

	rp := req.GetRequestParams()

	bbox := vec2d.Rect{Min: vec2d.T{-200000, -200000}, Max: vec2d.T{200000, 200000}}

	rp.SetBBox(bbox)

	bbox1 := rp.GetBBox()

	if bbox.Max[0] != bbox1.Max[0] {
		t.FailNow()
	}

	rp.SetBBoxSrs("EPSG:900913")

	srscode := rp.GetBBOxSrs()

	if srscode != "EPSG:900913" {
		t.FailNow()
	}

	rp.SetFormat("png")

	if rp.GetFormat() != "png" {
		t.FailNow()
	}

	rp.SetSize([2]uint32{512, 512})

	if rp.GetSize() != [2]uint32{512, 512} {
		t.FailNow()
	}

	ls := rp.GetLayers()

	if len(ls) != 1 && ls[0] != "foo" {
		t.FailNow()
	}

	rp.AddLayer("bar")

	ls = rp.GetLayers()

	if len(ls) != 2 && ls[1] != "bar" {
		t.FailNow()
	}

	uri := req.QueryString()

	if uri == "" {
		t.FailNow()
	}
}

func TestArcGISIdentifyRequest(t *testing.T) {
	param := http.Header{
		"layers": []string{"foo"},
	}
	req := NewArcGISIdentifyRequest(param, "/MapServer/export?map=foo")

	rp := req.GetRequestParams()

	rp.SetFormat("png")

	if rp.GetFormat() != "png" {
		t.FailNow()
	}

	bbox := vec2d.Rect{Min: vec2d.T{-200000, -200000}, Max: vec2d.T{200000, 200000}}

	rp.SetBBox(bbox)

	bbox1 := rp.GetBBox()

	if bbox.Max[0] != bbox1.Max[0] {
		t.FailNow()
	}

	rp.SetSize([2]uint32{512, 512})

	if rp.GetSize() != [2]uint32{512, 512} {
		t.FailNow()
	}

	rp.SetPos([2]float64{512, 512})

	if rp.GetPos() != [2]float64{512, 512} {
		t.FailNow()
	}

	rp.SetSrs("EPSG:900913")

	srscode := rp.GetSrs()

	if srscode != "EPSG:900913" {
		t.FailNow()
	}

	rp.SetTransparent(true)

	if !rp.GetTransparent() {
		t.FailNow()
	}

	ls := rp.GetLayers()

	if len(ls) != 1 && ls[0] != "foo" {
		t.FailNow()
	}

	rp.AddLayer("bar")

	ls = rp.GetLayers()

	if len(ls) != 2 && ls[1] != "bar" {
		t.FailNow()
	}

	uri := req.QueryString()

	if uri == "" {
		t.FailNow()
	}
}
