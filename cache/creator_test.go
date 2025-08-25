package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

// Mock implementations for testing

type creatorMockCache struct {
	Cache
	storedTiles map[string]*Tile
	loadError   error
	storeError  error
}

func newCreatorMockCache() *creatorMockCache {
	return &creatorMockCache{
		storedTiles: make(map[string]*Tile),
	}
}

func (m *creatorMockCache) LoadTile(tile *Tile, withMetadata bool) error {
	if m.loadError != nil {
		return m.loadError
	}
	key := creatorTileKey(tile.Coord)
	if stored, exists := m.storedTiles[key]; exists {
		tile.Source = stored.Source
		tile.Cacheable = stored.Cacheable
		return nil
	}
	return errors.New("tile not found")
}

func (m *creatorMockCache) StoreTile(tile *Tile) error {
	if m.storeError != nil {
		return m.storeError
	}
	key := creatorTileKey(tile.Coord)
	m.storedTiles[key] = tile
	tile.Stored = true
	return nil
}

func (m *creatorMockCache) LoadTiles(tiles *TileCollection, withMetadata bool) error {
	for _, tile := range tiles.tiles {
		if err := m.LoadTile(tile, withMetadata); err != nil {
			return err
		}
	}
	return nil
}

func (m *creatorMockCache) StoreTiles(tiles *TileCollection) error {
	for _, tile := range tiles.tiles {
		if err := m.StoreTile(tile); err != nil {
			return err
		}
	}
	return nil
}

func (m *creatorMockCache) IsCached(tile *Tile) bool {
	key := creatorTileKey(tile.Coord)
	_, exists := m.storedTiles[key]
	return exists
}

func creatorTileKey(coord [3]int) string {
	return string(rune(coord[0])) + string(rune(coord[1])) + string(rune(coord[2]))
}

type creatorMockLayer struct {
	layer.MapLayer
	mapData     tile.Source
	getMapError error
}

func newCreatorMockLayer(data string) *creatorMockLayer {
	// Create a proper ImageSource that returns image.Image instead of string
	imageOpts := &imagery.ImageOptions{Format: tile.TileFormat("png")}
	source := imagery.CreateImageSource([2]uint32{256, 256}, imageOpts)
	var cacheInfo tile.CacheInfo
	cacheInfo.Cacheable = true
	cacheInfo.Timestamp = time.Now()
	cacheInfo.Size = 1024
	source.SetCacheable(&cacheInfo)
	return &creatorMockLayer{
		mapData: source,
	}
}

func (m *creatorMockLayer) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if m.getMapError != nil {
		return nil, m.getMapError
	}
	return m.mapData, nil
}

type creatorMockManager struct {
	Manager
	cache           Cache
	sources         []layer.Layer
	grid            *geo.TileGrid
	metaGrid        *geo.MetaGrid
	requestFormat   string
	tileOptions     tile.TileOptions
	minimizeMetaReq bool
	lockFunc        func() error
	filterFunc      func(*Tile) (*Tile, error)
	cachedTiles     map[string]bool
}

func newCreatorMockManager() *creatorMockManager {
	// Create a proper TileGrid to avoid nil pointer dereference
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	grid := geo.NewTileGrid(opts)
	// Create proper ImageOptions for testing
	imageOpts := &imagery.ImageOptions{
		Format: tile.TileFormat("png"),
	}
	return &creatorMockManager{
		cache:         newCreatorMockCache(),
		grid:          grid,
		requestFormat: "png",
		tileOptions:   imageOpts,
		cachedTiles:   make(map[string]bool),
		lockFunc:      func() error { return nil },
		filterFunc:    func(t *Tile) (*Tile, error) { return t, nil },
	}
}

func (m *creatorMockManager) GetCache() Cache {
	return m.cache
}

func (m *creatorMockManager) GetSources() []layer.Layer {
	return m.sources
}

func (m *creatorMockManager) GetGrid() *geo.TileGrid {
	return m.grid
}

func (m *creatorMockManager) GetMetaGrid() *geo.MetaGrid {
	return m.metaGrid
}

