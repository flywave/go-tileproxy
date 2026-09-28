package cache

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"sync/atomic"
	"testing"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

// countingSource stands in for the origin sources and counts how often it
// was consulted. It is safe for concurrent use (the background refresher
// also calls it).
type countingSource struct {
	layer.MapLayer
	calls int64
}

func (s *countingSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	atomic.AddInt64(&s.calls, 1)
	rgba := image.NewRGBA(image.Rect(0, 0, int(query.Size[0]), int(query.Size[1])))
	buf := &bytes.Buffer{}
	png.Encode(buf, rgba)
	opts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	src := imagery.CreateImageSourceFromBufer(buf.Bytes(), opts)
	src.SetCacheable(&tile.CacheInfo{Cacheable: true})
	return src, nil
}

func policyTestSetup(t *testing.T, policy CachePolicy) (*TileManager, *countingSource) {
	t.Helper()

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	creater := &mockImageSourceCreater{imageopts: imageOpts}
	localCache := NewLocalCache(t.TempDir(), "tms", creater)

	source := &countingSource{}
	manager := NewTileManager(&TileManagerOptions{
		Sources:      []layer.Layer{source},
		Grid:         grid,
		Cache:        localCache,
		Locker:       &DummyTileLocker{},
		Identifier:   "policy-test",
		Format:       "png",
		Options:      imageOpts,
		Policy:       policy,
		RescaleTiles: -1,
		MetaSize:     [2]uint32{1, 1},
		MetaBuffer:   0,
	})

	return manager, source
}

func originCalls(src *countingSource) int64 {
	return atomic.LoadInt64(&src.calls)
}

func ageTile(t *testing.T, mgr *TileManager, coord [3]int, age time.Duration) {
	t.Helper()
	local, ok := mgr.GetCache().(*LocalCache)
	if !ok {
		t.Fatal("expected a LocalCache")
	}
	tile := NewTile(coord)
	loc, err := local.TileLocation(tile, false)
	if err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-age)
	if err := os.Chtimes(loc, stale, stale); err != nil {
		t.Fatal(err)
	}
}

func TestCachePolicy_TTLRefreshesFromSource(t *testing.T) {
	mgr, src := policyTestSetup(t, CachePolicy{Mode: PolicyTTL, TTL: time.Hour})
	coord := [3]int{1, 2, 3}
	dims := utils.Dimensions{}

	if _, err := mgr.LoadTileCoord(coord, dims, true); err != nil {
		t.Fatal(err)
	}
	if n := originCalls(src); n != 1 {
		t.Fatalf("expected 1 origin call, got %d", n)
	}

	// within the TTL the cached tile is served
	if _, err := mgr.LoadTileCoord(coord, dims, true); err != nil {
		t.Fatal(err)
	}
	if n := originCalls(src); n != 1 {
		t.Fatalf("cache hit within ttl must not consult origin, got %d calls", n)
	}

	ageTile(t, mgr, coord, 2*time.Hour)

	if _, err := mgr.LoadTileCoord(coord, dims, true); err != nil {
		t.Fatal(err)
	}
	if n := originCalls(src); n != 2 {
		t.Fatalf("expired tile must refresh from origin, got %d calls", n)
	}
}

func TestCachePolicy_CacheFirstNeverRevalidates(t *testing.T) {
	mgr, src := policyTestSetup(t, DefaultPolicy())
	coord := [3]int{1, 2, 3}
	dims := utils.Dimensions{}

	if _, err := mgr.LoadTileCoord(coord, dims, true); err != nil {
		t.Fatal(err)
	}

	ageTile(t, mgr, coord, 365*24*time.Hour)

	tile, err := mgr.LoadTileCoord(coord, dims, true)
	if err != nil {
		t.Fatal(err)
	}
	if tile.Source == nil {
		t.Fatal("cache_first must serve the cached tile")
	}
	if n := originCalls(src); n != 1 {
		t.Fatalf("cache_first must never revalidate, got %d origin calls", n)
	}
}

func TestCachePolicy_BypassSkipsCache(t *testing.T) {
	mgr, src := policyTestSetup(t, CachePolicy{Mode: PolicyBypass})
	coord := [3]int{1, 2, 3}
	dims := utils.Dimensions{}

	for i := 0; i < 2; i++ {
		if _, err := mgr.LoadTileCoord(coord, dims, true); err != nil {
			t.Fatal(err)
		}
	}
	if n := originCalls(src); n != 2 {
		t.Fatalf("bypass must consult the origin every time, got %d calls", n)
	}
	if mgr.IsCached(coord, dims) {
		t.Fatal("bypass must not populate the cache")
	}
	local := mgr.GetCache().(*LocalCache)
	tile := NewTile(coord)
	loc, err := local.TileLocation(tile, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(loc); !os.IsNotExist(err) {
		t.Fatal("bypass wrote a cache file")
	}
}

func TestCachePolicy_RevalidateServesStaleAndRefreshes(t *testing.T) {
	mgr, src := policyTestSetup(t, CachePolicy{Mode: PolicyRevalidate, TTL: time.Hour})
	coord := [3]int{1, 2, 3}
	dims := utils.Dimensions{}

	if _, err := mgr.LoadTileCoord(coord, dims, true); err != nil {
		t.Fatal(err)
	}

	ageTile(t, mgr, coord, 2*time.Hour)

	local := mgr.GetCache().(*LocalCache)
	loc, err := local.TileLocation(NewTile(coord), false)
	if err != nil {
		t.Fatal(err)
	}

	// the stale value is served immediately, without blocking on the origin
	before := originCalls(src)
	tile, err := mgr.LoadTileCoord(coord, dims, true)
	if err != nil {
		t.Fatal(err)
	}
	if tile.Source == nil {
		t.Fatal("revalidate must serve the stale tile")
	}

	// the background refresh consults the origin and rewrites the file
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if originCalls(src) > before {
			if info, err := os.Stat(loc); err == nil && time.Since(info.ModTime()) < time.Hour {
				return // refreshed
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("background refresh did not run")
}

func TestParseCachePolicy(t *testing.T) {
	p, err := ParseCachePolicy("cache_first", 0)
	if err != nil || p.Mode != PolicyCacheFirst {
		t.Fatalf("cache_first parse failed: %v", err)
	}
	if _, err := ParseCachePolicy("ttl", 0); err == nil {
		t.Fatal("ttl without ttl_hours must be rejected")
	}
	p, err = ParseCachePolicy("ttl", 24)
	if err != nil || p.TTL != 24*time.Hour {
		t.Fatalf("ttl parse failed: %v", err)
	}
	p, err = ParseCachePolicy("bypass", 0)
	if err != nil || p.Mode != PolicyBypass {
		t.Fatalf("bypass parse failed: %v", err)
	}
	if _, err := ParseCachePolicy("nonsense", 1); err == nil {
		t.Fatal("unknown mode must be rejected")
	}
	p, err = ParseCachePolicy("", 0)
	if err != nil || p.Mode != PolicyCacheFirst {
		t.Fatalf("empty mode should default to cache_first: %v", err)
	}
}
