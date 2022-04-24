package cache

import (
	"errors"
	"fmt"
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

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
	reprojectSrc geo.Proj
	reprojectDst geo.Proj
	queryBuffer  *int
}

func NewCacheMapLayer(tm Manager, ext *geo.MapExtent, opts tile.TileOptions, maxTileLimit *int, reprojectSrc geo.Proj, reprojectDst geo.Proj, queryBuffer *int) *CacheMapLayer {
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
		reprojectSrc: reprojectSrc,
		reprojectDst: reprojectDst,
		queryBuffer:  queryBuffer,
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

func bufferedBBox(g *geo.TileGrid, bbox vec2d.Rect, level int, queryBuffer int) vec2d.Rect {
	minx, miny, maxx, maxy := bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1]

	if queryBuffer > 0 {
		res := g.Resolution(level)
		minx -= float64(queryBuffer) * res
		miny -= float64(queryBuffer) * res
		maxx += float64(queryBuffer) * res
		maxy += float64(queryBuffer) * res
	}
	return vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
}

func (r *CacheMapLayer) getSource(query *layer.MapQuery) (tile.Source, error) {
	bbox, srs, dst_srs := query.BBox, query.Srs, query.Srs
	if r.queryBuffer != nil {
		_, level, err := r.grid.GetAffectedBBoxAndLevel(bbox, [2]uint32{r.grid.TileSize[0], r.grid.TileSize[1]}, srs)
		if err != nil {
			return nil, err
		}
		bbox = srs.TransformRectTo(r.grid.Srs, bbox, 16)
		bbox = bufferedBBox(r.grid, bbox, level, *r.queryBuffer)
		bbox = r.grid.Srs.TransformRectTo(srs, bbox, 16)
	}

	if r.reprojectSrc != nil {
		bbox = srs.TransformRectTo(r.reprojectSrc, bbox, 16)
		srs = r.reprojectSrc
		if r.reprojectDst != nil {
			dst_srs = r.reprojectDst
		} else {
			dst_srs = geo.NewProj("EPSG:4326")
		}
	}

	src_bbox, tile_grid, affected_tile_coords, err := r.grid.GetAffectedTiles(bbox, query.Size, dst_srs)
	if err != nil {
		return nil, err
	}

	num_tiles := tile_grid[0] * tile_grid[1]

	if r.maxTileLimit != nil && num_tiles >= *r.maxTileLimit {
		return nil, fmt.Errorf("too many tiles, max_tile_limit: %d, num_tiles: %d", *r.maxTileLimit, num_tiles)
	}

	if query.TiledOnly && num_tiles == 1 {
		if !geo.BBoxEquals(bbox, src_bbox, math.Abs((bbox.Max[0]-bbox.Min[0])/float64(query.Size[0])/10),
			math.Abs((bbox.Max[1]-bbox.Min[1])/float64(query.Size[1])/10)) {
			return nil, errors.New("query does not align to tile boundaries")
		}
	}

	coords := [][3]int{}

	for {
		x, y, z, done := affected_tile_coords.Next()

		coords = append(coords, [3]int{x, y, z})

		if done {
			break
		}
	}

	tile_collection, _ := r.tileManager.LoadTileCoords(coords, nil, query.TiledOnly)

	if tile_collection.Empty() {
		if r.emptySource == nil {
			r.emptySource = GetEmptyTile(query.Size, r.Options)
		}
		return r.emptySource, nil
	}

	if query.TiledOnly {
		if len(tile_collection.tiles) > 1 {
			tile_sources := []tile.Source{}
			for _, t := range tile_collection.tiles {
				tile_sources = append(tile_sources, t.Source)
			}
			return ResampleTiles(tile_sources, query.BBox, query.Srs, tile_grid, r.grid, src_bbox, srs, query.Size, r.Options)
		} else {
			t := tile_collection.GetItem(0)
			tile := t.Source
			tile.SetTileOptions(r.Options)
			tile.SetCacheable(t.GetCacheInfo())
			return tile, nil
		}
	}

	tile_sources := []tile.Source{}
	for _, t := range tile_collection.tiles {
		tile_sources = append(tile_sources, t.Source)
	}
	return ScaleTiles(tile_sources, query.BBox, query.Srs, tile_grid, r.grid, src_bbox, r.Options)
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
				r.emptySource = GetEmptyTile(query.Size, r.Options)
			}
			return r.emptySource, nil
		}
		size, offset, bbox := imagery.BBoxPositionInImage(query.BBox, query.Size, r.Extent.BBoxFor(query.Srs))
		if size[0] == 0 || size[1] == 0 {
			if r.emptySource == nil {
				r.emptySource = GetEmptyTile(query.Size, r.Options)
			}
			return r.emptySource, nil
		}
		src_query := &layer.MapQuery{BBox: bbox, Size: size, Srs: query.Srs, Format: query.Format}
		resp, err := r.getSource(src_query)
		if err != nil {
			if r.emptySource == nil {
				r.emptySource = GetEmptyTile(query.Size, r.Options)
			}
			return r.emptySource, nil
		}
		result = imagery.SubImageSource(resp.(*imagery.ImageSource), query.Size, offset[:], r.Options.(*imagery.ImageOptions), resp.GetCacheable())
	} else {
		result, _ = r.getSource(query)
	}
	return result, nil
}
