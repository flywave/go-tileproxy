package sources

import (
	"errors"
	"fmt"

	"github.com/flywave/go-tileproxy/client"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
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

type LuoKuangStyleSource struct {
	Client *client.LuoKuangStyleClient
	Cache  *resource.StyleCache
}

func NewLuoKuangStyleSource(c *client.LuoKuangStyleClient, cache *resource.StyleCache) *LuoKuangStyleSource {
	return &LuoKuangStyleSource{Client: c, Cache: cache}
}

func (s *LuoKuangStyleSource) GetSpriteJSON(query *layer.SpriteQuery) *resource.SpriteJSON {
	id := query.GetID()

	ret := &resource.SpriteJSON{BaseResource: resource.BaseResource{StoreID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetSpriteJSON(query)
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

func (s *LuoKuangStyleSource) GetSprite(query *layer.SpriteQuery) *resource.Sprite {
	id := query.GetID()

	ret := &resource.Sprite{BaseResource: resource.BaseResource{StoreID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetSprite(query)
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

func (s *LuoKuangStyleSource) GetStyle(query *layer.StyleQuery) *resource.Style {
	id := query.GetID()

	ret := &resource.Style{BaseResource: resource.BaseResource{StoreID: id}}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetStyle(query)
		if ret != nil {
			ret.StoreID = id
			s.Cache.Save(ret)
		}
	}

	return ret
}

type LuoKuangGlyphsSource struct {
	Client *client.LuoKuangGlyphsClient
	Cache  *resource.GlyphsCache
}

func NewLuoKuangGlyphsSource(c *client.LuoKuangGlyphsClient, cache *resource.GlyphsCache) *LuoKuangGlyphsSource {
	return &LuoKuangGlyphsSource{Client: c, Cache: cache}
}

func (s *LuoKuangGlyphsSource) GetGlyphs(query *layer.GlyphsQuery) *resource.Glyphs {
	id := fmt.Sprintf("%s-%d-%d", query.GetID(), query.Start, query.End)

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

type LuoKuangTileJSONSource struct {
	Client *client.LuoKuangTileJSONClient
	Cache  *resource.TileJSONCache
}

func NewLuoKuangTileJSONSource(c *client.LuoKuangTileJSONClient, cache *resource.TileJSONCache) *LuoKuangTileJSONSource {
	return &LuoKuangTileJSONSource{Client: c, Cache: cache}
}

func (s *LuoKuangTileJSONSource) GetTileJSON(query *layer.TileJSONQuery) *resource.TileJSON {
	ret := &resource.TileJSON{Id: query.TilesetID}

	if s.Cache != nil && s.Cache.Load(ret) != nil {
		ret = s.Client.GetTileJSON(query)
		if ret != nil {
			ret.StoreID = query.TilesetID
			s.Cache.Save(ret)
		}
	}

	return ret
}