func (m *creatorMockManager) GetRequestFormat() string {
	return m.requestFormat
}

func (m *creatorMockManager) GetTileOptions() tile.TileOptions {
	return m.tileOptions
}

func (m *creatorMockManager) GetMinimizeMetaRequests() bool {
	return m.minimizeMetaReq
}

func (m *creatorMockManager) Lock(ctx context.Context, tile *Tile, fn func() error) error {
	if err := m.lockFunc(); err != nil {
		return err
	}
	return fn() // Actually execute the function!
}

func (m *creatorMockManager) ApplyTileFilter(tile *Tile) (*Tile, error) {
	return m.filterFunc(tile)
}

func (m *creatorMockManager) IsCached(coord [3]int, dim utils.Dimensions) bool {
	key := creatorTileKey(coord)
	return m.cachedTiles[key]
}

func (m *creatorMockManager) setCached(coord [3]int, cached bool) {
	key := creatorTileKey(coord)
	m.cachedTiles[key] = cached
}

type creatorMockTileMerger struct {
	tile.Merger
	mergeError error
}

func (m *creatorMockTileMerger) AddSource(src tile.Source, cov geo.Coverage) {
	// Mock implementation
}

func (m *creatorMockTileMerger) Merge(opts tile.TileOptions, size []uint32, bbox vec2d.Rect, bbox_srs geo.Proj, coverage geo.Coverage) tile.Source {
	if m.mergeError != nil {
		return nil
	}
	source := newTestSource("merged")
	var cacheInfo tile.CacheInfo
	cacheInfo.Cacheable = true
	cacheInfo.Timestamp = time.Now()
	cacheInfo.Size = 1024
	source.SetCacheable(&cacheInfo)
	return source
}

// Test cases

func TestNewTileCreator(t *testing.T) {
	manager := newCreatorMockManager()
	dimensions := utils.Dimensions{}
	merger := &creatorMockTileMerger{}

	creator := NewTileCreator(manager, dimensions, merger, true)

	if creator == nil {
		t.Fatal("Expected non-nil TileCreator")
	}
	if creator.manager != manager {
		t.Error("Manager not set correctly")
	}
	if creator.bulkMetaTiles != true {
		t.Error("BulkMetaTiles not set correctly")
	}
}

func TestTileCreator_IsCached(t *testing.T) {
	manager := newCreatorMockManager()
	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	coord := [3]int{1, 2, 3}

	// Initially not cached
	if creator.IsCached(coord) {
		t.Error("Expected tile to not be cached initially")
	}

	// Set as cached
	manager.setCached(coord, true)
	if !creator.IsCached(coord) {
		t.Error("Expected tile to be cached")
	}
}

func TestTileCreator_CreateTiles_NoSources(t *testing.T) {
	manager := newCreatorMockManager()
	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tiles := []*Tile{NewTile([3]int{1, 2, 3})}
	result, err := creator.CreateTiles(tiles)

	if err == nil {
		t.Error("Expected error when sources is nil")
	}
	if len(result) != 0 {
		t.Error("Expected empty result when sources is nil")
	}
}

func TestTileCreator_CreateTiles_SingleTile(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tiles := []*Tile{NewTile([3]int{1, 2, 3})}
	result, err := creator.CreateTiles(tiles)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 tile, got %d", len(result))
	}
}

func TestTileCreator_CreateTiles_WithMetaGrid_MinimizeRequests(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)
	// Use proper MetaGrid initialization
	manager.metaGrid = geo.NewMetaGrid(manager.grid, [2]uint32{2, 2}, 0)
	manager.minimizeMetaReq = true

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tiles := []*Tile{
		NewTile([3]int{1, 2, 3}),
		NewTile([3]int{2, 3, 3}),
	}
	result, err := creator.CreateTiles(tiles)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Error("Expected some tiles to be created")
	}
}

