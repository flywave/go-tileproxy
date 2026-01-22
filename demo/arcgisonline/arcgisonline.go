package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	ARCGIS_API_URL = "https://services.arcgisonline.com/arcgis/rest/services"
)

var (
	arcgisImagerySource = setting.TileSource{
		URLTemplate: ARCGIS_API_URL + "/World_Imagery/MapServer/tile/{z}/{y}/{x}",
		Grid:        "global_webmercator",
		Options:     &setting.ImageOpts{Format: "png"},
	}

	arcgisStreetSource = setting.TileSource{
		URLTemplate: ARCGIS_API_URL + "/World_Street_Map/MapServer/tile/{z}/{y}/{x}",
		Grid:        "global_webmercator",
		Options:     &setting.ImageOpts{Format: "png"},
	}

	arcgisTopoSource = setting.TileSource{
		URLTemplate: ARCGIS_API_URL + "/World_Topo_Map/MapServer/tile/{z}/{y}/{x}",
		Grid:        "global_webmercator",
		Options:     &setting.ImageOpts{Format: "png"},
	}

	arcgisNatGeoSource = setting.TileSource{
		URLTemplate: ARCGIS_API_URL + "/NatGeo_World_Map/MapServer/tile/{z}/{y}/{x}",
		Grid:        "global_webmercator",
		Options:     &setting.ImageOpts{Format: "png"},
	}

	arcgisImageryCache = setting.CacheSource{
		Sources:       []string{"imagery"},
		Name:          "imagery_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/arcgis_imagery",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}

	arcgisStreetCache = setting.CacheSource{
		Sources:       []string{"street"},
		Name:          "street_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/arcgis_street",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.ImageOpts{Format: "png"},
	}

	arcgisService = setting.TMSService{
		Layers: []setting.TileLayer{
			{
				Source: "imagery_cache",
				Name:   "imagery",
			},
			{
				Source: "street_cache",
				Name:   "street",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("arcgis")
	pd.Grids = demo.GridMap

	pd.Sources["imagery"] = &arcgisImagerySource
	pd.Sources["street"] = &arcgisStreetSource
	pd.Sources["topo"] = &arcgisTopoSource
	pd.Sources["natgeo"] = &arcgisNatGeoSource

	pd.Caches["imagery_cache"] = &arcgisImageryCache
	pd.Caches["street_cache"] = &arcgisStreetCache

	pd.Service = &arcgisService
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
	err := http.ListenAndServe(":8008", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
