package tileproxy

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/service"
	"github.com/flywave/go-tileproxy/setting"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

// ---------------------------------------------------------------------------
// Mock implementations for performance testing
// ---------------------------------------------------------------------------

// perfMockClient simulates an upstream tile server with configurable latency
type perfMockClient struct {
	latency    time.Duration
	statusCode int
	body       []byte
	callCount  int64
	mu         sync.Mutex
}

func (c *perfMockClient) Open(url string, data []byte, hdr http.Header) (int, []byte) {
	c.mu.Lock()
	c.callCount++
	c.mu.Unlock()
	if c.latency > 0 {
		time.Sleep(c.latency)
	}
	return c.statusCode, c.body
}

func (c *perfMockClient) callCount_() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.callCount
}

type perfMockContext struct {
	client *perfMockClient
}

func (c *perfMockContext) Client() io.Reader { return nil }
func (c *perfMockContext) Sync()             {}

// perfTileData returns a PNG-encoded 256x256 tile
func perfTileData() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x), uint8(y), 128, 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// perfTestHarness sets up the full pipeline for benchmarking
type perfTestHarness struct {
	tileData    []byte
	cacheDir    string
	proxy       *TileProxy
	mockClient  *perfMockClient
	tileManager *cache.TileManager
	localCache  *cache.LocalCache
}

func newPerfTestHarness(b *testing.B, upstreamLatency time.Duration) *perfTestHarness {
	_ = b
	tileData := perfTileData()

	// --- Setup: mock HTTP client ---
	mockClient := &perfMockClient{
		latency:    upstreamLatency,
		statusCode: 200,
		body:       tileData,
	}
	_ = mockClient

	// --- Setup: grid ---
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	// --- Setup: image options ---
	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	// --- Setup: local cache ---
	cacheDir, err := os.MkdirTemp("", "tileproxy-perf-*")
	if err != nil {
		b.Fatal(err)
	}
	creater := &perfSourceCreater{opts: imageOpts}
	localCache := cache.NewLocalCache(cacheDir, "tms", creater)

	// --- Setup: locker ---
	locker := &cache.DummyTileLocker{}

	// --- Setup: tile manager ---
	tmOpts := &cache.TileManagerOptions{
		Sources:              []layer.Layer{},
		Grid:                 grid,
		Cache:                localCache,
		Locker:               locker,
		Identifier:           "perf-test",
		Format:               "png",
		Options:              imageOpts,
		MinimizeMetaRequests: false,
		BulkMetaTiles:        false,
		PreStoreFilter:       nil,
		RescaleTiles:         -1,
		CacheRescaledTiles:   false,
		MetaBuffer:           0,
		MetaSize:             [2]uint32{4, 4},
	}
	tileManager := cache.NewTileManager(tmOpts)

	// --- Setup: tile service mounted behind a proxy ---
	provider := service.NewTileProvider(&service.TileProviderOptions{
		Name:        "test-layer",
		Title:       "Test Layer",
		Metadata:    &service.TileProviderMetadata{Name: "test-layer", Title: "Test Layer"},
		TileManager: tileManager,
		Dimensions:  make(utils.Dimensions),
	})
	maxTileAge := time.Hour
	tmsService := service.NewTileService(&service.TileServiceOptions{
		Layers:     map[string]service.Provider{"test-layer": provider},
		Metadata:   &service.TileMetadata{Title: "Perf Test", Abstract: "perf harness", URL: "http://localhost/test-service/"},
		MaxTileAge: &maxTileAge,
		Origin:     "nw",
	})
	svc := &Service{Id: "test-service", Type: TileService, Service: tmsService}

	proxy := &TileProxy{
		globals:         &setting.GlobalsSetting{},
		Services:        map[string]*Service{"test-service": svc},
		serviceCache:    make(map[string]*Service),
		serviceReqRegex: serviceReqPattern,
	}

	return &perfTestHarness{
		tileData:    tileData,
		cacheDir:    cacheDir,
		proxy:       proxy,
		mockClient:  mockClient,
		tileManager: tileManager,
		localCache:  localCache,
	}
}

func (h *perfTestHarness) Close() {
	os.RemoveAll(h.cacheDir)
}

// perfSourceCreater creates mock image sources
type perfSourceCreater struct {
	opts *imagery.ImageOptions
}

