package client

import "github.com/flywave/go-tileproxy/layer"

type LuoKuangTileClient struct {
	MapboxClient
}

func NewLuoKuangTileClient(url string, token string, mapid string, ctx Context) *LuoKuangTileClient {
	return &LuoKuangTileClient{MapboxClient: MapboxClient{BaseClient: BaseClient{ctx: ctx}, BaseURL: url, AccessToken: token, MapId: mapid}}
}

func (c *LuoKuangTileClient) GetTile(q *layer.LuoKuangTileQuery) []byte {
	url, err := q.BuildURL(c.BaseURL, c.AccessToken, c.MapId)
	if err != nil {
		return nil
	}
	status, resp := c.GetHttpClient().Open(url, nil)
	if status == 200 {
		return resp
	}
	return nil
}
