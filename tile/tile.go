package tile

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	DefaultEpislon    = 10.0
	DefaultExtent     = 4096
	DefaultTileBuffer = 64.0
	MaxZ              = 22
)

var (
	TILE_EXTENSIONS = []string{".png", ".json", ".topojson", ".geojson", ".mvt", ".pbf", ".terrain", ".lerc"}
)

type Tile interface {
	SetStatus(isfail bool, errormsg string)
	SetHeader(header http.Header)
	GetHeader() http.Header
	SetCookies(cookies []*http.Cookie)
	GetCookies() []*http.Cookie
	Success() bool
	ErrorMsg() string
	//SetRequest(r *request.Request) Tile
	//GetRequest() *request.Request
	GetUrlTag() string
	SetBody(body []byte) Tile
	GetBody() []byte
}

/**
func NewTile(req *request.Request) Tile {
	return nil
}
**/
func IsFileExtTile(ext string) bool {
	ext = strings.ToLower(ext)
	for _, imgExt := range TILE_EXTENSIONS {
		if ext == imgExt {
			return true
		}
	}
	return false
}

func GetMimetype(ext string) (string, error) {
	switch ext {
	case ".png":
		return "image/png", nil
	case ".json":
		return "application/json", nil
	case ".topojson":
		return "application/json;type=topojson", nil
	case ".geojson":
		return "application/json;type=geojson", nil
	case ".mvt":
		return "application/vnd.mapbox-vector-tile", nil
	case ".pbf":
		return "application/x-protobuf;type=mapbox-vector;", nil
	case ".terrain":
		return "application/vnd.quantized-mesh", nil
	case ".lerc":
		return "application/octet-stream", nil
	default:
		return "", fmt.Errorf("unknown mimetype for extension \"%v\"", ext)
	}
}
