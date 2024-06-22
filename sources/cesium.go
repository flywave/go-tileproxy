package sources

import (
	"errors"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type CesiumTileSource struct {
	layer.MapLayer
	Grid          *geo.TileGrid
	Client        *client.CesiumTileClient
	SourceCreater tile.SourceCreater
	Cache         *resource.LayerJSONCache
}

func NewCesiumTileSource(
	grid *geo.TileGrid,
	c *client.CesiumTileClient,
	opts tile.TileOptions,
	creater tile.SourceCreater,
	cache *resource.LayerJSONCache,
) *CesiumTileSource {
	return &CesiumTileSource{
		Grid:   grid,
		Client: c,
		MapLayer: layer.MapLayer{
			Options: opts,
		},
		SourceCreater: creater,
		Cache:         cache,
	}
}

func (s *CesiumTileSource) GetLayerJSON(id string) *resource.LayerJson {
	ret := &resource.LayerJson{StoreID: id}
	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret, _ = s.Client.GetLayerJson()
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

func (s *CesiumTileSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if s.Grid.TileSize[0] != query.Size[0] || s.Grid.TileSize[1] != query.Size[1] {
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

	_, grid, tiles, err := s.Grid.GetAffectedTiles(query.BBox, query.Size, nil)

	if err != nil {
		return nil, err
	}

	if grid != [2]int{1, 1} {
		return nil, errors.New("bbox does not align to tile")
	}

	x, y, z, _ := tiles.Next()

	resp := s.Client.GetTile([3]int{x, y, z})
	if len(resp) == 0 {
		return nil, errors.New("data is nil")
	}
	return s.SourceCreater.Create(resp, [3]int{x, y, z}), nil
}
