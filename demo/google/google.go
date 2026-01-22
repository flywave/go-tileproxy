package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	GOOGLE_API_URL = "https://mt.googleapis.com/vt"
)

var (
	googleTMSSource = setting.TileSource{
		URLTemplate: GOOGLE_API_URL + "?x={x}&y={y}&z={z}",
		Grid:        "global_webmercator",
		Subdomains:  []string{"0", "1", "2", "3"},
		Options:     &setting.ImageOpts{Format: "png"},
	}

	googleTMSCache = setting.CacheSource{
		Sources:       []string{"google"},
		Name:          "google_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/google",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}

	googleService = setting.TMSService{
		Layers: []setting.TileLayer{
			{
				Source: "google_cache",
				Name:   "google_layer",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("google")
	pd.Grids = demo.GridMap

	pd.Sources["google"] = &googleTMSSource
	pd.Caches["google_cache"] = &googleTMSCache

	pd.Service = &googleService
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
	err := http.ListenAndServe(":8004", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
