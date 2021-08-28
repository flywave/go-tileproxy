package client

import (
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type LuoKuangTileClient struct {
	MapboxClient
	TilesetID string
}

func NewLuoKuangTileClient(url string, version string, token string, tilesetID string, ctx Context) *LuoKuangTileClient {
	return &LuoKuangTileClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, Version: version, AccessToken: token}, TilesetID: tilesetID}
}

func (c *LuoKuangTileClient) GetTile(q *layer.LuoKuangTileQuery) []byte {
	url, err := q.BuildURL(c.BaseURL, c.Version, c.AccessToken, c.TilesetID)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resp
	}
	return nil
}

type LuoKuangStyleClient struct {
	MapboxClient
	StyleID string
}

func NewLuoKuangStyleClient(url string, version string, userName string, token string, ctx Context) *LuoKuangStyleClient {
	return &LuoKuangStyleClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, Version: version, UserName: userName, AccessToken: token}}
}

func (c *LuoKuangStyleClient) GetSpriteJSON(q *layer.SpriteQuery) *resource.SpriteJSON {
	styleid := q.StyleID
	if c.StyleID != "" {
		styleid = c.StyleID
	}
	url, err := q.BuildURL(c.BaseURL, c.Version, c.UserName, styleid, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateSpriteJSON(resp)
	}
	return nil
}

func (c *LuoKuangStyleClient) GetSprite(q *layer.SpriteQuery) *resource.Sprite {
	styleid := q.StyleID
	if c.StyleID != "" {
		styleid = c.StyleID
	}
	url, err := q.BuildURL(c.BaseURL, c.Version, c.UserName, styleid, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateSprite(resp)
	}
	return nil
}

func (c *LuoKuangStyleClient) GetStyle(q *layer.StyleQuery) *resource.Style {
	styleid := q.StyleID
	if c.StyleID != "" {
		styleid = c.StyleID
	}
	url, err := q.BuildURL(c.BaseURL, c.Version, c.UserName, styleid, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateStyle(resp)
	}
	return nil
}

type LuoKuangGlyphsClient struct {
	MapboxClient
	Font string
}

func NewLuoKuangGlyphsClient(url string, version string, userName string, token string, ctx Context) *LuoKuangGlyphsClient {
	return &LuoKuangGlyphsClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, Version: version, UserName: userName, AccessToken: token}}
}

func (c *LuoKuangGlyphsClient) GetGlyphs(q *layer.GlyphsQuery) *resource.Glyphs {
	font := q.Font
	if c.Font != "" {
		font = c.Font
	}
	url, err := q.BuildURL(c.BaseURL, c.Version, c.UserName, font, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateGlyphs(resp)
	}
	return nil
}

type LuoKuangTileJSONClient struct {
	MapboxClient
	TilesetID string
}

func NewLuoKuangTileJSONClient(url string, version string, userName string, token string, ctx Context) *LuoKuangTileJSONClient {
	return &LuoKuangTileJSONClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, Version: version, UserName: userName, AccessToken: token}}
}

func (c *LuoKuangTileJSONClient) GetTileJSON(q *layer.TileJSONQuery) *resource.TileJSON {
	tilesetid := q.TilesetID
	if c.TilesetID != "" {
		tilesetid = c.TilesetID
	}
	url, err := q.BuildURL(c.BaseURL, c.Version, c.UserName, tilesetid, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateTileJSON(resp)
	}
	return nil
}
