package service

import (
	"context"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

// MockTileResponse implements TileResponse interface for testing
type MockTileResponse struct {
	buffer    []byte
	format    string
	timestamp time.Time
	size      int
	cacheable bool
}

func (m *MockTileResponse) getBuffer() []byte {
	return m.buffer
}

func (m *MockTileResponse) getFormat() string {
	return m.format
}

func (m *MockTileResponse) getTimestamp() *time.Time {
	return &m.timestamp
}

func (m *MockTileResponse) getSize() int {
	return m.size
}

func (m *MockTileResponse) getCacheable() bool {
	return m.cacheable
}

func (m *MockTileResponse) GetFormatMime() string {
	return tile.TileFormat(m.format).MimeType()
}

// MockCacheManager implements cache.Manager interface for testing
type MockCacheManager struct {
	grid          *geo.TileGrid
	format        string
	requestFormat string
	tileOptions   tile.TileOptions
	tileSource    tile.Source
}

func (m *MockCacheManager) Cleanup() bool                       { return true }
func (m *MockCacheManager) SiteURL() string                     { return "" }
func (m *MockCacheManager) GetSources() []layer.Layer           { return nil }
func (m *MockCacheManager) SetSources(layer []layer.Layer)      {}
func (m *MockCacheManager) GetGrid() *geo.TileGrid              { return m.grid }
func (m *MockCacheManager) GetCache() cache.Cache               { return nil }
func (m *MockCacheManager) SetCache(c cache.Cache)              {}
func (m *MockCacheManager) GetMetaGrid() *geo.MetaGrid          { return nil }
func (m *MockCacheManager) GetTileOptions() tile.TileOptions    { return m.tileOptions }
func (m *MockCacheManager) SetTileOptions(opt tile.TileOptions) {}
func (m *MockCacheManager) GetReprojectSrcSrs() geo.Proj        { return nil }
func (m *MockCacheManager) GetReprojectDstSrs() geo.Proj        { return nil }
func (m *MockCacheManager) GetQueryBuffer() *int                { return nil }
func (m *MockCacheManager) GetFormat() string                   { return m.format }
func (m *MockCacheManager) GetRequestFormat() string            { return m.requestFormat }
func (m *MockCacheManager) SetMinimizeMetaRequests(f bool)      {}
func (m *MockCacheManager) GetMinimizeMetaRequests() bool       { return false }
func (m *MockCacheManager) GetRescaleTiles() int                { return 0 }
func (m *MockCacheManager) LoadTileCoord(tileCoord [3]int, dimensions utils.Dimensions, with_metadata bool) (*cache.Tile, error) {
	tile := cache.NewTile(tileCoord)
	if m.tileSource != nil {
		tile.Source = m.tileSource
	} else {
		// Create a default mock source if none provided
		tile.Source = &MockTileSource{
			buffer:    []byte("mock tile data"),
			cacheable: true,
			options:   m.tileOptions,
			coord:     tileCoord,
		}
	}
	return tile, nil
}
func (m *MockCacheManager) LoadTileCoords(tileCoord [][3]int, dimensions utils.Dimensions, with_metadata bool) (*cache.TileCollection, error) {
	return cache.NewTileCollection(tileCoord), nil
}
func (m *MockCacheManager) RemoveTileCoords(tileCoord [][3]int) error                   { return nil }
func (m *MockCacheManager) StoreTile(tile *cache.Tile) error                            { return nil }
func (m *MockCacheManager) StoreTiles(tiles *cache.TileCollection) error                { return nil }
func (m *MockCacheManager) IsCached(tileCoord [3]int, dimensions utils.Dimensions) bool { return false }
func (m *MockCacheManager) IsStale(tileCoord [3]int, dimensions utils.Dimensions) bool  { return false }
func (m *MockCacheManager) ExpireTimestamp(tile *cache.Tile) *time.Time                 { return nil }
func (m *MockCacheManager) SetExpireTimestamp(t *time.Time)                             {}
func (m *MockCacheManager) ApplyTileFilter(tile *cache.Tile) (*cache.Tile, error)       { return tile, nil }
func (m *MockCacheManager) Creator(dimensions utils.Dimensions) *cache.TileCreator      { return nil }
func (m *MockCacheManager) Lock(ctx context.Context, tile *cache.Tile, run func() error) error {
	return run()
}

// MockTileSource implements tile.Source interface for testing
type MockTileSource struct {
	buffer    []byte
	cacheable bool
	options   tile.TileOptions
	coord     [3]int
}

func (m *MockTileSource) GetType() tile.TileType       { return tile.TILE_IMAGERY }
func (m *MockTileSource) GetSource() interface{}       { return m.buffer }
func (m *MockTileSource) SetSource(source interface{}) { m.buffer = source.([]byte) }
func (m *MockTileSource) GetFileName() string          { return "mock.tile" }
func (m *MockTileSource) GetSize() [2]uint32           { return [2]uint32{256, 256} }
func (m *MockTileSource) GetBuffer(format *tile.TileFormat, opts tile.TileOptions) []byte {
	return m.buffer
}
func (m *MockTileSource) GetTile() interface{} { return m.coord }
func (m *MockTileSource) GetCacheable() *tile.CacheInfo {
	return &tile.CacheInfo{Cacheable: m.cacheable}
}
func (m *MockTileSource) SetCacheable(cacheable *tile.CacheInfo)  { m.cacheable = cacheable.Cacheable }
func (m *MockTileSource) SetTileOptions(options tile.TileOptions) { m.options = options }
func (m *MockTileSource) GetTileOptions() tile.TileOptions        { return m.options }
func (m *MockTileSource) GetGeoReference() *geo.GeoReference      { return nil }
