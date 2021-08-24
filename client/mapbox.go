package client

import (
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type MapboxClient struct {
	BaseClient
	BaseURL     string
	UserName    string
	AccessToken string
	MapId       string
}

type MapboxTileClient struct {
	MapboxClient
}

func NewMapboxTileClient(url string, userName string, token string, mapid string, ctx Context) *MapboxTileClient {
	return &MapboxTileClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, UserName: userName, AccessToken: token, MapId: mapid}}
}

func (c *MapboxTileClient) GetTile(q *layer.TileQuery) []byte {
	url, err := q.BuildURL(c.BaseURL, c.AccessToken, c.MapId)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resp
	}
	return nil
}

type MapboxStyleClient struct {
	MapboxClient
}

func NewMapboxStyleClient(url string, userName string, token string, ctx Context) *MapboxStyleClient {
	return &MapboxStyleClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, UserName: userName, AccessToken: token}}
}

func (c *MapboxStyleClient) GetSpriteJSON(q *layer.SpriteQuery) *resource.SpriteJSON {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateSpriteJSON(resp)
	}
	return nil
}

func (c *MapboxStyleClient) GetSprite(q *layer.SpriteQuery) *resource.Sprite {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateSprite(resp)
	}
	return nil
}

func (c *MapboxStyleClient) GetStyle(q *layer.StyleQuery) *resource.Style {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateStyle(resp)
	}
	return nil
}

type MapboxGlyphsClient struct {
	MapboxClient
}

func NewMapboxGlyphsClient(url string, userName string, token string, ctx Context) *MapboxGlyphsClient {
	return &MapboxGlyphsClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, UserName: userName, AccessToken: token}}
}

func (c *MapboxGlyphsClient) GetGlyphs(q *layer.GlyphsQuery) *resource.Glyphs {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateGlyphs(resp)
	}
	return nil
}

type MapboxTileJSONClient struct {
	MapboxClient
}

func NewMapboxTileJSONClient(url string, userName string, token string, ctx Context) *MapboxTileJSONClient {
	return &MapboxTileJSONClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, UserName: userName, AccessToken: token}}
}

func (c *MapboxTileJSONClient) GetTileJSON(q *layer.TileJSONQuery) *resource.TileJSON {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateTileJSON(resp)
	}
	return nil
}
