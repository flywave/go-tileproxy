package cache

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type CacheSource struct {
	CacheMapLayer
	TiledOnly bool
}

func NewCacheSource(tm Manager, ext *geo.MapExtent, image_opts *imagery.ImageOptions, maxTileLimit *int, tiled_only bool) *CacheSource {
	if ext == nil {
		ext = geo.MapExtentFromGrid(tm.GetGrid())
	}
	ret := &CacheSource{CacheMapLayer: CacheMapLayer{MapLayer: layer.MapLayer{SupportMetaTiles: !tiled_only, Extent: ext, Options: image_opts}, tileManager: tm, grid: tm.GetGrid(), maxTileLimit: maxTileLimit}, TiledOnly: tiled_only}
	ret.ResRange = nil
	if tm.GetRescaleTiles() == -1 {
		ret.ResRange = layer.MergeLayerResRanges(tm.GetSources())
	}
	return ret
}

func (r *CacheSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if r.TiledOnly {
		query.TiledOnly = true
	}
	return r.CacheMapLayer.GetMap(query)
}