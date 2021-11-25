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
	MAPBOX_ACCESSTOKEN = "{token}"
)

var (
	mapboxMVTSource = setting.MapboxTileSource{
		Url:           MAPBOX_TILE_URL + "/v4/mapbox.mapbox-streets-v8/{z}/{x}/{y}.vector.pbf",
		AccessToken:   MAPBOX_ACCESSTOKEN,
		Options:       &setting.VectorOpts{Format: "mvt", Extent: 4096},
		Grid:          "global_webmercator",
		TilejsonUrl:   MAPBOX_TILE_URL + "/v4/mapbox.mapbox-streets-v8.json",
		TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
	}
	mapboxRasterSource = setting.MapboxTileSource{
		Url:           MAPBOX_TILE_URL + "/v4/mapbox.satellite/{z}/{x}/{y}.png",
		AccessToken:   MAPBOX_ACCESSTOKEN,
		Options:       &setting.ImageOpts{Format: "png"},
		Grid:          "global_webmercator",
		TilejsonUrl:   MAPBOX_TILE_URL + "/v4/mapbox.satellite.json",
		TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
	}
	mapboxRasterDemSource = setting.MapboxTileSource{
		Url:           MAPBOX_TILE_URL + "/raster/v1/mapbox.mapbox-terrain-dem-v1/{z}/{x}/{y}.webp",
		AccessToken:   MAPBOX_ACCESSTOKEN,
		Options:       &setting.ImageOpts{Format: "webp"},
		Grid:          "global_webmercator",
		TilejsonUrl:   MAPBOX_TILE_URL + "/v4/mapbox.mapbox-terrain-dem-v1.json",
		TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
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
		Styles: []setting.MapboxStyleLayer{
			{
				Url:         MAPBOX_API_URL + "/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7",
				StyleID:     "style",
				AccessToken: MAPBOX_ACCESSTOKEN,
				Store:       &setting.StoreInfo{Directory: "./cache_data/style/"},
				Sprite:      MAPBOX_API_URL + "/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7/sprite",
				Glyphs:      MAPBOX_API_URL + "/fonts/v1/examples/{fontstack}/{range}.pbf",
				GlyphsStore: &setting.StoreInfo{Directory: "./cache_data/glyphs/"},
				Fonts:       []string{"Arial Unicode MS Regular"},
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

//http://127.0.0.1:8000/v4/mvt_layer.json
//http://127.0.0.1:8000/v4/mvt_layer/1/0/0.mvt
//http://127.0.0.1:8000/v4/raster_layer.json
//http://127.0.0.1:8000/v4/raster_layer/1/0/0.png
//http://127.0.0.1:8000/v4/rasterdem_layer.json
//http://127.0.0.1:8000/v4/rasterdem_layer/14/13733/6366.webp
//http://127.0.0.1:8000/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7
//http://127.0.0.1:8000/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7/sprite@3x
//http://127.0.0.1:8000/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7/sprite.png
//http://127.0.0.1:8000/fonts/v1/examples/Arial%20Unicode%20MS%20Regular/0-255.pbf
func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
