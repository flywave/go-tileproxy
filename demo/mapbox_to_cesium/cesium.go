package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	MAPBOX_TILE_URL    = "https://api.mapbox.com"
	MAPBOX_ACCESSTOKEN = "pk.eyJ1IjoiaGF3a2luZzIyMTUiLCJhIjoiY2lqcDB3OGFnMDEwbXRva292emduaXhpcSJ9.QwZ6QKL_shDTOYNuVTYbUw"
)

var (
	unilateral string = "unilateral"
	bilateral  string = "bilateral"

	mapboxRasterDemSource = setting.MapboxTileSource{
		MapboxTileSourcePart: setting.MapboxTileSourcePart{
			Url:           MAPBOX_TILE_URL + "/raster/v1/mapbox.mapbox-terrain-dem-v1/{z}/{x}/{y}.webp",
			AccessToken:   MAPBOX_ACCESSTOKEN,
			Sku:           "101XxiLvoFYxL",
			Grid:          "global_webmercator_512",
			TilejsonUrl:   MAPBOX_TILE_URL + "/v4/mapbox.mapbox-terrain-dem-v1.json",
			TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"}},
		Options: &setting.RasterOpts{Format: "webp", Mode: &bilateral},
	}
	mapboxRasterDemCache = setting.CacheSource{
		CacheSourcePart: setting.CacheSourcePart{
			Sources:       []string{"rasterdem"},
			Name:          "rasterdem_cache",
			Grid:          "global_webmercator_512",
			Format:        "webp",
			RequestFormat: "webp",
			CacheInfo: &setting.CacheInfo{
				Directory:       "./cache_data/rasterdem/",
				DirectoryLayout: "tms",
			},
			QueryBuffer: setting.NewInt(1),
		},
		TileOptions: &setting.RasterOpts{Format: "webp", Mode: &bilateral},
	}
	cesiumTerrainCache = setting.CacheSource{
		CacheSourcePart: setting.CacheSourcePart{
			Sources:       []string{"rasterdem_cache"},
			Name:          "terrain_cache",
			Grid:          "global_geodetic_sw",
			Format:        "terrain",
			RequestFormat: "terrain",
			CacheInfo: &setting.CacheInfo{
				Directory:       "./cache_data/terrain/",
				DirectoryLayout: "tms",
			}},
		TileOptions: &setting.RasterOpts{Format: "terrain", Mode: &unilateral},
	}
	cesiumService = setting.CesiumService{
		Layers: []setting.CesiumTileLayer{
			{
				Source:    "terrain_cache",
				Name:      "terrain_layer",
				LayerJSON: "terrain",
				ZoomRange: &[2]int{0, 13},
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("cesium")
	pd.Grids = demo.GridMap

	pd.Sources["rasterdem"] = &mapboxRasterDemSource
	pd.Caches["rasterdem_cache"] = &mapboxRasterDemCache
	pd.Caches["terrain_cache"] = &cesiumTerrainCache

	pd.Service = &cesiumService

	return pd
}

func getService() *tileproxy.Service {
	return tileproxy.NewService(getProxyService(), "../", &demo.Globals, nil)
}

var dataset *tileproxy.Service

func ProxyServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if dataset == nil {
		dataset = getService()
	}
	dataset.Service.ServeHTTP(w, req)
}

func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
