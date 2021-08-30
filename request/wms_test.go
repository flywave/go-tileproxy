package request

import (
	"net/http"
	"testing"

	"github.com/flywave/go-tileproxy/utils"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestWMSMapRequest(t *testing.T) {
	param := http.Header{
		"layers": []string{"foo"},
	}
	req := NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	rp := req.GetRequestParams()

	bbox := vec2d.Rect{Min: vec2d.T{-200000, -200000}, Max: vec2d.T{200000, 200000}}

	rp.SetBBox(bbox)

	bbox1 := rp.GetBBox()

	if bbox.Max[0] != bbox1.Max[0] {
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

	rp.SetSize([2]uint32{512, 512})

	if rp.GetSize() != [2]uint32{512, 512} {
		t.FailNow()
	}

	rp.SetSrs("EPSG:900913")

	srscode := rp.GetSrs()

	if srscode != "EPSG:900913" {
		t.FailNow()
	}

	rp.SetFormat("png")

	if rp.GetFormatMimeType() != "image/png" {
		t.FailNow()
	}

	rp.SetTransparent(true)

	if !rp.GetTransparent() {
		t.FailNow()
	}

	c := utils.HexColor("#FFC0CB")

	rp.SetBGColor(c)

	r1, g1, b1, a1 := c.RGBA()
	r, g, b, a := rp.GetBGColor().RGBA()

	if r1 != r || g1 != g || b1 != b || a1 != a {
		t.FailNow()
	}

	uri := req.CompleteUrl()

	if uri == "" {
		t.FailNow()
	}
}

func TestWMSLegendGraphicRequest(t *testing.T) {
	param := http.Header{
		"layer": []string{"foo"},
	}
	req := NewWMSLegendGraphicRequest(param, "/service?map=foo", false, nil, false)
	rp := req.GetRequestParams()

	ls := rp.GetLayer()

	if ls != "foo" {
		t.FailNow()
	}

	rp.SetScale(2)

	if rp.GetScale() != 2 {
		t.FailNow()
	}

	rp.SetFormat("png")

	if rp.GetFormatMimeType() != "image/png" {
		t.FailNow()
	}

	uri := req.CompleteUrl()

	if uri == "" {
		t.FailNow()
	}
}

func TestWMSFeatureInfoRequest(t *testing.T) {
	param := http.Header{
		"layers": []string{"foo"},
	}
	req := NewWMSFeatureInfoRequest(param, "/service?map=foo", false, nil, false)
	rp := req.GetRequestParams()

	ls := rp.GetLayers()

	if len(ls) != 1 && ls[0] != "foo" {
		t.FailNow()
	}

	rp.AddLayer("bar")

	ls = rp.GetLayers()

	if len(ls) != 2 && ls[1] != "bar" {
		t.FailNow()
	}

	rp.SetPos([2]float64{512, 512})

	if rp.GetPos() != [2]float64{512, 512} {
		t.FailNow()
	}

	uri := req.CompleteUrl()

	if uri == "" {
		t.FailNow()
	}
}

func TestWMSCapabilitiesRequest(t *testing.T) {
	param := http.Header{
		"layers": []string{"foo"},
	}
	req := NewWMSCapabilitiesRequest(param, "/service?map=foo", false, nil)
	rp := req.GetRequestParams()

	rp.SetFormat("png")

	if rp.GetFormatMimeType() != "image/png" {
		t.FailNow()
	}

	uri := req.CompleteUrl()

	if uri == "" {
		t.FailNow()
	}
}
