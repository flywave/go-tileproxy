package seed

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type mockClient struct {
	client.HttpClient
	data []byte
	url  string
	body []byte
	code int
}

func (c *mockClient) Open(url string, data []byte) (statusCode int, body []byte) {
	c.data = data
	c.url = url
	return c.code, c.body
}

type dummyCreater struct {
}

func (c *dummyCreater) Create(size [2]uint32, opts tile.TileOptions, data interface{}) tile.Source {
	if data != nil {
		reader := data.(io.Reader)
		buf, _ := ioutil.ReadAll(reader)
		return &tile.DummyTileSource{Data: string(buf)}
	}
	return nil
}

func makeBBoxTask(tile_mgr cache.Manager, bbox vec2d.Rect, srs geo.Proj, levels []int) *TileSeedTask {
	md := map[string]string{"name": "", "cache_name": "", "grid_name": ""}

	coverage := geo.NewBBoxCoverage(bbox, srs, false)
	return NewTileSeedTask(md, tile_mgr, levels, nil, coverage)
}

type MockProgressLogger struct {
}

func (m *MockProgressLogger) LogMessage() {

}

func (m *MockProgressLogger) LogStep(progress *SeedProgress) {

}

func (m *MockProgressLogger) LogProgress(seed *SeedProgress, level int, bbox vec2d.Rect, tiles int) {

}

func (m *MockProgressLogger) SetCurrentTaskID(id string) {

}

func TestSeeder(t *testing.T) {
	mock := &mockClient{code: 200, body: []byte{0}}
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	urlTemplate := client.NewURLTemplate("/{{ .tms_path }}.png", "")

	client := client.NewTileClient(grid, urlTemplate, mock)

	creater := &dummyCreater{}

	ccreater := func(location string) tile.Source {
		source := vector.NewMVTSource([3]int{13515, 6392, 14}, vector.PBF_PTOTO_MAPBOX, &vector.VectorOptions{Format: vector.PBF_MIME_MAPBOX})
		source.SetSource("../data/3194.mvt")
		return source
	}

	source := &sources.TileSource{Grid: grid, Client: client, SourceCreater: creater}

	c := cache.NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	locker := &cache.DummyTileLocker{}

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, 0, false, 0, [2]uint32{2, 2})

	seedTask := makeBBoxTask(manager, vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, geo.NewSRSProj4("EPSG:4326"), []int{0, 1, 2})

	ctx := &Context{}

	mockLog := &MockProgressLogger{}

	seeder := NewTileWalker(seedTask, ctx, false, 0, mockLog, nil, false, true)
	seeder.Walk()
}
