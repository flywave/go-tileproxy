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

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
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

func (c *mockClient) Open(url string, data []byte) (statusCode int, body []byte) {
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

func TestTileProvider(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := cache.NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, mock)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &cache.DummyTileLocker{}

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, -1, false, 0, [2]uint32{1, 1})

	md := make(map[string]string)

	info := []layer.Layer{}

	dimensions := make(utils.Dimensions)

	tp := NewTileProvider("test", "test", md, manager, info, dimensions)

	if tp == nil {
		t.FailNow()
	}

	hreq := &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0/landsat2000/16/53958/24829.png")

	tileReq := request.NewTileRequest(hreq)

	resp := tp.Render(tileReq, false, nil, nil)

	if resp == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}

func TestTileServiceGetMap(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 256, 256))
	imagedata := &bytes.Buffer{}
	png.Encode(imagedata, rgba)

	mock := &mockClient{code: 200, body: imagedata.Bytes()}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
		return s
	}

	c := cache.NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	param := http.Header{
		"layers": []string{"foo"},
	}
	req := request.NewWMSMapRequest(param, "/service?map=foo", false, nil, false)

	client := client.NewWMSClient(req, mock)

	source := sources.NewWMSSource(client, imageopts, nil, nil, nil, nil, nil, nil, nil)

	locker := &cache.DummyTileLocker{}

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, -1, false, 0, [2]uint32{1, 1})

	md := map[string]string{"name_path": "test"}

	info := []layer.Layer{}

	dimensions := make(utils.Dimensions)

	tp := NewTileProvider("test", "test", md, manager, info, dimensions)

	if tp == nil {
		t.FailNow()
	}

	layers := map[string]Provider{"landsat2000": tp}
	metadata := map[string]string{}
	service := NewTileService(layers, metadata, nil, false, "ul")

	hreq := &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0/landsat2000/16/53958/24829.png")

	tileReq := request.NewTileRequest(hreq)

	resp := service.GetMap(tileReq)

	if resp == nil {
		t.FailNow()
	}

	hreq = &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0/landsat2000")

	tmsReq := request.NewTMSRequest(hreq)

	resp = service.GetCapabilities(tmsReq)

	if resp == nil {
		t.FailNow()
	}

	hreq = &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tiles/1.0.0")

	tmsReq = request.NewTMSRequest(hreq)

	resp = service.GetCapabilities(tmsReq)

	if resp == nil {
		t.FailNow()
	}

	hreq = &http.Request{}
	hreq.URL, _ = url.Parse("http://tms.osgeo.org/tms/")

	tmsReq = request.NewTMSRequest(hreq)

	resp = service.RootResource(tmsReq)

	if resp == nil {
		t.FailNow()
	}

	os.RemoveAll("./test_cache")
}
