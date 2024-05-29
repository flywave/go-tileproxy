package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	CESIUM_AUTH_URL    = "https://api.cesium.com"
	CESIUM_ASSETS_URL  = "https://assets.cesium.com"
	CESIUM_ACCESSTOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIzNWFkYjc1Ny04ODZkLTQ5NmItOTMyYy05ZWQyZTkwYWVlNjIiLCJpZCI6NTk3NTYsImlhdCI6MTY0NDIwMjEzNH0.iUeNB_nAKhVIq5LMwQjrNlV1LFko_N5DQeLUk7e-naw"
)

var (
	cesiumTerrainDemSource = setting.CesiumTileSource{
		AuthUrl:        CESIUM_AUTH_URL,
		Url:            CESIUM_ASSETS_URL,
		AssetId:        1,
		Version:        "1.2.0",
		AccessToken:    CESIUM_ACCESSTOKEN,
		Grid:           "global_geodetic",
		LayerjsonStore: &setting.StoreInfo{Directory: "./cache_data/tilejson/"},
		Options:        &setting.RasterOpts{Format: "terrain"},
	}
	cesiumTerrainDemCache = setting.CacheSource{
		Sources:       []string{"terrain"},
		Name:          "terrain_cache",
		Grid:          "global_geodetic",
		Format:        "terrain",
		RequestFormat: "terrain",
		CacheInfo: &setting.CacheInfo{
			Directory:       "./cache_data/terrain/",
			DirectoryLayout: "tms",
		},
		TileOptions: &setting.RasterOpts{Format: "terrain"},
	}
	cesiumService = setting.CesiumService{
		Layers: []setting.CesiumTileLayer{
			{
				Source:    "terrain_cache",
				Name:      "terrain_layer",
				LayerJSON: "terrain",
				ZoomRange: &[2]int{0, 13},
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("cesium")
	pd.Grids = demo.GridMap

	pd.Sources["terrain"] = &cesiumTerrainDemSource
	pd.Caches["terrain_cache"] = &cesiumTerrainDemCache

	pd.Service = &cesiumService
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

// http://127.0.0.1:8000/terrain_layer/layer.json
// http://127.0.0.1:8000/terrain_layer/14/13733/6366.terrain?extensions=octvertexnormals-metadata&v=1.2.0
func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
