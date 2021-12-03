package service

import (
	"net/http"
	"net/url"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
)

type mockMVTSourceCreater struct {
}

func (c *mockMVTSourceCreater) GetExtension() string {
	return "mvt"
}

func (c *mockMVTSourceCreater) CreateEmpty(size [2]uint32, opts tile.TileOptions) tile.Source {
	return nil
}

func (c *mockMVTSourceCreater) Create(data []byte, tile [3]int) tile.Source {
	source := vector.NewMVTSource([3]int{13515, 6392, 14}, vector.PBF_PTOTO_MAPBOX, &vector.VectorOptions{Format: vector.PBF_MIME_MAPBOX})
	source.SetSource("../data/3194.mvt")
	return source
}

func TestMapboxServiceGetTile(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{}}
	ctx := &mockContext{c: mock}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	ccreater := &mockMVTSourceCreater{}

	c := cache.NewLocalCache("./test_cache", "quadkey", ccreater)

	tileClient := client.NewMapboxTileClient("http://a.tiles.mapbox.com/v4/mapbox.mapbox-streets-v8/{z}/{x}/{y}.vector.pbf", "http://a.tiles.mapbox.com/v4/mapbox.mapbox-streets-v8", "{token}", "access_token", ctx)

	source := &sources.MapboxTileSource{Grid: grid, Client: tileClient, SourceCreater: ccreater}

	locker := &cache.DummyTileLocker{}

	mockGlyphsClient := client.NewMapboxStyleClient("http://api.mapbox.com/fonts/v1/examples", "{token}", "access_token", ctx)
	glyphsCache := resource.NewGlyphsCache(resource.NewLocalStore("./test_glyphs_cache"))

	mockStyleClient := client.NewMapboxStyleClient("https://api.mapbox.com/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7", "{token}", "access_token", ctx)
	stylesCache := resource.NewStyleCache(resource.NewLocalStore("./test_styles_cache"))
	styleProvider := &StyleProvider{styleSource: sources.NewMapboxStyleSource(mockStyleClient, nil, stylesCache), glyphsSource: sources.NewMapboxGlyphsSource(mockGlyphsClient, []string{"mock"}, glyphsCache)}

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
		MetaSize:             [2]uint32{2, 2},
	}

	manager := cache.NewTileManager(topts)

	lmd := &MapboxLayerMetadata{}

	tiopts := &MapboxTileOptions{Name: "test", Type: MapboxVector, Metadata: lmd, TileManager: manager}

	tp := NewMapboxTileProvider(tiopts)

	if tp == nil {
		t.FailNow()
	}

	md := &MapboxMetadata{}

	sopts := &MapboxServiceOptions{Tilesets: map[string]Provider{"mapbox.mapbox-streets-v8": tp}, Styles: map[string]*StyleProvider{"cjikt35x83t1z2rnxpdmjs7y7": styleProvider}, Metadata: md, MaxTileAge: nil}

	service := NewMapboxService(sopts)

	hreq := &http.Request{}
	hreq.URL, _ = url.Parse("https://127.0.0.1/v4/mapbox.mapbox-streets-v8/14/13515/6392.mvt")

	tileReq := request.NewMapboxTileRequest(hreq, false)

	resp := service.GetTile(tileReq)

	if resp == nil {
		t.FailNow()
	}

	hreq = &http.Request{}
	hreq.URL, _ = url.Parse("https://api.mapbox.com/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7")

	styleReq := request.NewMapboxStyleRequest(hreq, false)

	resp = service.GetStyle(styleReq)

	if resp == nil {
		t.FailNow()
	}

	//os.RemoveAll("./test_cache")
}
