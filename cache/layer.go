package cache

import (
	"errors"
	"fmt"
	"math"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/layer"
)

type CacheMapLayer struct {
	layer.MapLayer
	tileManager  Manager
	grid         *geo.TileGrid
	maxTileLimit *int
}

func NewCacheMapLayer(tm Manager, ext *geo.MapExtent, image_opts *images.ImageOptions, maxTileLimit *int) *CacheMapLayer {
	if ext == nil {
		ext = geo.MapExtentFromGrid(tm.GetGrid())
	}
	ret := &CacheMapLayer{MapLayer: layer.MapLayer{SupportMetaTiles: true, Extent: ext, Options: image_opts}, tileManager: tm, grid: tm.GetGrid(), maxTileLimit: maxTileLimit}
	ret.ResRange = nil
	if tm.GetRescaleTiles() == -1 {
		ret.ResRange = layer.MergeLayerResRanges(tm.GetSources())
	}
	return ret
}

func (r *CacheMapLayer) checkTiled(query *layer.MapQuery) error {
	if string(query.Format) != r.tileManager.GetFormat() {
		return errors.New(fmt.Sprintf("invalid tile format, use %s", r.tileManager.GetFormat()))
	}
	if query.Size[0] != r.grid.TileSize[0] || query.Size[1] != r.grid.TileSize[1] {
		return errors.New(fmt.Sprintf("invalid tile size (use %dx%d)", r.grid.TileSize[0], r.grid.TileSize[1]))
	}
	return nil
}

func (r *CacheMapLayer) getSource(query *layer.MapQuery) (images.Source, error) {
	src_bbox, tile_grid, affected_tile_coords, err :=
		r.grid.GetAffectedTiles(query.BBox, query.Size, query.Srs)
	if err != nil {
		return nil, err
	}

	num_tiles := tile_grid[0] * tile_grid[1]

	if r.maxTileLimit != nil && num_tiles >= *r.maxTileLimit {
		return nil, errors.New(fmt.Sprintf("too many tiles, max_tile_limit: %d, num_tiles: %d", *r.maxTileLimit, num_tiles))
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

	_, tile_collection := r.tileManager.LoadTileCoords(coords, nil, query.TiledOnly)

	if tile_collection.Empty() {
		return &images.BlankImageSource{}, nil
	}

	if query.TiledOnly {
		t := tile_collection.GetItem(0)
		tile := t.Source
		tile.SetImageOptions(r.tileManager.GetImageOptions())
		tile.SetCacheable(t.Cacheable)
		return tile, nil
	}

	tile_sources := []images.Source{}
	for _, t := range tile_collection.tiles {
		tile_sources = append(tile_sources, t.Source)
	}
	tiled_image := images.NewTiledImage(tile_sources, tile_grid, [2]uint32{r.grid.TileSize[0], r.grid.TileSize[1]}, src_bbox, r.grid.Srs)
	return tiled_image.Transform(query.BBox, query.Srs, query.Size, r.tileManager.GetImageOptions()), nil
}

func (r *CacheMapLayer) GetMap(query *layer.MapQuery) images.Source {
	if err := r.CheckResRange(query); err != nil {
		return nil
	}

	if query.TiledOnly {
		r.checkTiled(query)
	}

	query_extent := &geo.MapExtent{BBox: query.BBox, Srs: query.Srs}
	var result images.Source
	if !query.TiledOnly && r.Extent != nil && !r.Extent.Contains(query_extent) {
		if !r.Extent.Intersects(query_extent) {
			return &images.BlankImageSource{}
		}
		size, offset, bbox := images.BBoxPositionInImage(query.BBox, query.Size, r.Extent.BBoxFor(query.Srs))
		if size[0] == 0 || size[1] == 0 {
			return &images.BlankImageSource{}
		}
		src_query := &layer.MapQuery{BBox: bbox, Size: size, Srs: query.Srs, Format: query.Format}
		resp, err := r.getSource(src_query)
		if err != nil {
			return &images.BlankImageSource{}
		}
		result = images.SubImageSource(resp.(*images.ImageSource), query.Size, offset[:], r.Options, resp.GetCacheable())
	} else {
		result, _ = r.getSource(query)
	}
	return result
}
