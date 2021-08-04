package cache

import (
	"context"
	"errors"
	"time"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"

	mapset "github.com/deckarep/golang-set"
)

type TileCreator struct {
	Cache         Cache
	Sources       []layer.Layer
	Grid          *geo.TileGrid
	MetaGrid      *geo.MetaGrid
	BulkMetaTiles bool
	Manager       Manager
	Dimensions    utils.Dimensions
	TileMerger    tile.Merger
}

func NewTileCreator(m Manager, dimensions utils.Dimensions, merger tile.Merger, bulk_meta_tiles bool) *TileCreator {
	return &TileCreator{Manager: m, Sources: m.GetSources(), Dimensions: dimensions, TileMerger: merger, BulkMetaTiles: bulk_meta_tiles}
}

func (c *TileCreator) IsCached(tile [3]int) bool {
	return c.Manager.IsCached(tile, nil)
}

func (c *TileCreator) CreateTiles(tiles []*Tile) []*Tile {
	if c.Sources == nil {
		return []*Tile{}
	}
	var created_tiles []*Tile
	if c.MetaGrid == nil {
		created_tiles = c.createSingleTiles(tiles)
	} else if c.Manager.GetMinimizeMetaRequests() && len(tiles) > 1 {
		coords := [][3]int{}
		for i := range tiles {
			coords = append(coords, tiles[i].Coord)
		}
		meta_tile := c.MetaGrid.MinimalMetaTile(coords)
		created_tiles = append(created_tiles, c.createMetaTile(meta_tile)...)
	} else {
		meta_tiles := []*geo.MetaTile{}
		meta_bboxes := mapset.NewSet()
		for _, tile := range tiles {
			meta_tile := c.MetaGrid.GetMetaTile(tile.Coord)
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
	tile_bbox := c.Grid.TileBBox(t.Coord, false)
	query := &layer.MapQuery{BBox: tile_bbox, Size: [2]uint32{c.Grid.TileSize[0], c.Grid.TileSize[1]}, Srs: c.Grid.Srs,
		Format: tile.TileFormat(c.Manager.GetRequestFormat()), Dimensions: c.Dimensions}

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.Manager.Lock(lockCtx, t, func() error {
		if !c.IsCached(t.Coord) {
			source, err := c.querySources(query)
			if source == nil || err != nil {
				return nil
			}
			source.SetTileOptions(c.Manager.GetTileOptions())
			t.Source = source
			t.SetCacheInfo(source.GetCacheable())
			t = c.Manager.ApplyTileFilter(t)
			if source.GetCacheable() != nil {
				return c.Cache.StoreTile(t)
			}
		} else {
			return c.Cache.LoadTile(t, false)
		}
		return nil
	})
	return t
}

func (c *TileCreator) querySources(query *layer.MapQuery) (tile.Source, error) {
	if len(c.Sources) == 1 &&
		c.TileMerger == nil && !(c.Sources[0].GetCoverage() != nil &&
		c.Sources[0].GetCoverage().IsClip() &&
		c.Sources[0].GetCoverage().Intersects(query.BBox, query.Srs)) {
		return c.Sources[0].GetMap(query)
	}

	layers := []tile.Source{}

	for i := range c.Sources {
		img, err := c.Sources[i].GetMap(query)

		if err == nil {
			layers = append(layers, img)
		}
	}
	imageOptions, ok := c.Manager.GetTileOptions().(*images.ImageOptions)
	if ok && imageOptions != nil {
		return images.MergeImages(layers, c.Manager.GetTileOptions().(*images.ImageOptions), query.Size, query.BBox, query.Srs, c.TileMerger), nil
	}
	return nil, errors.New("error")
}

func (c *TileCreator) createMetaTiles(meta_tiles []*geo.MetaTile) []*Tile {
	if c.BulkMetaTiles {
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

func splitMetaTiles(meta_tile tile.Source, tiles []geo.TilePattern, tile_size [2]uint32, image_opts *images.ImageOptions) *TileCollection {
	splitter := images.NewTileSplitter(meta_tile, image_opts)
	split_tiles := &TileCollection{}
	for _, tile := range tiles {
		tile_coord, crop_coord := tile.Tiles, tile.Sizes
		if tile_coord[0] < 0 || tile_coord[1] < 0 || tile_coord[2] < 0 {
			continue
		}
		data := splitter.GetTile(crop_coord, tile_size)
		new_tile := NewTile(tile_coord)
		new_tile.SetCacheInfo(meta_tile.GetCacheable())
		new_tile.Source = data
		split_tiles.SetItem(new_tile)
	}
	return split_tiles
}

func (c *TileCreator) createMetaTile(meta_tile *geo.MetaTile) []*Tile {
	tile_size := c.Grid.TileSize
	query := &layer.MapQuery{BBox: meta_tile.GetBBox(), Size: meta_tile.GetSize(), Srs: c.Grid.Srs, Format: tile.TileFormat(c.Manager.GetRequestFormat()), Dimensions: c.Dimensions}
	main_tile := NewTile(meta_tile.GetMainTileCoord())
	var splitted_tiles *TileCollection

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.Manager.Lock(lockCtx, main_tile, func() error {
		flag := true

		for _, t := range meta_tile.GetTiles() {
			if !c.IsCached(t) {
				flag = false
			}
		}

		if flag {
			meta_tile_image, err := c.querySources(query)
			if meta_tile_image == nil || err != nil {
				return nil
			}
			splitted_tiles = splitMetaTiles(meta_tile_image, meta_tile.GetTilePattern(),
				[2]uint32{tile_size[0], tile_size[1]}, c.Manager.GetTileOptions().(*images.ImageOptions))
			for i, t := range splitted_tiles.tiles {
				splitted_tiles.UpdateItem(i, c.Manager.ApplyTileFilter(t))
			}
			if meta_tile_image.GetCacheable() != nil {
				c.Cache.StoreTiles(splitted_tiles)
			}
		}
		return nil
	})

	tiles := &TileCollection{}
	for _, coord := range meta_tile.GetTiles() {
		tiles.SetItem(NewTile(coord))
	}
	c.Cache.LoadTiles(tiles, false)
	return tiles.tiles
}

func (c *TileCreator) queryTile(coord [3]int, tile_size []uint32) *Tile {
	query := &layer.MapQuery{BBox: c.Grid.TileBBox(coord, false), Size: [2]uint32{tile_size[0], tile_size[1]}, Srs: c.Grid.Srs, Format: tile.TileFormat(c.Manager.GetRequestFormat()),
		Dimensions: c.Dimensions}
	tile_data, err := c.querySources(query)
	if tile_data == nil || err != nil {
		return nil
	}

	tile := NewTile(coord)
	tile.SetCacheInfo(tile_data.GetCacheable())
	tile.Source = tile_data
	tile = c.Manager.ApplyTileFilter(tile)
	return tile
}

func (c *TileCreator) createBulkMetaTile(meta_tile *geo.MetaTile) []*Tile {
	tile_size := c.Grid.TileSize
	main_tile := NewTile(meta_tile.GetMainTileCoord())
	var tiles []*Tile

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.Manager.Lock(lockCtx, main_tile, func() error {
		flag := true

		for _, t := range meta_tile.GetTiles() {
			if !c.IsCached(t) {
				flag = false
			}
		}

		if flag {
			for _, coord := range meta_tile.GetTiles() {
				tile := c.queryTile(coord, tile_size)
				tiles = append(tiles, tile)
			}

			cacheTiles := &TileCollection{}
			for _, t := range tiles {
				if t.Cacheable {
					cacheTiles.SetItem(t)
				}
			}

			c.Cache.StoreTiles(cacheTiles)
			return nil
		}
		return nil
	})

	if tiles != nil {
		return tiles
	}

	ctiles := &TileCollection{}
	for _, coord := range meta_tile.GetTiles() {
		ctiles.SetItem(NewTile(coord))
	}
	c.Cache.LoadTiles(ctiles, false)

	return ctiles.tiles
}
