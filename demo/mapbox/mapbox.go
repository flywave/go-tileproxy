package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	MAPBOX_API_URL     = "https://api.mapbox.com"
	MAPBOX_TILE_URL    = "https://a.tiles.mapbox.com"
	MAPBOX_ACCESSTOKEN = "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja3pjOXRqcWkybWY3MnVwaGxkbTgzcXAwIn0._tCv9fpOyCT4O_Tdpl6h0w"
)

var (
	mapboxMVTSource = setting.MapboxTileSource{
		Url:           MAPBOX_TILE_URL + "/v4/mapbox.mapbox-streets-v8.json",
		AccessToken:   MAPBOX_ACCESSTOKEN,
		Grid:          "global_webmercator",
		TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
		Options:       &setting.VectorOpts{Format: "mvt", Extent: 4096},
	}
	mapboxRasterSource = setting.MapboxTileSource{
		Url:           MAPBOX_TILE_URL + "/v4/mapbox.satellite.json",
		AccessToken:   MAPBOX_ACCESSTOKEN,
		Grid:          "global_webmercator",
		TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
		Options:       &setting.ImageOpts{Format: "png"},
	}
	mapboxRasterDemSource = setting.MapboxTileSource{
		Url:           MAPBOX_TILE_URL + "/v4/mapbox.mapbox-terrain-dem-v1.json",
		AccessToken:   MAPBOX_ACCESSTOKEN,
		Sku:           "101XxiLvoFYxL",
		Grid:          "global_webmercator",
		TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
		Options:       &setting.RasterOpts{Format: "webp"},
	}

	mapboxMVTCache = setting.CacheSource{
		Sources:       []string{"mvt"},
		Name:          "mvt_cache",
		Grid:          "global_webmercator",
		Format:        "mvt",
		RequestFormat: "mvt",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/mvt",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.VectorOpts{Format: "mvt", Extent: 4096},
	}

	mapboxRasterCache = setting.CacheSource{
		Sources:       []string{"raster"},
		Name:          "raster_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/raster/",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}
	mapboxRasterDemCache = setting.CacheSource{
		Sources:       []string{"rasterdem"},
		Name:          "rasterdem_cache",
		Grid:          "global_webmercator",
		Format:        "webp",
		RequestFormat: "webp",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/rasterdem/",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.RasterOpts{Format: "webp"},
	}
	mapboxService = setting.MapboxService{
		Layers: []setting.MapboxTileLayer{
			{
				Source:   "mvt_cache",
				Name:     "mvt_layer",
				TileJSON: "mvt",
			},
			{
				Source:   "raster_cache",
				Name:     "raster_layer",
				TileJSON: "raster",
			},
			{
				Source:   "rasterdem_cache",
				Name:     "rasterdem_layer",
				TileJSON: "rasterdem",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("mapbox")
	pd.Grids = demo.GridMap

	pd.Sources["mvt"] = &mapboxMVTSource
	pd.Sources["raster"] = &mapboxRasterSource
	pd.Sources["rasterdem"] = &mapboxRasterDemSource

	pd.Caches["mvt_cache"] = &mapboxMVTCache
	pd.Caches["raster_cache"] = &mapboxRasterCache
	pd.Caches["rasterdem_cache"] = &mapboxRasterDemCache

	pd.Service = &mapboxService
	return pd
}

func getService() *tileproxy.Service {
	return tileproxy.NewService(getProxyService(), "../", &demo.Globals, nil)
}

var dataset *tileproxy.Service

func ProxyServer(w http.ResponseWriter, req *http.Request) {
	if dataset == nil {
		dataset = getService()
	}
	dataset.Service.ServeHTTP(w, req)
}

// http://127.0.0.1:8000/mvt_layer/source.json
// http://127.0.0.1:8000/mvt_layer/1/0/0.mvt
// http://127.0.0.1:8000/raster_layer/source.json
// http://127.0.0.1:8000/raster_layer/1/0/0.png
// http://127.0.0.1:8000/rasterdem_layer/source.json
// http://127.0.0.1:8000/rasterdem_layer/14/13733/6366.webp
func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