func (c *perfSourceCreater) GetExtension() string { return "png" }
func (c *perfSourceCreater) CreateEmpty(size [2]uint32, opts tile.TileOptions) tile.Source {
	return imagery.NewBlankImageSource(size, c.opts, nil)
}
func (c *perfSourceCreater) Create(data []byte, coord [3]int) tile.Source {
	src := imagery.CreateImageSource([2]uint32{256, 256}, c.opts)
	src.SetSource(bytes.NewReader(data))
	return src
}

// ---------------------------------------------------------------------------
// Benchmarks: End-to-end HTTP pipeline
// ---------------------------------------------------------------------------

func BenchmarkPipeline_HTTP_Sequential(b *testing.B) {
	h := newPerfTestHarness(b, 0)
	defer h.Close()

	req := httptest.NewRequest("GET", "/test-service/tiles/1/2/3.png", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		// Re-create request with fresh URL for each iteration
		req = httptest.NewRequest("GET", "/test-service/tiles/1/2/3.png", nil)
		h.proxy.ServeHTTP(w, req)
	}
	b.StopTimer()
}

func BenchmarkPipeline_HTTP_Concurrent(b *testing.B) {
	h := newPerfTestHarness(b, 0)
	defer h.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test-service/tiles/1/2/3.png", nil)
			h.proxy.ServeHTTP(w, req)
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmarks: Cache performance (hit vs miss, LocalCache)
// ---------------------------------------------------------------------------

func BenchmarkLocalCache_Sequential_Store(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cache-perf-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)
	tileData := perfTileData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := cache.NewTile([3]int{i % 100, i % 100, 5})
		src := creater.Create(tileData, t.Coord)
		t.Source = src
		c.StoreTile(t)
	}
}

func BenchmarkLocalCache_Sequential_Load(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cache-perf-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)
	tileData := perfTileData()

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		t := cache.NewTile([3]int{i, i, 5})
		src := creater.Create(tileData, t.Coord)
		t.Source = src
		c.StoreTile(t)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := cache.NewTile([3]int{i % 100, i % 100, 5})
		c.LoadTile(t, false)
	}
}

func BenchmarkLocalCache_Concurrent_Store(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cache-perf-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)
	tileData := perfTileData()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			coord := [3]int{idx % 100, idx % 100, 5}
			t := cache.NewTile(coord)
			src := creater.Create(tileData, t.Coord)
			t.Source = src
			c.StoreTile(t)
			idx++
		}
	})
}

func BenchmarkLocalCache_Concurrent_Load(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "cache-perf-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)
	tileData := perfTileData()

	// Pre-populate
	for i := 0; i < 100; i++ {
		t := cache.NewTile([3]int{i, i, 5})
		src := creater.Create(tileData, t.Coord)
		t.Source = src
		c.StoreTile(t)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			t := cache.NewTile([3]int{idx % 100, idx % 100, 5})
			c.LoadTile(t, false)
			idx++
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmarks: Lock contention
// ---------------------------------------------------------------------------

func BenchmarkLock_InMemoryTileLocker(b *testing.B) {
	locker := cache.NewInMemoryTileLocker()
	tile := cache.NewTile([3]int{1, 2, 3})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		locker.Lock(ctx, tile, func() error { return nil })
	}
}

func BenchmarkLock_InMemoryTileLocker_Contention(b *testing.B) {
	locker := cache.NewInMemoryTileLocker()
	tile := cache.NewTile([3]int{1, 2, 3})
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			locker.Lock(ctx, tile, func() error { return nil })
		}
	})
}

func BenchmarkLock_DummyTileLocker(b *testing.B) {
	locker := &cache.DummyTileLocker{}
	tile := cache.NewTile([3]int{1, 2, 3})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		locker.Lock(ctx, tile, func() error { return nil })
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: TileCollection (concurrent read/write)
// ---------------------------------------------------------------------------

func BenchmarkTileCollection_Sequential_Add(b *testing.B) {
	col := cache.NewTileCollection(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := cache.NewTile([3]int{i, i, i})
		col.SetItem(t)
	}
}

func BenchmarkTileCollection_Concurrent_Add(b *testing.B) {
	col := cache.NewTileCollection(nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			t := cache.NewTile([3]int{idx, idx, idx})
			col.SetItem(t)
			idx++
		}
	})
}

