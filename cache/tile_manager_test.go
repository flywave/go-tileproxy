package cache

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"os"
	"testing"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

type mockClient struct {
	client.HttpClient
	data []byte
	url  []string
	body []byte
	code int
}

func (c *mockClient) Open(url string, data []byte) (statusCode int, body []byte) {
	c.data = data
	c.url = append(c.url, url)
	return c.code, c.body
}

func create_cached_tile(tile [3]int, data []byte, cache *LocalCache, timestamp *time.Time) {
	loc := cache.tile_location(NewTile(tile), true)
	if f, err := os.Create(loc); err != nil {
		f.Write(data)
		f.Close()
	}

	if timestamp != nil {
		os.Chtimes(loc, *timestamp, *timestamp)
	}
}

func TestTileManager(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
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

	manager := NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, 0, false, 0, [2]uint32{2, 2})

	if manager.IsStale([3]int{0, 0, 1}, nil) {
		t.FailNow()
	}

	create_cached_tile([3]int{0, 0, 1}, imagedata.Bytes(), c, nil)

	if manager.IsStale([3]int{0, 0, 1}, nil) {
		t.FailNow()
	}

	exp := time.Now().Add(time.Duration(3600))

	create_cached_tile([3]int{0, 0, 1}, imagedata.Bytes(), c, &exp)
	now := time.Now()
	manager.expireTimestamp = &now

	if !manager.IsStale([3]int{0, 0, 1}, nil) {
		t.FailNow()
	}

	manager.RemoveTileCoords([][3]int{{0, 0, 0}, {0, 0, 1}})

	if manager.IsCached([3]int{0, 0, 1}, nil) {
		t.FailNow()
	}

	err, tiles := manager.LoadTileCoords([][3]int{{0, 0, 2}, {2, 0, 2}}, nil, false)

	if tiles == nil || err != nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestTileManagerMinimalMetaRequests(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
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

	err, tiles := manager.LoadTileCoords([][3]int{{0, 0, 2}}, nil, false)

	if tiles == nil || err != nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

type requestInfo struct {
	bbox vec2d.Rect
	size [2]uint32
	srs  string
}

type mockSource struct {
	layer.MapLayer
	requested []requestInfo
}

func (s *mockSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	rgba := image.NewRGBA(image.Rect(0, 0, int(query.Size[0]), int(query.Size[1])))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	s.requested = append(s.requested, requestInfo{bbox: query.BBox, size: query.Size, srs: query.Srs.GetSrsCode()})

	return imagery.CreateImageSourceFromBufer(imagedata.Bytes(), imageopts), nil
}

func TestTileManagerMultipleSources(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	locker := &DummyTileLocker{}

	source_base := &mockSource{}
	source_overlay := &mockSource{}

	manager := NewTileManager([]layer.Layer{source_base, source_overlay}, grid, c, locker, "test", "png", imageopts, true, false, nil, -1, false, 10, [2]uint32{2, 2})

	err, tiles := manager.LoadTileCoords([][3]int{{0, 0, 2}}, nil, false)

	if tiles == nil || err != nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestTileManagerMultipleSourcesWithMetaTiles(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	locker := &DummyTileLocker{}

	source_base := &mockSource{MapLayer: layer.MapLayer{SupportMetaTiles: true}}
	source_overlay := &mockSource{MapLayer: layer.MapLayer{SupportMetaTiles: true}}

	manager := NewTileManager([]layer.Layer{source_base, source_overlay}, grid, c, locker, "test", "png", imageopts, true, false, nil, -1, false, 0, [2]uint32{2, 2})

	err, tiles := manager.LoadTileCoords([][3]int{{0, 0, 1}, {1, 0, 1}}, nil, false)

	if tiles == nil || err != nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestTileManagerBulkMetaTiles(t *testing.T) {
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png")}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	locker := &DummyTileLocker{}

	source_base := &mockSource{MapLayer: layer.MapLayer{SupportMetaTiles: false}}
	source_overlay := &mockSource{MapLayer: layer.MapLayer{SupportMetaTiles: true}}

	manager := NewTileManager([]layer.Layer{source_base, source_overlay}, grid, c, locker, "test", "png", imageopts, false, true, nil, -1, false, 0, [2]uint32{2, 2})

	err, tiles := manager.LoadTileCoords([][3]int{{1, 0, 2}, {2, 0, 2}}, nil, false)

	if tiles == nil || err != nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}
