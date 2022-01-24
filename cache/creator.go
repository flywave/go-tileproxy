package cache

import (
	"context"
	"errors"
	"time"

	mapset "github.com/deckarep/golang-set"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

type TileCreator struct {
	cache         Cache
	sources       []layer.Layer
	grid          *geo.TileGrid
	metaGrid      *geo.MetaGrid
	bulkMetaTiles bool
	manager       Manager
	dimensions    utils.Dimensions
	tileMerger    tile.Merger
}

func NewTileCreator(m Manager, dimensions utils.Dimensions, merger tile.Merger, bulk_meta_tiles bool) *TileCreator {
	return &TileCreator{
		manager:       m,
		sources:       m.GetSources(),
		grid:          m.GetGrid(),
		cache:         m.GetCache(),
		metaGrid:      m.GetMetaGrid(),
		dimensions:    dimensions,
		tileMerger:    merger,
		bulkMetaTiles: bulk_meta_tiles,
	}
}

func (c *TileCreator) IsCached(tile [3]int) bool {
	return c.manager.IsCached(tile, nil)
}

func (c *TileCreator) CreateTiles(tiles []*Tile) []*Tile {
	if c.sources == nil {
		return []*Tile{}
	}
	var created_tiles []*Tile
	if c.metaGrid == nil {
		created_tiles = c.createSingleTiles(tiles)
	} else if c.manager.GetMinimizeMetaRequests() && len(tiles) > 1 {
		coords := [][3]int{}
		for i := range tiles {
			coords = append(coords, tiles[i].Coord)
		}
		meta_tile := c.metaGrid.MinimalMetaTile(coords)
		created_tiles = append(created_tiles, c.createMetaTile(meta_tile)...)
	} else {
		meta_tiles := []*geo.MetaTile{}
		meta_bboxes := mapset.NewSet()
		for _, tile := range tiles {
			meta_tile := c.metaGrid.GetMetaTile(tile.Coord)
			if !meta_bboxes.Contains(meta_tile.GetBBox()) {
				meta_tiles = append(meta_tiles, meta_tile)
				meta_bboxes.Add(meta_tile.GetBBox())
			}
		}
		created_tiles = c.createMetaTiles(meta_tiles)
	}

	return created_tiles
}

func (c *TileCreator) createSingleTiles(tiles []*Tile) []*Tile {
	created_tiles := []*Tile{}
	for _, tile := range tiles {
		created_tiles = append(created_tiles, c.createSingleTile(tile))
	}
	return created_tiles
}

func (c *TileCreator) createSingleTile(t *Tile) *Tile {
	tile_bbox := c.grid.TileBBox(t.Coord, false)
	query := &layer.MapQuery{
		BBox:       tile_bbox,
		Size:       [2]uint32{c.grid.TileSize[0], c.grid.TileSize[1]},
		Srs:        c.grid.Srs,
		Format:     tile.TileFormat(c.manager.GetRequestFormat()),
		Dimensions: c.dimensions,
	}

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.manager.Lock(lockCtx, t, func() error {
		if !c.IsCached(t.Coord) {
			source, err := c.querySources(query)
			if source == nil || err != nil {
				return nil
			}
			source.SetTileOptions(c.manager.GetTileOptions())
			t.Source = source
			t.SetCacheInfo(source.GetCacheable())
			t = c.manager.ApplyTileFilter(t)
			if source.GetCacheable() != nil {
				return c.cache.StoreTile(t)
			}
		} else {
			return c.cache.LoadTile(t, false)
		}
		return nil
	})
	return t
}

func (c *TileCreator) querySources(query *layer.MapQuery) (tile.Source, error) {
	if len(c.sources) == 1 &&
		c.tileMerger == nil && !(c.sources[0].GetCoverage() != nil &&
		c.sources[0].GetCoverage().IsClip() &&
		c.sources[0].GetCoverage().Intersects(query.BBox, query.Srs)) {
		return c.sources[0].GetMap(query)
	}

	layers := []tile.Source{}

	for i := range c.sources {
		img, err := c.sources[i].GetMap(query)

		if err == nil {
			layers = append(layers, img)
		}
	}
	ret := MergeTiles(layers, c.manager.GetTileOptions(), query.Size, query.BBox, query.Srs, c.tileMerger)

	if ret == nil {
		return nil, errors.New("no blend")
	}
	return ret, nil
}

func (c *TileCreator) createMetaTiles(meta_tiles []*geo.MetaTile) []*Tile {
	if c.bulkMetaTiles {
		created_tiles := []*Tile{}
		for _, meta_tile := range meta_tiles {
			created_tiles = append(created_tiles, c.createBulkMetaTile(meta_tile)...)
		}
		return created_tiles
	}

	created_tiles := []*Tile{}
	for _, meta_tile := range meta_tiles {
		created_tiles = append(created_tiles, c.createMetaTile(meta_tile)...)
	}
	return created_tiles
}

func (c *TileCreator) createMetaTile(meta_tile *geo.MetaTile) []*Tile {
	tile_size := c.grid.TileSize
	query := &layer.MapQuery{
		BBox:       meta_tile.GetBBox(),
		Size:       meta_tile.GetSize(),
		Srs:        c.grid.Srs,
		Format:     tile.TileFormat(c.manager.GetRequestFormat()),
		Dimensions: c.dimensions,
	}
	main_tile := NewTile(meta_tile.GetMainTileCoord())
	var splittedTiles *TileCollection

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.manager.Lock(lockCtx, main_tile, func() error {
		flag := true

		for _, t := range meta_tile.GetTiles() {
			if !c.IsCached(t) {
				flag = false
			}
		}

		if !flag {
			metaTileImage, err := c.querySources(query)
			if metaTileImage == nil || err != nil {
				return nil
			}
			splittedTiles = SplitTiles(metaTileImage, meta_tile.GetTilePattern(), [2]uint32{tile_size[0], tile_size[1]}, c.manager.GetTileOptions())
			for i, t := range splittedTiles.tiles {
				splittedTiles.UpdateItem(i, c.manager.ApplyTileFilter(t))
			}
			if metaTileImage.GetCacheable() != nil {
				c.cache.StoreTiles(splittedTiles)
			}
		}
		return nil
	})

	tiles := NewTileCollection(nil)
	for _, coord := range meta_tile.GetTiles() {
		tiles.SetItem(NewTile(coord))
	}
	c.cache.LoadTiles(tiles, false)
	return tiles.tiles
}

func (c *TileCreator) queryTile(coord [3]int, tile_size []uint32) *Tile {
	query := &layer.MapQuery{
		BBox:       c.grid.TileBBox(coord, false),
		Size:       [2]uint32{tile_size[0], tile_size[1]},
		Srs:        c.grid.Srs,
		Format:     tile.TileFormat(c.manager.GetRequestFormat()),
		Dimensions: c.dimensions,
	}
	tile_data, err := c.querySources(query)
	if tile_data == nil || err != nil {
		return nil
	}

	tile := NewTile(coord)
	tile.SetCacheInfo(tile_data.GetCacheable())
	tile.Source = tile_data
	tile = c.manager.ApplyTileFilter(tile)
	return tile
}

func (c *TileCreator) createBulkMetaTile(meta_tile *geo.MetaTile) []*Tile {
	tile_size := c.grid.TileSize
	main_tile := NewTile(meta_tile.GetMainTileCoord())
	var tiles []*Tile

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.manager.Lock(lockCtx, main_tile, func() error {
		flag := true

		for _, t := range meta_tile.GetTiles() {
			if !c.IsCached(t) {
				flag = false
			}
		}

		if !flag {
			for _, coord := range meta_tile.GetTiles() {
				tile := c.queryTile(coord, tile_size)
				tiles = append(tiles, tile)
			}

			cacheTiles := NewTileCollection(nil)
			for _, t := range tiles {
				if t.Cacheable {
					cacheTiles.SetItem(t)
				}
			}

			c.cache.StoreTiles(cacheTiles)
			return nil
		}
		return nil
	})

	if tiles != nil {
		return tiles
	}

	ctiles := NewTileCollection(nil)
	for _, coord := range meta_tile.GetTiles() {
		ctiles.SetItem(NewTile(coord))
	}
	c.cache.LoadTiles(ctiles, false)

	return ctiles.tiles
}
