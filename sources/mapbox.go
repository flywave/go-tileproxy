package sources

import (
	"errors"
	"fmt"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxTileSource struct {
	layer.MapLayer
	Grid          *geo.TileGrid
	Client        *client.MapboxTileClient
	SourceCreater tile.SourceCreater
	Cache         *resource.TileJSONCache
}

func NewMapboxTileSource(grid *geo.TileGrid, c *client.MapboxTileClient, opts tile.TileOptions, creater tile.SourceCreater, cache *resource.TileJSONCache) *MapboxTileSource {
	return &MapboxTileSource{
		Grid:   grid,
		Client: c,
		MapLayer: layer.MapLayer{
			Options: opts,
		},
		SourceCreater: creater,
		Cache:         cache,
	}
}

func (s *MapboxTileSource) GetTileJSON(id string) *resource.TileJSON {
	ret := &resource.TileJSON{StoreID: id}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetTileJSON()
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

func (s *MapboxTileSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if s.Grid.TileSize[0] != (query.Size[0]/query.MetaSize[0]) || s.Grid.TileSize[1] != (query.Size[1]/query.MetaSize[0]) {
		return nil, errors.New("tile size of cache and tile source do not match")
	}

	if !s.Grid.Srs.Eq(query.Srs) {
		return nil, errors.New("srs of cache and tile source do not match")
	}

	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return s.SourceCreater.CreateEmpty(query.Size, s.Options), nil
	}

	if s.Coverage != nil && !s.Coverage.Intersects(query.BBox, query.Srs) {
		return s.SourceCreater.CreateEmpty(query.Size, s.Options), nil
	}

	_, _, tiles, err := s.Grid.GetAffectedTiles(query.BBox, query.Size, nil)

	if err != nil {
		return nil, err
	}

	sources := []tile.Source{}
	for {
		x, y, z, done := tiles.Next()

		resp := s.Client.GetTile([3]int{x, y, z})
		if len(resp) == 0 {
			return nil, fmt.Errorf("tile %d %d %d %s", x, y, z, "have no data")
		}
		sources = append(sources, s.SourceCreater.Create(resp, [3]int{x, y, z}))
		if done {
			break
		}
	}
	if len(sources) == 1 {
		return sources[0], nil
	}
	return cache.MergeTiles(sources, s.Options, query, nil)
}
