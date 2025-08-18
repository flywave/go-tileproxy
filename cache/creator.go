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

func NewTileCreator(m Manager, dimensions utils.Dimensions, merger tile.Merger, bulkMetaTiles bool) *TileCreator {
	return &TileCreator{
		manager:       m,
		sources:       m.GetSources(),
		grid:          m.GetGrid(),
		cache:         m.GetCache(),
		metaGrid:      m.GetMetaGrid(),
		dimensions:    dimensions,
		tileMerger:    merger,
		bulkMetaTiles: bulkMetaTiles,
	}
}

func (c *TileCreator) IsCached(tile [3]int) bool {
	return c.manager.IsCached(tile, nil)
}

func (c *TileCreator) CreateTiles(tiles []*Tile) ([]*Tile, error) {
	if c.sources == nil {
		return []*Tile{}, errors.New("source is nil")
	}
	var created_tiles []*Tile
	var err error
	if c.metaGrid == nil {
		created_tiles, err = c.createSingleTiles(tiles)
		if err != nil {
			return nil, err
		}
	} else if c.manager.GetMinimizeMetaRequests() && len(tiles) > 1 {
		coords := [][3]int{}
		for i := range tiles {
			coords = append(coords, tiles[i].Coord)
		}
		meta_tile := c.metaGrid.MinimalMetaTile(coords)
		ctiles, err := c.createMetaTile(meta_tile)
		if err != nil {
			return nil, err
		}
		created_tiles = append(created_tiles, ctiles...)
	} else if c.metaGrid != nil && len(tiles) == 1 {
		meta_tile := c.metaGrid.GetMetaTile(tiles[0].Coord)
		ctiles, err := c.createSingleMetaTile(meta_tile)
		if err != nil {
			return nil, err
		}
		created_tiles = ctiles

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
		tiles, err := c.createMetaTiles(meta_tiles)
		if err != nil {
			return nil, err
		}
		created_tiles = tiles
	}

	return created_tiles, nil
}

func (c *TileCreator) createSingleTiles(tiles []*Tile) ([]*Tile, error) {
	created_tiles := []*Tile{}
	for _, tile := range tiles {
		t, err := c.createSingleTile(tile)
		if err != nil {
			return nil, err
		}
		created_tiles = append(created_tiles, t)
	}
	return created_tiles, nil
}

func (c *TileCreator) createSingleTile(t *Tile) (*Tile, error) {
	tile_bbox := c.grid.TileBBox(t.Coord, false)

	query := &layer.MapQuery{
		TileId:     t.Coord,
		BBox:       tile_bbox,
		Size:       [2]uint32{c.grid.TileSize[0], c.grid.TileSize[1]},
		Srs:        c.grid.Srs,
		Format:     tile.TileFormat(c.manager.GetRequestFormat()),
		Dimensions: c.dimensions,
		MetaSize:   [2]uint32{1, 1},
	}

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.manager.Lock(lockCtx, t, func() error {
		if !c.IsCached(t.Coord) {
			source, err := c.querySources(query)
			if err != nil {
				return err
			}
			source.SetTileOptions(c.manager.GetTileOptions())
			t.Source = source
			t.SetCacheInfo(source.GetCacheable())
			t, err = c.manager.ApplyTileFilter(t)
			if err != nil {
				return err
			}
			if source.GetCacheable() != nil {
				return c.cache.StoreTile(t)
			}
		} else {
			return c.cache.LoadTile(t, false)
		}
		return nil
	})
	return t, err
}

func (c *TileCreator) querySources(query *layer.MapQuery) (tile.Source, error) {
	layers := []tile.Source{}
	for i := range c.sources {
		if c.sources[i].GetCoverage() == nil ||
			(c.sources[i].GetCoverage().Intersects(query.BBox, query.Srs)) {
			img, err := c.sources[i].GetMap(query)
			if err == nil && img != nil {
				layers = append(layers, img)
			}
		}
	}
	if len(layers) == 0 {
		return nil, errors.New("no source create")
	}

	if len(layers) == 1 {
		return layers[0], nil
	}
	ret, err := MergeTiles(layers, c.manager.GetTileOptions(), query, c.tileMerger)

	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *TileCreator) createMetaTiles(metaTiles []*geo.MetaTile) ([]*Tile, error) {
	if c.bulkMetaTiles {
		created_tiles := []*Tile{}
		for _, meta_tile := range metaTiles {
			tiles, err := c.createBulkMetaTile(meta_tile)
			if err != nil {
				return nil, err
			}
			created_tiles = append(created_tiles, tiles...)
		}
		return created_tiles, nil
	}

	created_tiles := []*Tile{}
	for _, meta_tile := range metaTiles {
		tiles, err := c.createMetaTile(meta_tile)
		if err != nil {
			return nil, err
		}
		created_tiles = append(created_tiles, tiles...)
	}
	return created_tiles, nil
}

