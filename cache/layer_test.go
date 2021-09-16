package cache

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

type mockImageSourceCreater struct {
	imageopts *imagery.ImageOptions
}

func (c *mockImageSourceCreater) GetExtension() string {
	return "png"
}

func (c *mockImageSourceCreater) CreateEmpty(size [2]uint32, opts tile.TileOptions) tile.Source {
	return nil
}

func (c *mockImageSourceCreater) Create(data []byte, tile [3]int) tile.Source {
	s := imagery.CreateImageSourceFromBufer(data, c.imageopts)
	return s
}

func TestCacheMapLayer(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := &mockImageSourceCreater{imageopts: imageopts}

	c := NewLocalCache("./test_cache", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, ctx)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &DummyTileLocker{}

	manager := NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, true, false, nil, -1, false, 10, [2]uint32{2, 2})

	cachelayer := NewCacheMapLayer(manager, nil, imageopts, nil)

	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Size: [2]uint32{300, 150}, Srs: geo.NewProj(4326), Format: tile.TileFormat("png")}

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
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := &mockImageSourceCreater{imageopts: imageopts}

	c := NewLocalCache("./test_cache", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, ctx)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &DummyTileLocker{}

	manager := NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, true, false, nil, -1, false, 10, [2]uint32{2, 2})

	cachelayer := NewCacheMapLayer(manager, nil, imageopts, nil)

	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-20037508.34, -20037508.34}, Max: vec2d.T{20037508.34, 20037508.34}}, Size: [2]uint32{500, 500}, Srs: geo.NewProj(900913), Format: tile.TileFormat("png")}

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
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	ccreater := &mockImageSourceCreater{imageopts: imageopts}

	c := NewLocalCache("./test_cache", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, ctx)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &DummyTileLocker{}

	manager := NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, -1, false, 0, [2]uint32{1, 1})

	cachelayer := NewCacheMapLayer(manager, nil, imageopts, nil)
	cov := geo.NewBBoxCoverage(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{90, 45}}, geo.NewProj(4326), false)
	cachelayer.Extent = cov.GetExtent()
	query := &layer.MapQuery{BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, Size: [2]uint32{300, 150}, Srs: geo.NewProj(4326), Format: tile.TileFormat("png")}

	resp, err := cachelayer.GetMap(query)

	if err != nil || resp.GetBuffer(nil, nil) == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}
