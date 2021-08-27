package client

import "github.com/flywave/go-tileproxy/layer"

type LuoKuangTileClient struct {
	MapboxClient
}

func NewLuoKuangTileClient(url string, version string, token string, tilesetID string, ctx Context) *LuoKuangTileClient {
	return &LuoKuangTileClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, Version: version, AccessToken: token, TilesetID: tilesetID}}
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
