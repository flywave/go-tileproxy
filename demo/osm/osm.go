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
		Options:     &setting.ImageOpts{Format: "png"},
		Grid:        "global_webmercator",
	}
	osmTMSCache = setting.CacheSource{
		Sources:       []string{"tms"},
		Name:          "tms_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.LocalCache{
			Directory:       "./cache_data/osm",
			DirectoryLayout: "tms",
		},
	}
	osmService = setting.TMSService{
		Layers: []setting.TileLayer{
			{
				TileSource: "tms_cache",
				Name:       "tms_layer",
			},
		},
	}
)

func getProxyDataset() *setting.ProxyDataset {
	pd := setting.NewProxyDataset("osm")
	pd.Grids = demo.GridMap

	pd.Sources["tms"] = &osmTMSSource
	pd.Caches["tms_cache"] = &osmTMSCache

	pd.Service = &osmService
	return pd
}

func getDataset() *tileproxy.Dataset {
	return tileproxy.NewDataset(getProxyDataset(), "../", &demo.Globals)
}

var dataset *tileproxy.Dataset

func DatasetServer(w http.ResponseWriter, req *http.Request) {
	if dataset == nil {
		dataset = getDataset()
	}
	dataset.Service.ServeHTTP(w, req)
}

func main() {
	http.HandleFunc("/", DatasetServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
