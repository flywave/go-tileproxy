package demo

import "github.com/flywave/go-tileproxy/setting"

var (
	GridMap = map[string]setting.GridOpts{}
)

func init() {
	GridMap["global_geodetic"] = setting.GridOpts{Name: "GLOBAL_GEODETIC", Srs: "EPSG:4326", Origin: "sw"}
	GridMap["global_geodetic_sqrt2"] = setting.GridOpts{Name: "GLOBAL_GEODETIC", Srs: "EPSG:4326", Origin: "sw", ResFactor: "sqrt2"}
	GridMap["global_mercator"] = setting.GridOpts{Name: "GLOBAL_MERCATOR", Srs: "EPSG:900913", Origin: "sw"}
	GridMap["global_webmercator"] = setting.GridOpts{Name: "GLOBAL_WEB_MERCATOR", Srs: "EPSG:3857", Origin: "nw"}
}