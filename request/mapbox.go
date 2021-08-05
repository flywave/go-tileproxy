package request

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type MapboxRequest struct {
	BaseRequest
	RequestHandlerName string
	AccessToken        string
	ReqRegex           *regexp.Regexp
	Version            string
}

type MapboxTileRequest struct {
	MapboxRequest
	TilesetID string
	Tile      []int
	Format    *tile.TileFormat
	Retina    *int
	Origin    string
}

func (r *MapboxTileRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "tile"
	r.Version = "v4"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.Retina = nil
	r.ReqRegex = regexp.MustCompile(`^/(?P<version>[^/]+)/(?P<tileset_id>[^/]+)/(?P<zoom>-?\d+)/(?P<x>-?\d+)/(?P<y>-?\d+)(@(?P<retina>[^/]+)x)?\.(?P<format>\w+)`)
	r.initRequest()
}

func (r *MapboxTileRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if match == nil || len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["tileset_id"]; ok {
		r.TilesetID = v
	}
	if v, ok := result["retina"]; ok {
		rt, _ := strconv.ParseInt(v, 10, 64)
		r.Retina = geo.NewInt(int(rt))
	}
	if r.Tile != nil {
		x, _ := strconv.ParseInt(result["x"], 10, 64)
		y, _ := strconv.ParseInt(result["y"], 10, 64)
		z, _ := strconv.ParseInt(result["zoom"], 10, 64)
		r.Tile = []int{int(x), int(y), int(z)}
	}
	if r.Format != nil {
		*r.Format = tile.TileFormat(result["format"])
	}
	return nil
}

type MapboxStyleRequest struct {
	MapboxRequest
	Username string
	StyleID  string
}

func (r *MapboxStyleRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "style"
	r.Version = "v1"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.ReqRegex = regexp.MustCompile(`^/styles/(?P<version>[^/]+)/(?P<username>[^/]+)/(?P<style_id>[^/]+)`)
	r.initRequest()
}

func (r *MapboxStyleRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if match == nil || len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["username"]; ok {
		r.Username = v
	}
	if v, ok := result["style_id"]; ok {
		r.StyleID = v
	}
	return nil
}

type MapboxSpriteRequest struct {
	MapboxRequest
	Username string
	StyleID  string
	Format   *tile.TileFormat
	Retina   *int
}

func (r *MapboxSpriteRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "sprite"
	r.Version = "v1"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.ReqRegex = regexp.MustCompile(`^/styles/(?P<version>[^/]+)/(?P<username>[^/]+)/(?P<style_id>[^/]+)/sprite(@(?P<retina>[^/]+)x)?\.?(?P<format>\w+)?`)
	r.initRequest()
}

func (r *MapboxSpriteRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		result[name] = match[i]
	}

	if match == nil || len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["username"]; ok {
		r.Username = v
	}
	if v, ok := result["style_id"]; ok {
		r.StyleID = v
	}
	if v, ok := result["retina"]; ok {
		rt, _ := strconv.ParseInt(v, 10, 64)
		r.Retina = geo.NewInt(int(rt))
	}
	if r.Format != nil {
		*r.Format = tile.TileFormat(result["format"])
	}
	return nil
}

type MapboxGlyphsRequest struct {
	MapboxRequest
	Username string
	Font     string
}

func (r *MapboxGlyphsRequest) init(param interface{}, url string, validate bool, http *http.Request) {
	r.BaseRequest.init(param, url, validate, http)
	r.RequestHandlerName = "glyphs"
	r.Version = "v1"
	r.AccessToken = r.Params.GetOne("access_token", "")
	r.ReqRegex = regexp.MustCompile(`^/fonts/(?P<version>[^/]+)/(?P<username>[^/]+)/(?P<font>[^/]+)/(?P<start>-?\d+)-(?P<end>-?\d+)\.(?P<format>\w+)`)
	r.initRequest()
}

func (r *MapboxGlyphsRequest) initRequest() error {
	match := r.ReqRegex.FindStringSubmatch(r.Http.URL.Path)
	groupNames := r.ReqRegex.SubexpNames()
	result := make(map[string]string)
	for i, name := range groupNames {
		if name != "" && match[i] != "" {
			result[name] = match[i]
		}
	}

	if match == nil || len(match) == 0 || result["version"] != r.Version {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["format"]; !ok || v != "pbf" {
		return errors.New(fmt.Sprintf("invalid request (%s)", r.Http.URL.Path))
	}

	if v, ok := result["username"]; ok {
		str, _ := url.PathUnescape(v)
		r.Username = str
	}
	if v, ok := result["font"]; ok {
		r.Font = v
	}

	return nil
}
