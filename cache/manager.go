package cache

import (
	"context"
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
)

type Manager interface {
	GetSources() []layer.Layer
	GetGrid() *geo.TileGrid
	GetCache() Cache
	GetMetaGrid() *geo.MetaGrid
	GetImageOptions() *images.ImageOptions
	Cleanup() bool
	GetFormat() string
	GetRequestFormat() string
	SetMinimizeMetaRequests(f bool)
	GetMinimizeMetaRequests() bool
	GetRescaleTiles() int
	LoadTileCoord(tile_coord [3]int, dimensions map[string]string, with_metadata bool) (error, *Tile)
	LoadTileCoords(tile_coords [][3]int, dimensions map[string]string, with_metadata bool) (error, *TileCollection)
	RemoveTileCoords(tile_coord [][3]int) error
	IsCached(tile_coord [3]int, dimensions map[string]string) bool
	IsStale(tile_coord [3]int, dimensions map[string]string) bool
	ExpireTimestamp(tile *Tile) *time.Time
	ApplyTileFilter(tile *Tile) *Tile
	Creator(dimensions map[string]string) *TileCreator
	Lock(ctx context.Context, tile *Tile, run func() error) error
}
