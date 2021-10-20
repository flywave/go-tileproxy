package exports

import "github.com/flywave/go-tileproxy/cache"

type ExportIO interface {
	StoreTile(t *cache.Tile) error
	StoreTileCollection(ts *cache.TileCollection) error
}
