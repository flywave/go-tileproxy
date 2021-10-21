package exports

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type ExportIO interface {
	GetTileFormat() *tile.TileFormat
	SetGrid(g geo.Grid)
	StoreTile(t *cache.Tile) error
	StoreTileCollection(ts *cache.TileCollection) error
}
