package cache

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

func TestCacheMapLayer(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &images.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := images.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, mock)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &DummyTileLocker{}

	manager := NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, true, false, nil, -1, false, 10, [2]uint32{2, 2})

	cachelayer := NewCacheMapLayer(manager, nil, imageopts, nil)

	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Size: [2]uint32{300, 150}, Srs: geo.NewSRSProj4("EPSG:4326"), Format: tile.TileFormat("png")}

	resp, err := cachelayer.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestCacheMapLayerGetLarge(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &images.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := images.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, mock)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &DummyTileLocker{}

	manager := NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, true, false, nil, -1, false, 10, [2]uint32{2, 2})

	cachelayer := NewCacheMapLayer(manager, nil, imageopts, nil)

	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-20037508.34, -20037508.34}, Max: vec2d.T{20037508.34, 20037508.34}}, Size: [2]uint32{500, 500}, Srs: geo.NewSRSProj4("EPSG:900913"), Format: tile.TileFormat("png")}

	resp, err := cachelayer.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestCacheMapLayerWithExtent(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &images.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := images.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, mock)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &DummyTileLocker{}

	manager := NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, -1, false, 0, [2]uint32{1, 1})

	cachelayer := NewCacheMapLayer(manager, nil, imageopts, nil)
	cov := geo.NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{90, 45}}, geo.NewSRSProj4("EPSG:4326"), false)
	cachelayer.Extent = cov.GetExtent()
	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Size: [2]uint32{300, 150}, Srs: geo.NewSRSProj4("EPSG:4326"), Format: tile.TileFormat("png")}

	resp, err := cachelayer.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}
