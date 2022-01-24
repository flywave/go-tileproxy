package cache

import (
	"errors"
	"fmt"
	"math"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type CacheMapLayer struct {
	layer.MapLayer
	tileManager  Manager
	grid         *geo.TileGrid
	maxTileLimit *int
	emptySource  tile.Source
}

func NewCacheMapLayer(tm Manager, ext *geo.MapExtent, opts tile.TileOptions, maxTileLimit *int) *CacheMapLayer {
	if ext == nil {
		ext = geo.MapExtentFromGrid(tm.GetGrid())
	}
	ret := &CacheMapLayer{
		MapLayer: layer.MapLayer{
			SupportMetaTiles: true,
			Extent:           ext,
			Options:          opts,
		},
		tileManager:  tm,
		grid:         tm.GetGrid(),
		maxTileLimit: maxTileLimit,
		emptySource:  nil,
	}
	ret.ResRange = nil
	if tm.GetRescaleTiles() == -1 {
		ret.ResRange = layer.MergeLayerResRanges(tm.GetSources())
	}
	return ret
}

func (r *CacheMapLayer) checkTiled(query *layer.MapQuery) error {
	if string(query.Format) != r.tileManager.GetFormat() {
		return fmt.Errorf("invalid tile format, use %s", r.tileManager.GetFormat())
	}
	if query.Size[0] != r.grid.TileSize[0] || query.Size[1] != r.grid.TileSize[1] {
		return fmt.Errorf("invalid tile size (use %dx%d)", r.grid.TileSize[0], r.grid.TileSize[1])
	}
	return nil
}

func (r *CacheMapLayer) getSource(query *layer.MapQuery) (tile.Source, error) {
	src_bbox, tile_grid, affected_tile_coords, err := r.grid.GetAffectedTiles(query.BBox, query.Size, query.Srs)
	if err != nil {
		return nil, err
	}

	num_tiles := tile_grid[0] * tile_grid[1]

	if r.maxTileLimit != nil && num_tiles >= *r.maxTileLimit {
		return nil, fmt.Errorf("too many tiles, max_tile_limit: %d, num_tiles: %d", *r.maxTileLimit, num_tiles)
	}

	if query.TiledOnly {
		if num_tiles > 1 {
			return nil, errors.New("not a single tile")
		}
		bbox := query.BBox
		if !geo.BBoxEquals(bbox, src_bbox, math.Abs((bbox.Max[0]-bbox.Min[0])/float64(query.Size[0])/10),
			math.Abs((bbox.Max[1]-bbox.Min[1])/float64(query.Size[1])/10)) {
			return nil, errors.New("query does not align to tile boundaries")
		}
	}

	coords := [][3]int{}
	x, y, zoom, done := affected_tile_coords.Next()
	for !done {
		coords = append(coords, [3]int{x, y, zoom})
		x, y, zoom, done = affected_tile_coords.Next()
	}
	if len(coords) == 0 {
		coords = append(coords, [3]int{x, y, zoom})
	}

	tile_collection, _ := r.tileManager.LoadTileCoords(coords, nil, query.TiledOnly)

	if tile_collection.Empty() {
		if r.emptySource == nil {
			r.emptySource = GetEmptyTile(query.Size, r.tileManager.GetTileOptions())
		}
		return r.emptySource, nil
	}

	if query.TiledOnly {
		if len(tile_collection.tiles) > 1 {
			tile_sources := []tile.Source{}
			for _, t := range tile_collection.tiles {
				tile_sources = append(tile_sources, t.Source)
			}
			return ResampleTiles(tile_sources, query.BBox, query.Srs, tile_grid, r.grid, src_bbox, query.Size, r.tileManager.GetTileOptions()), nil
		} else {
			t := tile_collection.GetItem(0)
			tile := t.Source
			tile.SetTileOptions(r.tileManager.GetTileOptions())
			tile.SetCacheable(t.GetCacheInfo())
			return tile, nil
		}
	}

	tile_sources := []tile.Source{}
	for _, t := range tile_collection.tiles {
		tile_sources = append(tile_sources, t.Source)
	}
	return ScaleTiles(tile_sources, query.BBox, query.Srs, tile_grid, r.grid, src_bbox, r.tileManager.GetTileOptions()), nil
}

func (r *CacheMapLayer) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if err := r.CheckResRange(query); err != nil {
		return nil, errors.New("res range error")
	}

	if query.TiledOnly {
		r.checkTiled(query)
	}

	queryExtent := &geo.MapExtent{BBox: query.BBox, Srs: query.Srs}
	var result tile.Source
	if !query.TiledOnly && r.Extent != nil && !r.Extent.Contains(queryExtent) {
		if !r.Extent.Intersects(queryExtent) {
			if r.emptySource == nil {
				r.emptySource = GetEmptyTile(query.Size, r.tileManager.GetTileOptions())
			}
			return r.emptySource, nil
		}
		size, offset, bbox := imagery.BBoxPositionInImage(query.BBox, query.Size, r.Extent.BBoxFor(query.Srs))
		if size[0] == 0 || size[1] == 0 {
			if r.emptySource == nil {
				r.emptySource = GetEmptyTile(query.Size, r.tileManager.GetTileOptions())
			}
			return r.emptySource, nil
		}
		src_query := &layer.MapQuery{BBox: bbox, Size: size, Srs: query.Srs, Format: query.Format}
		resp, err := r.getSource(src_query)
		if err != nil {
			if r.emptySource == nil {
				r.emptySource = GetEmptyTile(query.Size, r.tileManager.GetTileOptions())
			}
			return r.emptySource, nil
		}
		result = imagery.SubImageSource(resp.(*imagery.ImageSource), query.Size, offset[:], r.Options.(*imagery.ImageOptions), resp.GetCacheable())
	} else {
		result, _ = r.getSource(query)
	}
	return result, nil
}
