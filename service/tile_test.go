package service

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type mockClient struct {
	client.HttpClient
	data []byte
	url  []string
	body []byte
	code int
}

func (c *mockClient) Open(url string, data []byte, hdr http.Header) (statusCode int, body []byte) {
	c.data = data
	c.url = append(c.url, url)
	return c.code, c.body
}

func create_cached_tile(coord [3]int, data []byte, cache_ *cache.LocalCache, timestamp *time.Time) {
	loc := cache_.TileLocation(cache.NewTile(coord), true)
	if f, err := os.Create(loc); err != nil {
		f.Write(data)
		f.Close()
	}

	if timestamp != nil {
		os.Chtimes(loc, *timestamp, *timestamp)
	}
}

type mockContext struct {
	client.Context
	c *mockClient
}

func (c *mockContext) Client() client.HttpClient {
	return c.c
}

func (c *mockContext) Sync() {
}

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

func TestTileProvider(t *testing.T) {
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

	c := cache.NewLocalCache("./test_cache", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, nil, nil, ctx)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &cache.DummyTileLocker{}

	topts := &cache.TileManagerOptions{
		Sources:              []layer.Layer{source},
		Grid:                 grid,
		Cache:                c,
		Locker:               locker,
		Identifier:           "test",
		Format:               "png",
		Options:              imageopts,
		MinimizeMetaRequests: false,
		BulkMetaTiles:        false,
		PreStoreFilter:       nil,
		RescaleTiles:         -1,
		CacheRescaledTiles:   false,
		MetaBuffer:           0,
		MetaSize:             [2]uint32{1, 1},
	}

	manager := cache.NewTileManager(topts)

	md := &TileProviderMetadata{}

	info := []layer.InfoLayer{}

	dimensions := make(utils.Dimensions)

	tpopts := &TileProviderOptions{Name: "test", Title: "test", Metadata: md, TileManager: manager, InfoSources: info, Dimensions: dimensions, ErrorHandler: &TMSExceptionHandler{}}

	tp := NewTileProvider(tpopts)

	if tp == nil {
		t.FailNow()
	}

	hreq := &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0/landsat2000/16/53958/24829.png")

	tileReq := request.NewTileRequest(hreq)

	err, resp := tp.Render(tileReq, false, nil, nil)

	if resp == nil || err != nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestTileServiceGetMap(t *testing.T) {
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

	c := cache.NewLocalCache("./test_cache", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, nil, nil, ctx)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &cache.DummyTileLocker{}

	topts := &cache.TileManagerOptions{
		Sources:              []layer.Layer{source},
		Grid:                 grid,
		Cache:                c,
		Locker:               locker,
		Identifier:           "test",
		Format:               "png",
		Options:              imageopts,
		MinimizeMetaRequests: false,
		BulkMetaTiles:        false,
		PreStoreFilter:       nil,
		RescaleTiles:         -1,
		CacheRescaledTiles:   false,
		MetaBuffer:           0,
		MetaSize:             [2]uint32{1, 1},
	}

	manager := cache.NewTileManager(topts)

	md := &TileProviderMetadata{Name: "test"}

	info := []layer.InfoLayer{}

	dimensions := make(utils.Dimensions)

	tpopts := &TileProviderOptions{Name: "test", Title: "test", Metadata: md, TileManager: manager, InfoSources: info, Dimensions: dimensions, ErrorHandler: &TMSExceptionHandler{}}

	tp := NewTileProvider(tpopts)

	if tp == nil {
		t.FailNow()
	}

	layers := map[string]Provider{"landsat2000": tp}
	metadata := &TileMetadata{}

	tsopts := &TileServiceOptions{Layers: layers, Metadata: metadata, MaxTileAge: nil, UseDimensionLayers: false, Origin: "ul"}

	service := NewTileService(tsopts)

	hreq := &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0/landsat2000/16/53958/24829.png")

	tileReq := request.NewTileRequest(hreq)

	resp := service.GetMap(tileReq)

	if resp == nil {
		t.FailNow()
	}

	hreq = &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0/landsat2000")

	tmsReq := request.NewTileRequest(hreq)

	resp = service.GetCapabilities(tmsReq)

	if resp == nil {
		t.FailNow()
	}

	hreq = &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0")

	tmsReq = request.NewTileRequest(hreq)

	resp = service.GetCapabilities(tmsReq)

	if resp == nil {
		t.FailNow()
	}

	hreq = &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tms/")

	tmsReq = request.NewTileRequest(hreq)

	resp = service.RootResource(tmsReq)

	if resp == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}