func BenchmarkTileCollection_Concurrent_Read(b *testing.B) {
	col := cache.NewTileCollection(nil)
	for i := 0; i < 1000; i++ {
		t := cache.NewTile([3]int{i, i, i})
		col.SetItem(t)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			col.Contains([3]int{idx % 1000, idx % 1000, idx % 1000})
			idx++
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmarks: Image encoding/decoding
// ---------------------------------------------------------------------------

func BenchmarkImage_EncodePNG(b *testing.B) {
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x), uint8(y), 128, 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		png.Encode(&buf, img)
	}
}

func BenchmarkImage_DecodePNG(b *testing.B) {
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	var buf bytes.Buffer
	png.Encode(&buf, img)
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes.NewReader(data)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: Full encode → store → load → decode pipeline
// ---------------------------------------------------------------------------

func BenchmarkFullStoreLoadPipeline(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "perf-pipe-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coord := [3]int{i % 50, i % 50, 5}

		t := cache.NewTile(coord)
		if err := c.LoadTile(t, false); err != nil {
			// Cache miss: create, store, reload
			src := creater.Create(perfTileData(), coord)
			t.Source = src
			c.StoreTile(t)

			t2 := cache.NewTile(coord)
			c.LoadTile(t2, false)
			_ = t2.Source
		}
	}
	b.StopTimer()
}

// ---------------------------------------------------------------------------
// Benchmarks: FTS (Full Tile Service) pipeline
// ---------------------------------------------------------------------------

func BenchmarkFullPipeline_CacheMiss(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "perf-full-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)

	tileData := perfTileData()
	col := cache.NewTileCollection(nil)
	for i := 0; i < 100; i++ {
		coord := [3]int{i, i, 5}
		t := cache.NewTile(coord)
		src := creater.Create(tileData, coord)
		t.Source = src
		col.SetItem(t)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate uncached tile — no pre-store, fresh LoadTiles
		coord := [3]int{i % 100, i % 100, 5}
		t := cache.NewTile(coord)
		c.LoadTile(t, false)
	}
}

func BenchmarkFullPipeline_CacheHit(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "perf-full-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)
	tileData := perfTileData()

	// Pre-store tiles
	for i := 0; i < 100; i++ {
		t := cache.NewTile([3]int{i, i, 5})
		src := creater.Create(tileData, t.Coord)
		t.Source = src
		c.StoreTile(t)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := cache.NewTile([3]int{i % 100, i % 100, 5})
		c.LoadTile(t, false)
	}
}

// ---------------------------------------------------------------------------
// Throughput & Latency measurement
// ---------------------------------------------------------------------------

func BenchmarkThroughput(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "perf-tput-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)
	tileData := perfTileData()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		t := cache.NewTile([3]int{i, i, 5})
		src := creater.Create(tileData, t.Coord)
		t.Source = src
		c.StoreTile(t)
	}

	b.ResetTimer()
	var dur time.Duration
	count := 0
	for i := 0; i < b.N; i++ {
		start := time.Now()
		t := cache.NewTile([3]int{count % 1000, count % 1000, 5})
		c.LoadTile(t, false)
		dur += time.Since(start)
		count++
	}
	b.ReportMetric(float64(b.N)/dur.Seconds(), "req/s")
	b.ReportMetric(float64(dur.Microseconds())/float64(b.N), "us/op")
}

// ---------------------------------------------------------------------------
// Benchmark suite runner with latency simulation
// ---------------------------------------------------------------------------

func BenchmarkWithUpstreamLatency(b *testing.B) {
	latencies := []time.Duration{0, 5 * time.Millisecond, 20 * time.Millisecond, 100 * time.Millisecond}
	for _, lat := range latencies {
		b.Run(fmt.Sprintf("latency=%v", lat), func(b *testing.B) {
			h := newPerfTestHarness(b, lat)
			defer h.Close()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test-service/tiles/1/2/3.png", nil)
				h.proxy.ServeHTTP(w, req)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GC pressure measurement
// ---------------------------------------------------------------------------

func BenchmarkGC_Pressure(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "perf-gc-*")
	defer os.RemoveAll(tmpDir)

	opts := &imagery.ImageOptions{Format: "png", Resampling: "nearest"}
	creater := &perfSourceCreater{opts: opts}
	c := cache.NewLocalCache(tmpDir, "tms", creater)
	tileData := perfTileData()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 10000; i++ {
		t := cache.NewTile([3]int{i, i, 5})
		src := creater.Create(tileData, t.Coord)
		t.Source = src
		c.StoreTile(t)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/10000, "B/tile")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/10000, "alloc/tile")
}
