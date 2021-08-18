package layer

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

type mockSource struct {
	MapLayer
	requested bool
}

func (s *mockSource) GetMap(query *MapQuery) (tile.Source, error) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	s.requested = true

	return imagery.CreateImageSourceFromBufer(imagedata.Bytes(), imageopts), nil
}

var (
	GLOBAL_GEOGRAPHIC_EXTENT = &geo.MapExtent{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Srs: geo.NewSRSProj4("EPSG:4326")}
)

func TestResolutionConditional(t *testing.T) {
	type QueryInfo struct {
		key           string
		query         *MapQuery
		low_requested bool
	}
	testQuery := []QueryInfo{
		{"low", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10000, 10000}}, Size: [2]uint32{100, 100}, Srs: geo.NewSRSProj4("EPSG:3857")}, true},
		{"high", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100, 100}}, Size: [2]uint32{100, 100}, Srs: geo.NewSRSProj4("EPSG:3857")}, false},
		{"match", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}}, Size: [2]uint32{100, 100}, Srs: geo.NewSRSProj4("EPSG:3857")}, false},
		{"low_transform", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0.1, 0.1}}, Size: [2]uint32{100, 100}, Srs: geo.NewSRSProj4("EPSG:4326")}, true},
		{"high_transform", &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0.005, 0.005}}, Size: [2]uint32{100, 100}, Srs: geo.NewSRSProj4("EPSG:4326")}, false},
	}

	for i := range testQuery {
		_, map_query, low_requested := testQuery[i].key, testQuery[i].query, testQuery[i].low_requested
		low := &mockSource{}
		high := &mockSource{}
		layer := NewResolutionConditional(low, high, 10, geo.NewSRSProj4("EPSG:3857"), GLOBAL_GEOGRAPHIC_EXTENT, nil)

		layer.GetMap(map_query)
		if low.requested != low_requested {
			t.FailNow()
		}
		if high.requested == low_requested {
			t.FailNow()
		}
	}
}

func TestSRSConditional(t *testing.T) {
	l4326 := &mockSource{}
	l3857 := &mockSource{}
	l25832 := &mockSource{}
	preferred := geo.PreferredSrcSRS{}
	preferred.Add("EPSG:31467", []geo.Proj{geo.NewSRSProj4("EPSG:25832"), geo.NewSRSProj4("EPSG:3857")})
	layer := NewSRSConditional(map[string]Layer{
		"EPSG:4326":  l4326,
		"EPSG:3857":  l3857,
		"EPSG:25832": l25832,
	}, GLOBAL_GEOGRAPHIC_EXTENT, nil, preferred)

	if layer.selectLayer(geo.NewSRSProj4("EPSG:4326")) != l4326 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewSRSProj4("EPSG:3857")) != l3857 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewSRSProj4("EPSG:25832")) != l25832 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewSRSProj4("EPSG:31466")) != l3857 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewSRSProj4("EPSG:32633")) != l3857 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewSRSProj4("EPSG:4258")) != l4326 {
		t.FailNow()
	}
	if layer.selectLayer(geo.NewSRSProj4("EPSG:31467")) != l25832 {
		t.FailNow()
	}
}

type requestInfo struct {
	bbox vec2d.Rect
	size [2]uint32
	srs  string
}

type mockRequestSource struct {
	MapLayer
	requested []requestInfo
}

func (s *mockRequestSource) GetMap(query *MapQuery) (tile.Source, error) {
	rgba := image.NewRGBA(image.Rect(0, 0, int(query.Size[0]), int(query.Size[1])))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	s.requested = append(s.requested, requestInfo{bbox: query.BBox, size: query.Size, srs: query.Srs.GetSrsCode()})

	return imagery.CreateImageSourceFromBufer(imagedata.Bytes(), imageopts), nil
}

func TestDirectMapLayer(t *testing.T) {
	source := &mockRequestSource{}
	cl := NewDirectMapLayer(source, GLOBAL_GEOGRAPHIC_EXTENT)

	query := &MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Size: [2]uint32{300, 150}, Srs: geo.NewSRSProj4("EPSG:4326"), Format: tile.TileFormat("png")}

	resp, err := cl.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}
}
