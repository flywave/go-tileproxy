package sources

import (
	"errors"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
)

type LuoKuangTileSource struct {
	layer.MapLayer
	Grid          *geo.TileGrid
	Coverage      geo.Coverage
	ResRange      *geo.ResolutionRange
	Client        *client.LuoKuangTileClient
	Options       tile.TileOptions
	SourceCreater tile.SourceCreater
}

func NewLuoKuangTileSource(grid *geo.TileGrid, c *client.LuoKuangTileClient, opts tile.TileOptions, creater tile.SourceCreater) *LuoKuangTileSource {
	return &LuoKuangTileSource{Grid: grid, Client: c, Options: opts, SourceCreater: creater}
}

func (s *LuoKuangTileSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if s.Grid.TileSize[0] != query.Size[0] || s.Grid.TileSize[1] != query.Size[1] {
		return nil, errors.New("tile size of cache and tile source do not match")
	}

	if !s.Grid.Srs.Eq(query.Srs) {
		return nil, errors.New("SRS of cache and tile source do not match")
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
		return nil, errors.New("BBOX does not align to tile")
	}

	x, y, z, _ := tiles.Next()

	tilequery := s.buildTileQuery(x, y, z, query)
	resp := s.Client.GetTile(tilequery)
	src := s.SourceCreater.Create(resp, [3]int{x, y, z})
	return src, nil
}

func (s *LuoKuangTileSource) buildTileQuery(x, y, z int, query *layer.MapQuery) *layer.LuoKuangTileQuery {
	tile := &layer.LuoKuangTileQuery{X: x, Y: y, Zoom: z, Width: int(query.Size[0]), Height: int(query.Size[1]), Format: query.Format.Extension(), Style: "main"}
	return tile
}