func TestTileCreator_CreateSingleTile_NotCached(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tile := NewTile([3]int{1, 2, 3})
	result, err := creator.createSingleTile(tile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestTileCreator_CreateSingleTile_Cached(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tileid := NewTile([3]int{1, 2, 3})
	manager.setCached(tileid.Coord, true)

	// Store the tile in cache first
	testTile := NewTile(tileid.Coord)
	testSource := newTestSource("cached_data")
	var cacheInfo tile.CacheInfo
	cacheInfo.Cacheable = true
	cacheInfo.Timestamp = time.Now()
	cacheInfo.Size = 1024
	testSource.SetCacheable(&cacheInfo)
	testTile.Source = testSource
	manager.cache.StoreTile(testTile)

	result, err := creator.createSingleTile(tileid)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestTileCreator_QuerySources_NoLayers(t *testing.T) {
	manager := newCreatorMockManager()
	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	query := &layer.MapQuery{
		BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:  geo.NewProj("EPSG:3857"),
	}

	result, err := creator.querySources(query)

	if err == nil {
		t.Error("Expected error when no sources available")
	}
	if result != nil {
		t.Error("Expected nil result when no sources available")
	}
}

func TestTileCreator_QuerySources_SingleLayer(t *testing.T) {
	manager := newCreatorMockManager()
	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)
	creator.sources = []layer.Layer{newCreatorMockLayer("test_data")}

	query := &layer.MapQuery{
		BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:  geo.NewProj("EPSG:3857"),
	}

	result, err := creator.querySources(query)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestTileCreator_QuerySources_LayerError(t *testing.T) {
	manager := newCreatorMockManager()
	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	mockLayer := newCreatorMockLayer("test_data")
	mockLayer.getMapError = errors.New("layer error")
	creator.sources = []layer.Layer{mockLayer}

	query := &layer.MapQuery{
		BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:  geo.NewProj("EPSG:3857"),
	}

	result, err := creator.querySources(query)

	if err == nil {
		t.Error("Expected error when layer returns error")
	}
	if result != nil {
		t.Error("Expected nil result when layer returns error")
	}
}

func TestTileCreator_LockTimeout(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)

	// Simulate lock taking some time but within timeout
	manager.lockFunc = func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tile := NewTile([3]int{1, 2, 3})

	result, err := creator.createSingleTile(tile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestTileCreator_FilterError(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)
	manager.filterFunc = func(tile *Tile) (*Tile, error) {
		return nil, errors.New("filter error")
	}

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tile := NewTile([3]int{1, 2, 3})
	// Ensure tile is NOT cached so that the filter will be applied
	manager.setCached(tile.Coord, false)
	result, err := creator.createSingleTile(tile)

	if err == nil {
		t.Error("Expected filter error")
	}
	if result != nil {
		t.Error("Expected nil result on filter error")
	}
}

func TestTileCreator_CacheStoreError(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)

	mockCache := manager.cache.(*creatorMockCache)
	mockCache.storeError = errors.New("store error")

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tile := NewTile([3]int{1, 2, 3})
	// Ensure tile is NOT cached so that it will try to store
	manager.setCached(tile.Coord, false)
	result, err := creator.createSingleTile(tile)

	if err == nil {
		t.Error("Expected store error")
	}
	if result == nil {
		t.Error("Expected tile object even on store error")
	}
}

func TestTileCreator_CacheLoadError(t *testing.T) {
	manager := newCreatorMockManager()
	manager.sources = []layer.Layer{newCreatorMockLayer("test_data")}
	// Use proper TileGrid initialization
	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:3857"
	manager.grid = geo.NewTileGrid(opts)

	mockCache := manager.cache.(*creatorMockCache)
	mockCache.loadError = errors.New("load error")

	creator := NewTileCreator(manager, utils.Dimensions{}, &creatorMockTileMerger{}, false)

	tile := NewTile([3]int{1, 2, 3})
	manager.setCached(tile.Coord, true)

	result, err := creator.createSingleTile(tile)

	if err == nil {
		t.Error("Expected load error")
	}
	if result == nil {
		t.Error("Expected tile object even on load error")
	}
}
