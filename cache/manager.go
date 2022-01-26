package cache

import (
	"context"
	"time"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type Manager interface {
	GetSources() []layer.Layer
	SetSources(layer []layer.Layer)
	GetGrid() *geo.TileGrid
	GetCache() Cache
	SetCache(c Cache)
	GetMetaGrid() *geo.MetaGrid
	GetTileOptions() tile.TileOptions
	SetTileOptions(opt tile.TileOptions)
	GetReprojectSrcSrs() geo.Proj
	GetReprojectDstSrs() geo.Proj
	GetQueryBuffer() *int
	Cleanup() bool
	GetFormat() string
	GetRequestFormat() string
	SetMinimizeMetaRequests(f bool)
	GetMinimizeMetaRequests() bool
	GetRescaleTiles() int
	LoadTileCoord(tileCoord [3]int, dimensions utils.Dimensions, with_metadata bool) (*Tile, error)
	LoadTileCoords(tileCoord [][3]int, dimensions utils.Dimensions, with_metadata bool) (*TileCollection, error)
	RemoveTileCoords(tileCoord [][3]int) error
	StoreTile(tile *Tile) error
	StoreTiles(tiles *TileCollection) error
	IsCached(tileCoord [3]int, dimensions utils.Dimensions) bool
	IsStale(tileCoord [3]int, dimensions utils.Dimensions) bool
	ExpireTimestamp(tile *Tile) *time.Time
	SetExpireTimestamp(t *time.Time)
	ApplyTileFilter(tile *Tile) (*Tile, error)
	Creator(dimensions utils.Dimensions) *TileCreator
	Lock(ctx context.Context, tile *Tile, run func() error) error
}
