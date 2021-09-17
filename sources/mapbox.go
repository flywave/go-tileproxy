package sources

import (
	"errors"

	"github.com/flywave/go-geo"
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
	return &MapboxTileSource{Grid: grid, Client: c, MapLayer: layer.MapLayer{Options: opts}, SourceCreater: creater, Cache: cache}
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
	return s.SourceCreater.Create(resp, [3]int{x, y, z}), nil
}

type MapboxStyleSource struct {
	Client *client.MapboxStyleClient
	Sprite *client.MapboxStyleClient
	Cache  *resource.StyleCache
}

func NewMapboxStyleSource(c *client.MapboxStyleClient, sprite *client.MapboxStyleClient, cache *resource.StyleCache) *MapboxStyleSource {
	return &MapboxStyleSource{Client: c, Sprite: sprite, Cache: cache}
}

func (s *MapboxStyleSource) GetSpriteJSON(id string) *resource.SpriteJSON {
	ret := &resource.SpriteJSON{BaseResource: resource.BaseResource{StoreID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Sprite.GetSpriteJSON()
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

func (s *MapboxStyleSource) GetSprite(id string) *resource.Sprite {
	ret := &resource.Sprite{BaseResource: resource.BaseResource{StoreID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Sprite.GetSprite()
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

func (s *MapboxStyleSource) GetStyle(id string) *resource.Style {
	ret := &resource.Style{BaseResource: resource.BaseResource{StoreID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetStyle()
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

type MapboxGlyphsSource struct {
	Client *client.MapboxStyleClient
	Cache  *resource.GlyphsCache
	Fonts  []string
}

func NewMapboxGlyphsSource(c *client.MapboxStyleClient, fonts []string, cache *resource.GlyphsCache) *MapboxGlyphsSource {
	return &MapboxGlyphsSource{Client: c, Cache: cache, Fonts: fonts}
}

func (s *MapboxGlyphsSource) GetGlyphs(query *layer.GlyphsQuery) *resource.Glyphs {
	id := query.GetID()

	ret := &resource.Glyphs{BaseResource: resource.BaseResource{StoreID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetGlyphs(query)
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}
