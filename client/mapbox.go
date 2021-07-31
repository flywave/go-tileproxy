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
}

type MapboxTileClient struct {
	MapboxClient
}

func (c *MapboxTileClient) GetVector(q *layer.TileQuery) []byte {
	url, err := q.BuildURL(c.BaseURL, c.AccessToken)
	if err != nil {
		return nil
	}
	resp := c.Get(url)
	if resp.StatusCode == 200 {
		return resp.Body
	}
	return nil
}

type MapboxSpriteClient struct {
	MapboxClient
}

func (c *MapboxSpriteClient) GetSprite(q *layer.SpriteQuery) *resource.Sprite {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	resp := c.Get(url)
	if resp.StatusCode == 200 {
		return resource.CreateSprite(resp.Body)
	}
	return nil
}

type MapboxStyleClient struct {
	MapboxClient
}

func (c *MapboxStyleClient) GetStyle(q *layer.StyleQuery) *resource.Style {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	resp := c.Get(url)
	if resp.StatusCode == 200 {
		return resource.CreateStyle(resp.Body)
	}
	return nil
}

type MapboxGlyphsClient struct {
	MapboxClient
}

func (c *MapboxGlyphsClient) GetGlyphs(q *layer.GlyphsQuery) *resource.Glyphs {
	url, err := q.BuildURL(c.BaseURL, c.UserName, c.AccessToken)
	if err != nil {
		return nil
	}
	resp := c.Get(url)
	if resp.StatusCode == 200 {
		return resource.CreateGlyphs(resp.Body)
	}
	return nil
}
