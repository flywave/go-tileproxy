package service

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
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

	tileClient := client.NewMapboxTileClient("https://api.mapbox.com", "v1", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", "mapbox.mapbox-streets-v8", ctx)

	source := &sources.MapboxTileSource{Grid: grid, Client: tileClient, SourceCreater: ccreater}

	locker := &cache.DummyTileLocker{}

	mockGlyphsClient := client.NewMapboxGlyphsClient("https://api.mapbox.com", "v1", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", ctx)
	glyphsCache := resource.NewGlyphsCache(resource.NewLocalStore("./test_glyphs_cache"))
	glyphProvider := &GlyphProvider{sources.NewMapboxGlyphsSource(mockGlyphsClient, glyphsCache)}

	mockStyleClient := client.NewMapboxStyleClient("https://api.mapbox.com", "v1", "flywave", "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", ctx)
	stylesCache := resource.NewStyleCache(resource.NewLocalStore("./test_styles_cache"))
	styleProvider := &StyleProvider{sources.NewMapboxStyleSource(mockStyleClient, stylesCache)}

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, -1, false, 0, [2]uint32{1, 1})

	md := make(map[string]string)

	tp := NewMapboxTileProvider("test", MapboxVector, md, manager, nil, nil, nil)

	if tp == nil {
		t.FailNow()
	}

	service := NewMapboxService(map[string]Provider{"mapbox.mapbox-streets-v8": tp}, map[string]*StyleProvider{"cjikt35x83t1z2rnxpdmjs7y7": styleProvider}, map[string]*GlyphProvider{"font": glyphProvider}, md, nil)

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

	os.RemoveAll("./test_cache")
}
