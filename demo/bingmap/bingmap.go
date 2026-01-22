package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	BING_API_URL = "https://ecn.t0.tiles.virtualearth.net/tiles"
)

var (
	bingTMSSource = setting.TileSource{
		URLTemplate: BING_API_URL + "/a{quadkey}.jpeg?g=1",
		Grid:        "global_webmercator",
		Subdomains:  []string{"0", "1", "2", "3"},
		Options:     &setting.ImageOpts{Format: "jpeg"},
	}

	bingTMSCache = setting.CacheSource{
		Sources:       []string{"bing"},
		Name:          "bing_cache",
		Grid:          "global_webmercator",
		Format:        "jpeg",
		RequestFormat: "jpeg",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/bing",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "jpeg"},
	}

	bingService = setting.TMSService{
		Layers: []setting.TileLayer{
			{
				Source: "bing_cache",
				Name:   "bing_layer",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("bing")
	pd.Grids = demo.GridMap

	pd.Sources["bing"] = &bingTMSSource
	pd.Caches["bing_cache"] = &bingTMSCache

	pd.Service = &bingService
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

func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8005", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
