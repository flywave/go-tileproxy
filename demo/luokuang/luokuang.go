package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
	"github.com/flywave/go-tileproxy/vector"
)

const (
	LK_API_URL     = "https://api.luokuang.com"
	LK_ACCESSTOKEN = "DE16394725636226419F0AB3765FFA4D2EB27A998BD259465DGXZ8TJUZKE2067"
)

var (
	GridMap = map[string]setting.GridOpts{}
)

func init() {
	GridMap["global_webmercator"] = setting.GridOpts{Name: "GLOBAL_WEB_MERCATOR", Srs: "EPSG:3857", Origin: "ul", TileSize: &[2]uint32{4096, 4096}}
}

var (
	luokuangMVTSource = setting.MapboxTileSource{
		MapboxTileSourcePart: setting.MapboxTileSourcePart{
			Url:             LK_API_URL + "/emg/v2/map/tile?format=pbf&layer=basic&style=main&zoom={z}&x={x}&y={y}",
			AccessToken:     LK_ACCESSTOKEN,
			AccessTokenName: "ak",
			Grid:            "global_webmercator",
			TilejsonUrl:     LK_API_URL + "/view/map/lkstreetv2.json",
			TilejsonStore:   &setting.StoreInfo{Directory: "./cache_data/tilejson/"}},
		Options: &setting.VectorOpts{Format: "mvt", Extent: 4096, Proto: setting.NewInt(int(vector.PBF_PTOTO_LUOKUANG))},
	}
	luokuangMVTCache = setting.CacheSource{
		CacheSourcePart: setting.CacheSourcePart{
			Sources:       []string{"lk_mvt"},
			Name:          "lk_mvt_cache",
			Grid:          "global_webmercator",
			Format:        "mvt",
			RequestFormat: "mvt",
			CacheInfo: &setting.CacheInfo{
				Directory:       "./cache_data/lk_mvt",
				DirectoryLayout: "tms",
			},
			ReprojectSrs: &setting.Reproject{SrcSrs: "EPSG:GCJ02", DestSrs: "EPSG:4326"}},
		TileOptions: &setting.VectorOpts{Format: "mvt", Extent: 4096, Proto: setting.NewInt(int(vector.PBF_PTOTO_LUOKUANG))},
	}
	mapboxMVTCache = setting.CacheSource{
		CacheSourcePart: setting.CacheSourcePart{
			Sources:       []string{"lk_mvt_cache"},
			Name:          "mvt_cache",
			Grid:          "global_webmercator",
			Format:        "mvt",
			RequestFormat: "mvt",
			CacheInfo: &setting.CacheInfo{
				Directory:       "./cache_data/mvt",
				DirectoryLayout: "tms",
			}},
		TileOptions: &setting.VectorOpts{Format: "mvt", Extent: 4096, Proto: setting.NewInt(int(vector.PBF_PTOTO_MAPBOX))},
	}
	mapboxService = setting.MapboxService{
		Layers: []setting.MapboxTileLayer{
			{
				Source:   "mvt_cache",
				Name:     "mvt_layer",
				TileJSON: "mvt",
			},
		},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("mapbox")
	pd.Grids = GridMap

	pd.Sources["lk_mvt"] = &luokuangMVTSource
	pd.Caches["mvt_cache"] = &mapboxMVTCache
	pd.Caches["lk_mvt_cache"] = &luokuangMVTCache

	pd.Service = &mapboxService
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

//http://127.0.0.1:8000/v4/mvt_layer.json
//http://127.0.0.1:8000/v4/mvt_layer/11/1687/775.mvt
//http://127.0.0.1:8000/fonts/v1/Noto%20Sans%20CJK%20SC%20DemiLight/0-255.pbf
func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
