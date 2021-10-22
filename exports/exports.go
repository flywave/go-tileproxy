package exports

import (
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/tile"
)

type ExportIO interface {
	GetTileFormat() tile.TileFormat
	StoreTile(t *cache.Tile) error
	StoreTileCollection(ts *cache.TileCollection) error
	Close() error
}
