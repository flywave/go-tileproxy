package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/resource"
)

type CesiumClient struct {
	BaseClient
	AuthURL     string
	BaseURL     string
	AssetId     int
	AuthToken   string
	AccessToken string
	AuthHeaders http.Header
	Version     string
	Extensions  []string
	TileURL     string
}

func (c *CesiumClient) buildAuthQuery() string {
	return fmt.Sprintf("%s/v1/assets/%d/endpoint?access_token=%s", c.AuthURL, c.AssetId, c.AccessToken)
}

type CesiumTileClient struct {
	CesiumClient
}

func NewCesiumTileClient(authUrl string, assetUrl string, assetId int, token string, ver string, tileUrl string, ctx Context) *CesiumTileClient {
	return &CesiumTileClient{
		CesiumClient: CesiumClient{
			BaseClient:  BaseClient{ctx: ctx},
			AuthURL:     authUrl,
			BaseURL:     assetUrl,
			TileURL:     tileUrl,
			AssetId:     assetId,
			AccessToken: token,
			Version:     ver,
			Extensions:  []string{"metadata"},
		},
	}
}

func (c *CesiumTileClient) IsAuth() bool {
	return c.AuthToken != ""
}

func (c *CesiumTileClient) Auth(tile_coord *[3]int) error {
	url := c.buildAuthQuery()
	c.AuthHeaders = make(http.Header)
	c.AuthHeaders["origin"] = []string{"http:127.0.0.1/test.html"}
	c.AuthHeaders["referer"] = []string{"http:127.0.0.1/test.html"}
	status, resp := c.httpClient().Open(url, nil, c.AuthHeaders)
	if status == 200 {
		type authResult struct {
			AccessToken string `json:"accessToken"`
		}
		result := &authResult{}

		err := json.Unmarshal(resp, result)
		if err != nil {
			return err
		}

		c.AuthToken = result.AccessToken
		c.AuthHeaders["authorization"] = []string{"Bearer " + c.AuthToken}
		return nil
	}
	return errors.New(string(resp))
}

func (c *CesiumTileClient) GetTile(tile_coord [3]int) []byte {
	if !c.IsAuth() {
		err := c.Auth(&tile_coord)
		if err != nil {
			return nil
		}
	}
	url := c.buildTileQuery(tile_coord)
	status, resp := c.httpClient().Open(url, nil, c.AuthHeaders)
	if status == 200 {
		return resp
	}
	return nil
}

func (c *CesiumTileClient) GetLayerJson() *resource.LayerJson {
	if !c.IsAuth() {
		err := c.Auth(nil)
		if err != nil {
			return nil
		}
	}
	url := c.buildLayerJson()
	status, resp := c.httpClient().Open(url, nil, c.AuthHeaders)
	if status == 200 {
		ret := resource.CreateLayerJson(resp)
		if ret != nil {
			c.Version = ret.Version
			c.Extensions = ret.Extensions
			c.TileURL = ret.Tiles[0]
		}
		return ret
	}
	return nil
}

func (c *CesiumTileClient) buildLayerJson() string {
	sub := fmt.Sprintf("%d/layer.json", c.AssetId)
	return fmt.Sprintf("%s/%s", c.BaseURL, sub)
}

func (c *CesiumTileClient) buildTileQuery(tile_coord [3]int) string {
	var url string
	if !strings.Contains(c.TileURL, "{z}") || !strings.Contains(c.TileURL, "{x}") || !strings.Contains(c.TileURL, "{y}") {
		c.TileURL = "{z}/{x}/{y}.terrain?{extensions}&{version}"
	}
	url = c.TileURL
	zstr := strconv.Itoa(tile_coord[2])
	xstr := strconv.Itoa(tile_coord[0])
	ystr := strconv.Itoa(tile_coord[1])

	url = strings.Replace(url, "{z}", zstr, 1)
	url = strings.Replace(url, "{x}", xstr, 1)
	url = strings.Replace(url, "{y}", ystr, 1)

	var extensions string
	if len(c.Extensions) == 1 {
		extensions = c.Extensions[0]
	} else {
		extensions = strings.Join(c.Extensions, "-")
	}
	url = strings.Replace(url, "{extensions}", "extensions="+extensions, 1)

	if strings.Contains(url, "{version}") {
		url = strings.Replace(url, "{version}", "v="+c.Version, 1)
	}
	url = fmt.Sprintf("%d/%s", c.AssetId, url)
	return fmt.Sprintf("%s/%s", c.BaseURL, url)
}
