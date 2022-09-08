package client

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/resource"
)

type MapboxClient struct {
	BaseClient
	BaseURL         string
	TilejsonURL     string
	Sku             string
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
	q.Set("sku", c.Sku)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type MapboxTileClient struct {
	MapboxClient
}

func NewMapboxTileClient(urlTemplate, tilejsonUrl, sku, token, tokenName string, ctx Context) *MapboxTileClient {
	return &MapboxTileClient{
		MapboxClient: MapboxClient{
			BaseClient:      BaseClient{ctx: ctx},
			BaseURL:         urlTemplate,
			TilejsonURL:     tilejsonUrl,
			AccessToken:     token,
			AccessTokenName: tokenName,
			Sku:             sku,
		},
	}
}

func (c *MapboxTileClient) GetTile(tile_coord [3]int) []byte {
	url := c.buildTileQuery(tile_coord)
	status, resp := c.httpClient().Open(url, nil, nil)
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
	status, resp := c.httpClient().Open(url, nil, nil)
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

		url, _ = c.MapboxClient.buildQuery(url)

		return url
	}
	return ""
}
