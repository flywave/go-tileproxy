package sources

import (
	"bytes"
	"errors"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxTileSource struct {
	layer.MapLayer
	Grid          *geo.TileGrid
	Coverage      geo.Coverage
	Extent        *geo.MapExtent
	ResRange      *geo.ResolutionRange
	Client        *client.MapboxTileClient
	ImageOpts     tile.TileOptions
	SourceCreater SourceCreater
}

func (s *MapboxTileSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
	if s.Grid.TileSize[0] != query.Size[0] || s.Grid.TileSize[1] != query.Size[1] {
		return nil, errors.New("tile size of cache and tile source do not match")
	}

	if !s.Grid.Srs.Eq(query.Srs) {
		return nil, errors.New("SRS of cache and tile source do not match")
	}

	if s.ResRange != nil && !s.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return s.SourceCreater(query.Size, s.ImageOpts, nil), nil
	}

	if s.Coverage != nil && !s.Coverage.Intersects(query.BBox, query.Srs) {
		return s.SourceCreater(query.Size, s.ImageOpts, nil), nil
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
	src := s.SourceCreater(query.Size, s.ImageOpts, bytes.NewBuffer(resp))
	return src, nil
}

func (s *MapboxTileSource) buildTileQuery(x, y, z int, query *layer.MapQuery) *layer.TileQuery {
	retina := false
	if query.Dimensions != nil {
		if _, ok := query.Dimensions["retina"]; ok {
			retina = true
		}
	}
	tile := &layer.TileQuery{X: x, Y: y, Zoom: z, Width: int(query.Size[0]), Height: int(query.Size[1]), Format: query.Format.MimeType(), Retina: retina}
	return tile
}

type MapboxSpriteSource struct {
	Client *client.MapboxSpriteClient
}

func (s *MapboxSpriteSource) GetSprite(query *layer.SpriteQuery) *resource.Sprite {
	return nil
}

type MapboxStyleSource struct {
	Client *client.MapboxStyleClient
}

func (s *MapboxStyleSource) GetStyle(query *layer.StyleQuery) *resource.Style {
	return nil
}

type MapboxGlyphsSource struct {
	Client *client.MapboxGlyphsClient
}

func (s *MapboxGlyphsSource) GetGlyphs(query *layer.GlyphsQuery) *resource.Glyphs {
	return nil
}
