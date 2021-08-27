package sources

import (
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
	Client        *client.MapboxTileClient
	SourceCreater tile.SourceCreater
}

func NewMapboxTileSource(grid *geo.TileGrid, c *client.MapboxTileClient, opts tile.TileOptions, creater tile.SourceCreater) *MapboxTileSource {
	return &MapboxTileSource{Grid: grid, Client: c, MapLayer: layer.MapLayer{Options: opts}, SourceCreater: creater}
}

func (s *MapboxTileSource) GetMap(query *layer.MapQuery) (tile.Source, error) {
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

func (s *MapboxTileSource) buildTileQuery(x, y, z int, query *layer.MapQuery) *layer.TileQuery {
	var retina *int
	if query.Dimensions != nil {
		if v, ok := query.Dimensions["retina"]; ok {
			r, _ := v.GetFirstValue().(int)
			retina = geo.NewInt(r)
		}
	}
	tile := &layer.TileQuery{X: x, Y: y, Zoom: z, Width: int(query.Size[0]), Height: int(query.Size[1]), Format: query.Format.Extension(), Retina: retina}
	return tile
}

type MapboxStyleSource struct {
	Client *client.MapboxStyleClient
	Cache  *resource.StyleCache
}

func NewMapboxStyleSource(c *client.MapboxStyleClient, cache *resource.StyleCache) *MapboxStyleSource {
	return &MapboxStyleSource{Client: c, Cache: cache}
}

func (s *MapboxStyleSource) GetSpriteJSON(query *layer.SpriteQuery) *resource.SpriteJSON {
	id := query.GetID()

	ret := &resource.SpriteJSON{BaseResource: resource.BaseResource{ID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetSpriteJSON(query)
	}

	return ret
}

func (s *MapboxStyleSource) GetSprite(query *layer.SpriteQuery) *resource.Sprite {
	id := query.GetID()

	ret := &resource.Sprite{BaseResource: resource.BaseResource{ID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetSprite(query)
	}

	return ret
}

func (s *MapboxStyleSource) GetStyle(query *layer.StyleQuery) *resource.Style {
	id := query.GetID()

	ret := &resource.Style{BaseResource: resource.BaseResource{ID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetStyle(query)
	}

	return ret
}

type MapboxGlyphsSource struct {
	Client *client.MapboxGlyphsClient
	Cache  *resource.GlyphsCache
}

func NewMapboxGlyphsSource(c *client.MapboxGlyphsClient, cache *resource.GlyphsCache) *MapboxGlyphsSource {
	return &MapboxGlyphsSource{Client: c, Cache: cache}
}

func (s *MapboxGlyphsSource) GetGlyphs(query *layer.GlyphsQuery) *resource.Glyphs {
	id := query.GetID()

	ret := &resource.Glyphs{BaseResource: resource.BaseResource{ID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetGlyphs(query)
	}

	return ret
}

type MapboxTileJSONSource struct {
	Client *client.MapboxTileJSONClient
	Cache  *resource.TileJSONCache
}

func NewMapboxTileJSONSource(c *client.MapboxTileJSONClient, cache *resource.TileJSONCache) *MapboxTileJSONSource {
	return &MapboxTileJSONSource{Client: c, Cache: cache}
}

func (s *MapboxTileJSONSource) GetTileJSON(query *layer.TileJSONQuery) *resource.TileJSON {
	ret := &resource.TileJSON{Id: query.TilesetID}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetTileJSON(query)
	}

	return ret
}
