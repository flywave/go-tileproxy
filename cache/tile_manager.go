package cache

import (
	"time"

	"github.com/flywave/go-tileproxy/geo"
)

type TileManager interface {
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
