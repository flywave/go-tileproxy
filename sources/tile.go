package sources

import (
	"bytes"
	"errors"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type SourceCreater func(size [2]uint32, opts tile.TileOptions, data interface{}) tile.Source

type TileSource struct {
	layer.MapLayer
	Grid          *geo.TileGrid
	Client        *client.TileClient
	Options       tile.TileOptions
	Coverage      geo.Coverage
	ResRange      *geo.ResolutionRange
	SourceCreater SourceCreater
}

func NewTileSource(grid *geo.TileGrid, c *client.TileClient, opts tile.TileOptions, creater SourceCreater) *TileSource {
	return &TileSource{Grid: grid, Client: c, Options: opts, SourceCreater: creater}
}

func (s *TileSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if s.Grid.TileSize[0] != query.Size[0] || s.Grid.TileSize[1] != query.Size[1] {
		return nil, errors.New("tile size of cache and tile source do not match")
	}

	if !s.Grid.Srs.Eq(query.Srs) {
		return nil, errors.New("SRS of cache and tile source do not match")
	}

	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return s.SourceCreater(query.Size, s.Options, nil), nil
	}

	if s.Coverage != nil && !s.Coverage.Intersects(query.BBox, query.Srs) {
		return s.SourceCreater(query.Size, s.Options, nil), nil
	}

	_, grid, tiles, err := s.Grid.GetAffectedTiles(query.BBox, query.Size, nil)

	if err != nil {
		return nil, err
	}

	if grid != [2]int{1, 1} {
		return nil, errors.New("BBOX does not align to tile")
	}

	x, y, z, _ := tiles.Next()

	resp := s.Client.GetTile([3]int{x, y, z}, &query.Format)
	src := s.SourceCreater(query.Size, s.Options, bytes.NewBuffer(resp))
	return src, nil
}
