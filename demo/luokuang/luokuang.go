package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	LK_API_URL     = "https://api.luokuang.com"
	LK_ACCESSTOKEN = "{token}"
)

var (
	mapboxMVTSource = setting.MapboxTileSource{
		Url:             LK_API_URL + "/emg/v2/map/tile?format=pbf&layer=basic&style=main&zoom={z}&x={x}&y={y}",
		AccessToken:     LK_ACCESSTOKEN,
		AccessTokenName: "AK",
		Options:         &setting.VectorOpts{Format: "mvt", Extent: 4096},
		Grid:            "global_webmercator",
		TilejsonUrl:     LK_API_URL + "/view/map/lkstreetv2.json",
		TilejsonStore:   &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
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
	mapboxService = setting.MapboxService{
		Layers: []setting.MapboxTileLayer{
			{
				Source:   "mvt_cache",
				Name:     "mvt_layer",
				TileJSON: "mvt",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("mapbox")
	pd.Grids = demo.GridMap

	pd.Sources["mvt"] = &mapboxMVTSource
	pd.Caches["mvt_cache"] = &mapboxMVTCache

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
//http://127.0.0.1:8000/v4/mvt_layer/11/1687/775.mvt
//http://127.0.0.1:8000/fonts/v1/Noto%20Sans%20CJK%20SC%20DemiLight/0-255.pbf
func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