func (c *TileCreator) createMetaTile(metaTile *geo.MetaTile) ([]*Tile, error) {
	tile_size := c.grid.TileSize
	query := &layer.MapQuery{
		TileId:     metaTile.GetMainTileCoord(),
		BBox:       metaTile.GetBBox(),
		Size:       metaTile.GetSize(),
		Srs:        c.grid.Srs,
		Format:     tile.TileFormat(c.manager.GetRequestFormat()),
		Dimensions: c.dimensions,
		MetaSize:   c.metaGrid.MetaSize,
	}

	main_tile := NewTile(metaTile.GetMainTileCoord())
	var splittedTiles *TileCollection

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.manager.Lock(lockCtx, main_tile, func() error {
		flag := true

		for _, t := range metaTile.GetTiles() {
			if !c.IsCached(t) {
				flag = false
			}
		}

		if !flag {
			metaTileImage, err := c.querySources(query)
			if err != nil {
				return err
			}

			splittedTiles, err = SplitTiles(metaTileImage, metaTile.GetTilePattern(), [2]uint32{tile_size[0], tile_size[1]}, c.manager.GetTileOptions())
			if err != nil {
				return err
			}
			for i, t := range splittedTiles.tiles {
				tt, err := c.manager.ApplyTileFilter(t)
				if err != nil {
					return err
				}
				splittedTiles.UpdateItem(i, tt)
			}
			if metaTileImage.GetCacheable() != nil {
				c.cache.StoreTiles(splittedTiles)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	tiles := NewTileCollection(nil)
	for _, coord := range metaTile.GetTiles() {
		tiles.SetItem(NewTile(coord))
	}
	c.cache.LoadTiles(tiles, false)
	return tiles.tiles, nil
}

func (c *TileCreator) createSingleMetaTile(metaTile *geo.MetaTile) ([]*Tile, error) {
	tile_size := c.grid.TileSize
	query := &layer.MapQuery{
		TileId:     metaTile.GetMainTileCoord(),
		BBox:       metaTile.GetBBox(),
		Size:       [2]uint32{tile_size[0], tile_size[1]},
		Srs:        c.grid.Srs,
		Format:     tile.TileFormat(c.manager.GetRequestFormat()),
		Dimensions: c.dimensions,
		MetaSize:   c.metaGrid.MetaSize,
	}

	main_tile := NewTile(metaTile.GetMainTileCoord())

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.manager.Lock(lockCtx, main_tile, func() error {
		if !c.IsCached(main_tile.Coord) {
			metaTileImage, err := c.querySources(query)
			if err != nil {
				return err
			}
			main_tile.SetCacheInfo(metaTileImage.GetCacheable())
			main_tile.Source = metaTileImage
			if metaTileImage.GetCacheable() != nil {
				c.cache.StoreTile(main_tile)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return []*Tile{main_tile}, nil
}

func (c *TileCreator) queryTile(coord [3]int, tile_size []uint32) (*Tile, error) {
	query := &layer.MapQuery{
		BBox:       c.grid.TileBBox(coord, false),
		Size:       [2]uint32{tile_size[0], tile_size[1]},
		Srs:        c.grid.Srs,
		Format:     tile.TileFormat(c.manager.GetRequestFormat()),
		Dimensions: c.dimensions,
	}

	tile_data, err := c.querySources(query)
	if err != nil {
		return nil, err
	}

	tile := NewTile(coord)
	tile.SetCacheInfo(tile_data.GetCacheable())
	tile.Source = tile_data
	tile, err = c.manager.ApplyTileFilter(tile)
	if err != nil {
		return nil, err
	}
	return tile, nil
}

func (c *TileCreator) createBulkMetaTile(meta_tile *geo.MetaTile) ([]*Tile, error) {
	tile_size := c.grid.TileSize
	main_tile := NewTile(meta_tile.GetMainTileCoord())
	var tiles []*Tile

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.manager.Lock(lockCtx, main_tile, func() error {
		flag := true

		for _, t := range meta_tile.GetTiles() {
			if !c.IsCached(t) {
				flag = false
			}
		}

		if !flag {
			for _, coord := range meta_tile.GetTiles() {
				tile, err := c.queryTile(coord, tile_size)
				if err != nil {
					return err
				}
				tiles = append(tiles, tile)
			}

			cacheTiles := NewTileCollection(nil)
			for _, t := range tiles {
				if t.Cacheable {
					cacheTiles.SetItem(t)
				}
			}

			err := c.cache.StoreTiles(cacheTiles)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	ctiles := NewTileCollection(nil)
	for _, coord := range meta_tile.GetTiles() {
		ctiles.SetItem(NewTile(coord))
	}
	c.cache.LoadTiles(ctiles, false)

	return ctiles.tiles, nil
}
