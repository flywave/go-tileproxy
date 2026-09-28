package tile

import "strings"

var (
	mimeMaps = map[string]string{
		"png":     "image/png",
		"gif":     "image/gif",
		"tif":     "image/tiff",
		"tiff":    "image/tiff",
		"jpg":     "image/jpeg",
		"jpeg":    "image/jpeg",
		"webp":    "image/webp",
		"atm":     "image/lerc",
		"mvt":     "application/vnd.mapbox-vector-tile",
		"pbf":     "application/x-protobuf",
		"omv":     "application/x-protobuf",
		"json":    "application/json",
		"geojson": "application/json",
		"terrain": "application/vnd.quantized-mesh",
	}
)

var (
	ImageryFormats = []string{"png", "jpg", "jpeg", "webp", "tif", "tiff", "gif"}
	VectorFormats  = []string{"mvt", "pbf", "omv", "json", "geojson"}
	TerrainFormats = []string{"terrain", "atm"}
)

func IsImageryFormat(format string) bool { return IsFormatSupported(format, ImageryFormats) }
func IsVectorFormat(format string) bool  { return IsFormatSupported(format, VectorFormats) }
func IsTerrainFormat(format string) bool { return IsFormatSupported(format, TerrainFormats) }
func IsKnownFormat(format string) bool {
	return IsImageryFormat(format) || IsVectorFormat(format) || IsTerrainFormat(format)
}

func IsFormatSupported(format string, allowed []string) bool {
	ext := TileFormat(format).Extension()
	for _, f := range allowed {
		if f == ext {
			return true
		}
	}
	return false
}

type TileFormat string

func (i TileFormat) MimeType() string {
	s := string(i)
	if strings.Contains(s, "/") {
		return s
	}
	if m, ok := mimeMaps[s]; ok {
		return m
	}
	return s
}

func (i TileFormat) Extension() string {
	ext := string(i)
	ext = strings.Trim(ext, " ")
	if strings.Contains(ext, "vnd.mapbox-vector-tile") {
		return "mvt"
	}
	if strings.Contains(ext, "application/x-protobuf") {
		return "pbf"
	}
	if strings.Contains(ext, "application/json") {
		return "geojson"
	}
	if strings.Contains(ext, "application/vnd.quantized-mesh") {
		return "terrain"
	}

	if strings.Contains(ext, "/") {
		ext = strings.Split(ext, "/")[1]
	}
	if strings.Contains(ext, ";") {
		ext = strings.Split(ext, ";")[0]
	}
	return ext
}
