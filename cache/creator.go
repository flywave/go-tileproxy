package cache

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
)

type TileCreator struct {
	Cache         Cache
	Sources       []sources.TileSource
	Grid          geo.TileGrid
	MetaGrid      geo.MetaGrid
	BulkMetaTiles bool
	Manager       TileManager
	Dimensions    map[string]interface{}
	ImageMerger   images.Merger
}

func (c *TileCreator) IsCached(tile *Tile) bool {
	return false
}

func (c *TileCreator) CreateTiles(tiles []*Tile) []*Tile {
	return nil
}

func (c *TileCreator) createSingleTiles(tiles []*Tile) {

}

func (c *TileCreator) querySources(query layer.Query) {

}

func (c *TileCreator) createMetaTiles(meta_tiles []*Tile) []*Tile {
	return nil
}

func (c *TileCreator) createMetaTile(meta_tiles *Tile) {

}

func (c *TileCreator) createBulkMetaTile(meta_tile *Tile) {

}
