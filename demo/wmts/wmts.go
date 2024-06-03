package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

var (
	subdomains            = []int{1, 2, 3, 4}
	wmts_capabilities_url = "https://basemap.at/wmts/1.0.0/WMTSCapabilities.xml"
	center                = []float64{1823849, 6143760}
	centerSrs             = "EPSG:3857"
	layers                = []string{"geolandbasemap", "bmapoverlay", "bmapgrau", "bmaphidpi", "bmaporthofoto30cm", "bmapgelaende", "bmapoberflaeche"}
)

const (
	WMTS_URL              = "https://maps{subdomains}.wien.gv.at/basemap/geolandbasemap"
	WMTS_CAPABILITIES_URL = "https://basemap.at/wmts/1.0.0/WMTSCapabilities.xml"
)

var (
	wmtsTMSSource = setting.TileSource{
		URLTemplate: WMTS_URL + "/normal/google3857/{tms_path}.png",
		Grid:        "global_webmercator",
		Subdomains:  []string{"1", "2", "3", "4"},
		Options:     &setting.ImageOpts{Format: "png"},
	}
	wmtsTMSCache = setting.CacheSource{
		Sources:       []string{"wmts"},
		Name:          "wmts_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/wmts",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}
	wmtsService = setting.WMTSService{
		Restful: setting.NewBool(true),
		Layers: []setting.TileLayer{
			{
				Source: "wmts_cache",
				Name:   "wmts_layer",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("wmts")
	pd.Grids = demo.GridMap

	pd.Sources["wmts"] = &wmtsTMSSource
	pd.Caches["wmts_cache"] = &wmtsTMSCache

	pd.Service = &wmtsService
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

// https://maps2.wien.gv.at/basemap/geolandbasemap/normal/google3857/11/710/1117.png
// http://127.0.0.1:8000?service=WMTS&request=GetTile&version=1.0.0&layer=wmts_layer&style=&format=image/png&TileMatrixSet=GLOBAL_WEB_MERCATOR&TileMatrix=11&TileRow=1117&TileCol=710
func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
