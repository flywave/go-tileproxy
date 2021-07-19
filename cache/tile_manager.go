package cache

import (
	"time"

	"github.com/flywave/go-tileproxy/filters"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/sources"
)

type Manager interface {
	GetGrid() *geo.TileGrid
	GetCache() Cache
	GetMetaGrid() *geo.MetaGrid
	Cleanup() bool
	SetMinimizeMetaRequests(f bool)
	GetRescaleTiles() int
	LoadTileCoord(tile_coord [3]int, dimensions map[string]interface{}, with_metadata bool) error
	LoadTileCoords(tile_coords [][3]int, dimensions map[string]interface{}, with_metadata bool) error
	RemoveTileCoords(tile_coord [3]int) error
	IsCached(tile_coord [3]int, dimensions map[string]interface{}) bool
	IsStale(tile_coord [3]int, dimensions map[string]interface{}) bool
	ExpireTimestamp(tile *Tile) time.Time
	ApplyTileFilter(tile *Tile) *Tile
	Creator(dimensions map[string]interface{}) *TileCreator
	Lock(tile *Tile)
}

type TileManager struct {
	Manager
	grid                 *geo.TileGrid
	cache                Cache
	identifier           string
	metaGrid             *geo.MetaGrid
	format               string
	imageOpts            *images.ImageOptions
	requestFormat        string
	sources              []sources.TileSource
	minimizeMetaRequests bool
	expireTimestamp      time.Time
	preStoreFilter       []filters.Filter
	rescaleTiles         int
	cacheRescaledTiles   bool
}

func (tm *TileManager) GetGrid() *geo.TileGrid {
	return tm.grid
}

func (tm *TileManager) GetCache() Cache {
	return tm.cache
}

func (tm *TileManager) GetMetaGrid() *geo.MetaGrid {
	return tm.metaGrid
}

func (tm *TileManager) Cleanup() bool {
	if xw, ok := tm.cache.(interface {
		Cleanup() bool
	}); ok {
		return xw.Cleanup()
	}
	return false
}

func (tm *TileManager) SetMinimizeMetaRequests(f bool) {
	tm.minimizeMetaRequests = f
}

func (tm *TileManager) GetRescaleTiles() int {
	return tm.rescaleTiles
}

func (tm *TileManager) LoadTileCoord(tile_coord [3]int, dimensions map[string]interface{}, with_metadata bool) (error, TileCollection) {
	return tm.LoadTileCoords([][3]int{tile_coord}, dimensions, with_metadata)
}

func (tm *TileManager) LoadTileCoords(tile_coords [][3]int, dimensions map[string]interface{}, with_metadata bool) (error, TileCollection) {
	return nil, TileCollection{}
}

func (tm *TileManager) RemoveTileCoords(tile_coord [3]int) error {
	return nil
}

func (tm *TileManager) IsCached(tile_coord [3]int, dimensions map[string]interface{}) bool {
	return false
}

func (tm *TileManager) IsStale(tile_coord [3]int, dimensions map[string]interface{}) bool {
	return false
}

func (tm *TileManager) ExpireTimestamp(tile *Tile) time.Time {
	return tm.expireTimestamp
}

func (tm *TileManager) ApplyTileFilter(tile *Tile) *Tile {
	return nil
}

func (tm *TileManager) Creator(dimensions map[string]interface{}) *TileCreator {
	return nil
}

func (tm *TileManager) Lock(tile *Tile) {

}
