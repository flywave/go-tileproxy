package sources

import (
	"errors"
	"fmt"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxTileSource struct {
	layer.MapLayer
	Grid           *geo.TileGrid
	Client         *client.MapboxTileClient
	SourceCreater  tile.SourceCreater
	TileJSONCache  *resource.TileJSONCache
	TileStatsCache *resource.TileStatsCache
}

func NewMapboxTileSource(grid *geo.TileGrid, coverage geo.Coverage, c *client.MapboxTileClient, opts tile.TileOptions, creater tile.SourceCreater,
	tileJSONCache *resource.TileJSONCache, tileStatsCache *resource.TileStatsCache) *MapboxTileSource {
	return &MapboxTileSource{
		Grid:   grid,
		Client: c,
		MapLayer: layer.MapLayer{
			Options:  opts,
			Coverage: coverage,
		},
		SourceCreater:  creater,
		TileJSONCache:  tileJSONCache,
		TileStatsCache: tileStatsCache,
	}
}

func (s *MapboxTileSource) GetTileJSON(id string) *resource.TileJSON {
	ret := &resource.TileJSON{StoreID: id}

	if s.TileJSONCache != nil && s.TileJSONCache.Load(ret) != nil {
		ret = s.Client.GetTileJSON()
		if ret != nil {
			ret.StoreID = id
			s.TileJSONCache.Save(ret)
		}
	}
	return ret
}

func (s *MapboxTileSource) GetTileStats(id string) *resource.TileStats {
	ret := &resource.TileStats{StoreID: id}

	if s.TileStatsCache != nil && s.TileStatsCache.Load(ret) != nil {
		ret = s.Client.GetTileStats()
		if ret != nil {
			ret.StoreID = id
			s.TileStatsCache.Save(ret)
		}
	}
	return ret
}

func (s *MapboxTileSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if s.Grid.TileSize[0] != (query.Size[0]/query.MetaSize[0]) || s.Grid.TileSize[1] != (query.Size[1]/query.MetaSize[0]) {
		return nil, errors.New("tile size of cache and tile source do not match")
	}

	// if !s.Grid.Srs.Eq(query.Srs) {
	// 	return nil, errors.New("srs of cache and tile source do not match")
	// }

	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return s.SourceCreater.CreateEmpty(query.Size, s.Options), nil
	}

	if s.Coverage != nil && !s.Coverage.Intersects(query.BBox, query.Srs) {
		return s.SourceCreater.CreateEmpty(query.Size, s.Options), nil
	}

	_, _, tiles, err := s.Grid.GetAffectedTiles(query.BBox, query.Size, query.Srs)

	if err != nil {
		return nil, err
	}

	x, y, z, _ := tiles.Next()

	resp := s.Client.GetTile([3]int{x, y, z})
	if len(resp) == 0 {
		return nil, fmt.Errorf("tile %d %d %d %s", x, y, z, "have no data")
	}
	return s.SourceCreater.Create(resp, [3]int{x, y, z}), nil
}
