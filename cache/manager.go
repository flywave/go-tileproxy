package cache

import (
	"context"
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type Manager interface {
	GetSources() []layer.Layer
	GetGrid() *geo.TileGrid
	GetCache() Cache
	GetMetaGrid() *geo.MetaGrid
	GetTileOptions() tile.TileOptions
	Cleanup() bool
	GetFormat() string
	GetRequestFormat() string
	SetMinimizeMetaRequests(f bool)
	GetMinimizeMetaRequests() bool
	GetRescaleTiles() int
	LoadTileCoord(tile_coord [3]int, dimensions utils.Dimensions, with_metadata bool) (*Tile, error)
	LoadTileCoords(tile_coords [][3]int, dimensions utils.Dimensions, with_metadata bool) (*TileCollection, error)
	RemoveTileCoords(tile_coord [][3]int) error
	IsCached(tile_coord [3]int, dimensions utils.Dimensions) bool
	IsStale(tile_coord [3]int, dimensions utils.Dimensions) bool
	ExpireTimestamp(tile *Tile) *time.Time
	ApplyTileFilter(tile *Tile) *Tile
	Creator(dimensions utils.Dimensions) *TileCreator
	Lock(ctx context.Context, tile *Tile, run func() error) error
}
