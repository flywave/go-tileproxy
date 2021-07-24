package tile

import "strings"

var (
	mimeMaps = map[string]string{
		"png":  "image/png",
		"tif":  "image/tiff",
		"tiff": "image/tiff",
		"jpe":  "image/jpeg",
		"jpeg": "image/jpeg",
		"wbmp": "image/vnd.wap.wbmp",
		"mvt":  "application/vnd.mapbox-vector-tile",
	}
)

type TileFormat string

func (i *TileFormat) MimeType() string {
	if strings.Contains(string(*i), "/") {
		return string(*i)
	}
	return mimeMaps[string(*i)]
}

func (i *TileFormat) Extension() string {
	ext := string(*i)
	if strings.Contains(ext, "/") {
		ext = strings.Split(ext, "/")[1]
	}
	if strings.Contains(ext, ";") {
		ext = strings.Split(ext, ";")[0]
	}
	return strings.Trim(ext, " ")
}
