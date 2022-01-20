package client

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/resource"
)

type MapboxClient struct {
	BaseClient
	BaseURL         string
	TilejsonURL     string
	AccessToken     string
	AccessTokenName string
}

func (c *MapboxClient) buildQuery(url_ string) (string, error) {
	u, err := url.Parse(url_)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set(c.AccessTokenName, c.AccessToken)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type MapboxTileClient struct {
	MapboxClient
}

func NewMapboxTileClient(urlTemplate string, tilejsonUrl string, token string, tokenName string, ctx Context) *MapboxTileClient {
	return &MapboxTileClient{
		MapboxClient: MapboxClient{
			BaseClient:      BaseClient{ctx: ctx},
			BaseURL:         urlTemplate,
			TilejsonURL:     tilejsonUrl,
			AccessToken:     token,
			AccessTokenName: tokenName,
		},
	}
}

func (c *MapboxTileClient) GetTile(tile_coord [3]int) []byte {
	url := c.buildTileQuery(tile_coord)
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resp
	}
	return nil
}

func (c *MapboxTileClient) GetTileJSON() *resource.TileJSON {
	url, err := c.buildQuery(c.TilejsonURL)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateTileJSON(resp)
	}
	return nil
}

func (c *MapboxTileClient) buildTileQuery(tile_coord [3]int) string {
	if strings.Contains(c.BaseURL, "{z}") && strings.Contains(c.BaseURL, "{x}") && strings.Contains(c.BaseURL, "{y}") {
		url := c.BaseURL

		zstr := strconv.Itoa(tile_coord[2])
		xstr := strconv.Itoa(tile_coord[0])
		ystr := strconv.Itoa(tile_coord[1])

		url = strings.Replace(url, "{z}", zstr, 1)
		url = strings.Replace(url, "{x}", xstr, 1)
		url = strings.Replace(url, "{y}", ystr, 1)

		return url
	}
	return ""
}

type MapboxStyleClient struct {
	MapboxClient
	StyleContentAttr *string
}

func NewMapboxStyleClient(url string, token string, tokenName string, ctx Context) *MapboxStyleClient {
	return &MapboxStyleClient{
		MapboxClient: MapboxClient{
			BaseClient:      BaseClient{ctx: ctx},
			BaseURL:         url,
			AccessToken:     token,
			AccessTokenName: tokenName,
		},
	}
}

func (c *MapboxStyleClient) GetSpriteJSON() *resource.SpriteJSON {
	url, err := c.buildQuery(c.BaseURL)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateSpriteJSON(resp)
	}
	return nil
}

func (c *MapboxStyleClient) GetSprite() *resource.Sprite {
	url, err := c.buildQuery(c.BaseURL)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateSprite(resp)
	}
	return nil
}

func (c *MapboxStyleClient) GetStyle() *resource.Style {
	url, err := c.buildQuery(c.BaseURL)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		if c.StyleContentAttr != nil {
			return resource.ExtractStyle(resp, *c.StyleContentAttr)
		} else {
			return resource.CreateStyle(resp)
		}
	}
	return nil
}

func (c *MapboxStyleClient) GetGlyphs(q *layer.GlyphsQuery) *resource.Glyphs {
	url, err := c.buildGlyphsURL(q.Font, q.Start, q.End)

	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil)
	if status == 200 {
		return resource.CreateGlyphs(resp)
	}
	return nil
}

func (c *MapboxStyleClient) buildGlyphsURL(font string, start, end int) (string, error) {
	furl := c.BaseURL
	if strings.Contains(furl, "{fontstack}") && strings.Contains(furl, "{range}") {
		rangestr := strconv.Itoa(start) + "-" + strconv.Itoa(end)

		furl = strings.Replace(furl, "{fontstack}", font, 1)
		furl = strings.Replace(furl, "{range}", rangestr, 1)

		u, err := url.Parse(furl)
		if err != nil {
			return "", err
		}

		q := u.Query()
		q.Set(c.AccessTokenName, c.AccessToken)
		u.RawQuery = q.Encode()
		return u.String(), nil
	}
	return "", errors.New("build glyphs url error")
}
