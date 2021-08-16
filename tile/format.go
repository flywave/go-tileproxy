package tile

import "strings"

var (
	mimeMaps = map[string]string{
		"png":     "image/png",
		"tif":     "image/tiff",
		"tiff":    "image/tiff",
		"jpg":     "image/jpeg",
		"jpeg":    "image/jpeg",
		"webp":    "image/webp",
		"atm":     "image/lerc",
		"mvt":     "application/vnd.mapbox-vector-tile",
		"pbf":     "application/x-protobuf",
		"json":    "application/json",
		"geojson": "application/json",
		"terrain": "application/vnd.quantized-mesh,application/octet-stream;q=0.9",
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

	if ext == "vnd.mapbox-vector-tile" {
		ext = "mvt"
	}
	if ext == "application/x-protobuf" {
		ext = "pbf"
	}
	if ext == "application/json" {
		ext = "geojson"
	}
	if strings.Contains(ext, "application/vnd.quantized-mesh") {
		ext = "terrain"
	}

	return ext
}
