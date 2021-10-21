package task

import (
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geos"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/vector"
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

func makeBBoxTask(tile_mgr cache.Manager, bbox vec2d.Rect, srs geo.Proj, levels []int) *TileSeedTask {
	md := map[string]string{"name": "", "cache_name": "", "grid_name": ""}
	coverage := geo.NewBBoxCoverage(bbox, srs, false)
	return NewTileSeedTask(md, tile_mgr, levels, nil, coverage)
}

func makeGeomTask(tile_mgr cache.Manager, geom *geos.Geometry, srs geo.Proj, levels []int) *TileSeedTask {
	md := map[string]string{"name": "", "cache_name": "", "grid_name": ""}
	coverage := geo.NewGeosCoverage(geom, srs, false)
	return NewTileSeedTask(md, tile_mgr, levels, nil, coverage)
}

type MockLogWriter struct {
	LogWriter
	out []string
}

func (l *MockLogWriter) WriteString(s string) (n int, err error) {
	l.out = append(l.out, s)
	return len(s), nil
}

func (l *MockLogWriter) Flush() error {
	return nil
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

type mockWorkerPool struct {
	seedTiles map[int][][2]int
}

func (p *mockWorkerPool) Process(work Work, progress *TaskProgress) {
	if p.seedTiles == nil {
		p.seedTiles = make(map[int][][2]int)
	}
	tiles := work.(*SeedWorker).tiles
	for j := range tiles {
		if _, ok := p.seedTiles[tiles[j][2]]; !ok {
			p.seedTiles[tiles[j][2]] = [][2]int{}
		}
		p.seedTiles[tiles[j][2]] = append(p.seedTiles[tiles[j][2]], [2]int{tiles[j][0], tiles[j][1]})
	}
}

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

func seeder(bbox vec2d.Rect, levels []int, seedProgress *TaskProgress, t *testing.T) map[int][][2]int {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	urlTemplate := client.NewURLTemplate("/{tms_path}.png", "", nil)

	client := client.NewTileClient(grid, urlTemplate, ctx)

	ccreater := &mockMVTSourceCreater{}

	source := &sources.TileSource{Grid: grid, Client: client, SourceCreater: ccreater}

	c := cache.NewLocalCache("./test_cache", "quadkey", ccreater)

	locker := &cache.DummyTileLocker{}

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, 0, false, 0, [2]uint32{2, 2})

	seedTask := makeBBoxTask(manager, bbox, geo.NewProj(4326), levels)

	local := NewLocalProgressStore("./test.task", false)

	logger := NewDefaultProgressLogger(&MockLogWriter{}, false, true, local)

	tile_worker_pool := &mockWorkerPool{}

	seeder := NewTileWalker(seedTask, tile_worker_pool, false, 0, logger, seedProgress, false, true)
	seeder.Walk()

	return tile_worker_pool.seedTiles
}

func seederGeom(geom *geos.Geometry, levels []int, t *testing.T) map[int][][2]int {
	mock := &mockClient{code: 200, body: []byte{0}}
	ctx := &mockContext{c: mock}

	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	urlTemplate := client.NewURLTemplate("/{tms_path}.png", "", nil)

	client := client.NewTileClient(grid, urlTemplate, ctx)

	ccreater := &mockMVTSourceCreater{}

	source := &sources.TileSource{Grid: grid, Client: client, SourceCreater: ccreater}

	c := cache.NewLocalCache("./test_cache", "quadkey", ccreater)

	locker := &cache.DummyTileLocker{}

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, 0, false, 0, [2]uint32{2, 2})

	seedTask := makeGeomTask(manager, geom, geo.NewProj(4326), levels)

	local := NewLocalProgressStore("./test.task", false)

	logger := NewDefaultProgressLogger(&MockLogWriter{}, false, true, local)

	tile_worker_pool := &mockWorkerPool{}

	seeder := NewTileWalker(seedTask, tile_worker_pool, false, 0, logger, nil, false, true)
	seeder.Walk()

	return tile_worker_pool.seedTiles
}

func assertTileInTiles(aa [2]int, b [][2]int, t *testing.T) {
	flag := false
	for i := range b {
		bb := b[i]

		if aa[0] == bb[0] && aa[1] == bb[1] {
			flag = true
		}
	}

	if !flag {
		t.FailNow()
	}
}

func assertTiles(a [][2]int, b [][2]int, t *testing.T) {
	if len(a) != len(b) {
		t.FailNow()
	}
	for i := range a {
		aa := a[i]
		assertTileInTiles(aa, b, t)
	}
}

func TestSeederBBox(t *testing.T) {
	seederLevelsCounts := []int{3, 3, 2}
	seederLevelsBBox := []vec2d.Rect{
		{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}},
		{Min: vec2d.T{-45, 0}, Max: vec2d.T{180, 90}},
		{Min: vec2d.T{-45, 0}, Max: vec2d.T{180, 90}},
	}
	seederLevelsLevels := [][]int{
		{0, 1, 2},
		{0, 1, 2},
		{0, 2},
	}

	seederLevelsResults := []map[int][][2]int{
		{
			0: [][2]int{{0, 0}},
			1: [][2]int{{0, 0}, {1, 0}},
			2: [][2]int{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {0, 1}, {1, 1}, {2, 1}, {3, 1}},
		},
		{
			0: [][2]int{{0, 0}},
			1: [][2]int{{0, 0}, {1, 0}},
			2: [][2]int{{1, 1}, {2, 1}, {3, 1}},
		},
		{
			0: [][2]int{{0, 0}},
			2: [][2]int{{1, 1}, {2, 1}, {3, 1}},
		},
	}

	for i := range seederLevelsCounts {
		seeded_tiles := seeder(seederLevelsBBox[i], seederLevelsLevels[i], nil, t)

		if len(seeded_tiles) != seederLevelsCounts[i] {
			t.FailNow()
		}

		for l := range seeded_tiles {
			assertTiles(seeded_tiles[l], seederLevelsResults[i][l], t)
		}
	}
}

func TestSeederGeom(t *testing.T) {
	geom := geos.CreateFromWKT("POLYGON((10 10, 10 50, -10 60, 10 80, 80 80, 80 10, 10 10))")
	seeded_tiles := seederGeom(geom, []int{0, 1, 2, 3, 4}, t)

	if len(seeded_tiles) != 5 {
		t.FailNow()
	}

	seederLevelsResults := map[int][][2]int{
		0: {{0, 0}},
		1: {{0, 0}, {1, 0}},
		2: {{1, 1}, {2, 1}},
		3: {{4, 2}, {5, 2}, {4, 3}, {5, 3}, {3, 3}},
	}

	for l := range seeded_tiles {
		if _, ok := seederLevelsResults[l]; ok {
			assertTiles(seeded_tiles[l], seederLevelsResults[l], t)
		} else {
			if len(seeded_tiles[l]) != 4*4+2 {
				t.FailNow()
			}
		}
	}
}

func TestSeederFullBBoxContinue(t *testing.T) {
	seedProgress := NewTaskProgress([]interface{}{[2]int{0, 1}, [2]int{1, 2}})
	seeded_tiles := seeder(vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}, []int{0, 1, 2}, seedProgress, t)

	if len(seeded_tiles) != 3 {
		t.FailNow()
	}

	seederLevelsResults := map[int][][2]int{
		0: {{0, 0}},
		1: {{0, 0}, {1, 0}},
		2: {{2, 0}, {3, 0}, {2, 1}, {3, 1}},
	}

	for l := range seeded_tiles {
		assertTiles(seeded_tiles[l], seederLevelsResults[l], t)
	}
}
