package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	MAPBOX_TILE_URL    = "https://a.tiles.mapbox.com"
	MAPBOX_ACCESSTOKEN = "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ"
)

var (
	mapboxRasterDemSource = setting.MapboxTileSource{
		Url:           MAPBOX_TILE_URL + "/raster/v1/mapbox.mapbox-terrain-dem-v1/{z}/{x}/{y}.webp",
		AccessToken:   MAPBOX_ACCESSTOKEN,
		Options:       &setting.ImageOpts{Format: "webp"},
		Grid:          "global_webmercator",
		TilejsonUrl:   MAPBOX_TILE_URL + "/v4/mapbox.mapbox-terrain-dem-v1.json",
		TilejsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
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
				Source:   "rasterdem_cache",
				Name:     "rasterdem_layer",
				TileJSON: "rasterdem",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("cesium")
	pd.Grids = demo.GridMap

	pd.Sources["rasterdem"] = &mapboxRasterDemSource
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

func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
