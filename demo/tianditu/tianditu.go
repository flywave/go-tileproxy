package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	TIANDITU_API_URL = "http://t0.tianditu.com"
)

var (
	tiandituWMSSource = setting.TileSource{
		URLTemplate: TIANDITU_API_URL + "/{layer}/wmts?service=wmts&request=GetTile&version=1.0.0&LAYER={layer}&tileMatrixSet=w&TileMatrix={z}&TileRow={y}&TileCol={x}&style=default&format=tiles",
		Grid:        "global_geodetic_cgcs2000",
		Subdomains:  []string{"0", "1", "2", "3", "4", "5", "6"},
		Options:     &setting.ImageOpts{Format: "png"},
	}

	tiandituWMSCache = setting.CacheSource{
		Sources:       []string{"tianditu"},
		Name:          "tianditu_cache",
		Grid:          "global_geodetic_cgcs2000",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/tianditu",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}

	tiandituService = setting.WMTSService{
		Restful: setting.NewBool(true),
		Layers: []setting.TileLayer{
			{
				Source: "tianditu_cache",
				Name:   "vec_w",
			},
			{
				Source: "tianditu_cache",
				Name:   "cva_w",
			},
			{
				Source: "tianditu_cache",
				Name:   "img_w",
			},
			{
				Source: "tianditu_cache",
				Name:   "cia_w",
			},
			{
				Source: "tianditu_cache",
				Name:   "ter_w",
			},
			{
				Source: "tianditu_cache",
				Name:   "cta_w",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("tianditu")
	pd.Grids = demo.GridMap

	pd.Sources["tianditu"] = &tiandituWMSSource
	pd.Caches["tianditu_cache"] = &tiandituWMSCache

	pd.Service = &tiandituService
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
	err := http.ListenAndServe(":8007", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
