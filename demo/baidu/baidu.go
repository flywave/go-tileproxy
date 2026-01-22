package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	BAIDU_API_URL = "http://online3.map.bdimg.com/onlinelabel"
)

var (
	baiduTMSSource = setting.TileSource{
		URLTemplate: BAIDU_API_URL + "?qt=tile&x={x}&y={y}&z={z}&styles=pl&scaler=1&p=1",
		Grid:        "global_webmercator",
		Subdomains:  []string{"0", "1", "2", "3"},
		Options:     &setting.ImageOpts{Format: "png"},
	}

	baiduTMSCache = setting.CacheSource{
		Sources:       []string{"baidu"},
		Name:          "baidu_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/baidu",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}

	baiduService = setting.TMSService{
		Layers: []setting.TileLayer{
			{
				Source: "baidu_cache",
				Name:   "baidu_layer",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("baidu")
	pd.Grids = demo.GridMap

	pd.Sources["baidu"] = &baiduTMSSource
	pd.Caches["baidu_cache"] = &baiduTMSCache

	pd.Service = &baiduService
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
	err := http.ListenAndServe(":8006", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
