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
		"lerc": "image/lerc",
		"mvt":  "application/vnd.mapbox-vector-tile",
		"pbf":  "application/x-protobuf",
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
	ext = strings.Trim(ext, " ")

	if ext == "vnd.wap.wbmp" {
		ext = "wbmp"
	}
	if ext == "vnd.mapbox-vector-tile" {
		ext = "mvt"
	}
	if ext == "application/x-protobuf" {
		ext = "pbf"
	}

	return ext
}
