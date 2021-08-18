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
	LoadTileCoord(tileCoord [3]int, dimensions utils.Dimensions, with_metadata bool) (*Tile, error)
	LoadTileCoords(tileCoord [][3]int, dimensions utils.Dimensions, with_metadata bool) (*TileCollection, error)
	RemoveTileCoords(tileCoord [][3]int) error
	IsCached(tileCoord [3]int, dimensions utils.Dimensions) bool
	IsStale(tileCoord [3]int, dimensions utils.Dimensions) bool
	ExpireTimestamp(tile *Tile) *time.Time
	ApplyTileFilter(tile *Tile) *Tile
	Creator(dimensions utils.Dimensions) *TileCreator
	Lock(ctx context.Context, tile *Tile, run func() error) error
}
