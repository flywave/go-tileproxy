package client

import (
	"math/rand"
	"net/url"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/resource"
)

type MapboxClient struct {
	BaseClient
	TilesURL        []string
	TilejsonURL     string
	TileStatsURL    string
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
	if c.AccessToken != "" {
		q.Set(c.AccessTokenName, c.AccessToken)
	}
	if c.Sku != "" {
		q.Set("sku", c.Sku)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type MapboxTileClient struct {
	MapboxClient
}

func NewMapboxTileClient(url, statsUrl, sku, token, tokenName string, ctx Context) *MapboxTileClient {
	return &MapboxTileClient{
		MapboxClient: MapboxClient{
			BaseClient:      BaseClient{ctx: ctx},
			TilejsonURL:     url,
			TileStatsURL:    statsUrl,
			AccessToken:     token,
			AccessTokenName: tokenName,
			Sku:             sku,
		},
	}
}

func (c *MapboxTileClient) GetTile(tile_coord [3]int) []byte {
	if len(c.TilesURL) == 0 {
		json := c.GetTileJSON()
		if json == nil {
			return nil
		}
	}

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
		ret := resource.CreateTileJSON(resp)
		if ret != nil {
			c.TilesURL = ret.Tiles[:]
			return ret
		}
	}
	return nil
}

func (c *MapboxTileClient) GetTileStats() *resource.TileStats {
	url, err := c.buildQuery(c.TileStatsURL)
	if err != nil {
		return nil
	}
	status, resp := c.httpClient().Open(url, nil, nil)
	if status == 200 {
		return resource.CreateTileStats(resp)
	}
	return nil
}

func (c *MapboxTileClient) buildTileQuery(tile_coord [3]int) string {
	url := c.TilesURL[rand.Intn(len(c.TilesURL)-1)]
	if strings.Contains(url, "{z}") && strings.Contains(url, "{x}") && strings.Contains(url, "{y}") {
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
