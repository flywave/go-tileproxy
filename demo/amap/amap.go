package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	AMAP_API_URL = "https://webrd0.is.autonavi.com/appmaptile"
)

var (
	amapTMSSource = setting.TileSource{
		URLTemplate: AMAP_API_URL + "?x={x}&y={y}&z={z}&lang=zh_cn&size=1&scale=1&style=8",
		Grid:        "global_webmercator",
		Subdomains:  []string{"1", "2", "3", "4"},
		Options:     &setting.ImageOpts{Format: "png"},
	}

	amapTMSCache = setting.CacheSource{
		Sources:       []string{"amap"},
		Name:          "amap_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/amap",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}

	amapService = setting.TMSService{
		Layers: []setting.TileLayer{
			{
				Source: "amap_cache",
				Name:   "amap_layer",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("amap")
	pd.Grids = demo.GridMap

	pd.Sources["amap"] = &amapTMSSource
	pd.Caches["amap_cache"] = &amapTMSCache

	pd.Service = &amapService
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
	err := http.ListenAndServe(":8003", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
