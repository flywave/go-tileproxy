package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	OSM_API_URL = "https://tile.openstreetmap.org"
)

var (
	osmTMSSource = setting.TileSource{
		URLTemplate: OSM_API_URL + "/{tms_path}.png",
		Grid:        "global_webmercator",
		Options:     &setting.ImageOpts{Format: "png"},
	}

	osmTMSCache = setting.CacheSource{
		Sources:       []string{"tms"},
		Name:          "tms_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/osm",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}

	osmService = setting.TMSService{
		Layers: []setting.TileLayer{
			{
				Source: "tms_cache",
				Name:   "tms_layer",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("osm")
	pd.Grids = demo.GridMap

	pd.Sources["tms"] = &osmTMSSource
	pd.Caches["tms_cache"] = &osmTMSCache

	pd.Service = &osmService
	return pd
}

func getService() *tileproxy.Service {
	return tileproxy.NewService(getProxyService(), &demo.Globals, nil)
}

var dataset *tileproxy.Service

func ProxyServer(w http.ResponseWriter, req *http.Request) {
	if dataset == nil {
		dataset = getService()
	}
	dataset.Service.ServeHTTP(w, req)
}

// http://127.0.0.1:8001/tiles/tms_layer/0/0/0.png
// http://127.0.0.1:8001/tms/tms_layer/0/0/0.png
func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
