package demo

import "github.com/flywave/go-tileproxy/setting"

var (
	GridMap = map[string]setting.GridOpts{}
)

func init() {
	GridMap["global_geodetic"] = setting.GridOpts{Name: "GLOBAL_GEODETIC", Srs: "EPSG:4326", Origin: "ul"}
	GridMap["global_geodetic_sw"] = setting.GridOpts{Name: "GLOBAL_GEODETIC", Srs: "EPSG:4326", Origin: "sw", TileSize: &[2]uint32{512, 512}}
	GridMap["global_geodetic_root"] = setting.GridOpts{Name: "GLOBAL_GEODETIC", Srs: "EPSG:4326", Origin: "sw", InitialResMin: setting.NewBool(true)}
	GridMap["global_geodetic_sqrt2"] = setting.GridOpts{Name: "GLOBAL_GEODETIC", Srs: "EPSG:4326", Origin: "sw", ResFactor: "sqrt2"}
	GridMap["global_mercator"] = setting.GridOpts{Name: "GLOBAL_MERCATOR", Srs: "EPSG:900913", Origin: "sw"}
	GridMap["global_webmercator"] = setting.GridOpts{Name: "GLOBAL_WEB_MERCATOR", Srs: "EPSG:3857", Origin: "ul"}
	GridMap["global_geodetic_cgcs2000"] = setting.GridOpts{Name: "GLOBAL_GEODETIC_CGCS2000", Srs: "EPSG:4490", Origin: "sw"}
	GridMap["global_geodetic_cgcs2000_root"] = setting.GridOpts{Name: "GLOBAL_GEODETIC_CGCS2000", Srs: "EPSG:4490", Origin: "sw", InitialResMin: setting.NewBool(true)}
	GridMap["global_mercator_cgcs2000"] = setting.GridOpts{Name: "GLOBAL_MERCATOR_CGCS2000", Srs: "EPSG:4479", Origin: "sw"}
	GridMap["global_webmercator_512"] = setting.GridOpts{Name: "GLOBAL_WEB_MERCATOR_512", Srs: "EPSG:3857", Origin: "ul", TileSize: &[2]uint32{512, 512}}

}
