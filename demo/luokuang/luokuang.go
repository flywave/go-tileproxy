package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	LK_API_URL     = "https://api.luokuang.com"
	LK_ACCESSTOKEN = "{token}"
)

var (
	mapboxMVTSource = setting.MapboxTileSource{
		Url:             LK_API_URL + "/emg/v2/map/tile?format=pbf&layer=basic&style=main&zoom={z}&x={x}&y={y}",
		AccessToken:     LK_ACCESSTOKEN,
		AccessTokenName: "AK",
		Options:         &setting.VectorOpts{Format: "mvt", Extent: 4096},
		Grid:            "global_webmercator",
		TilejsonUrl:     LK_API_URL + "/view/map/lkstreetv2.json",
		TilejsonStore:   &setting.LocalStore{Directory: "./cache_data/tilejson/"},
	}
	mapboxMVTCache = setting.Caches{
		Sources:       []string{"mvt"},
		Name:          "mvt_cache",
		Grid:          "global_webmercator",
		Format:        "mvt",
		RequestFormat: "mvt",
		CacheInfo: &setting.LocalCache{
			Directory:       "./cache_data/mvt",
			DirectoryLayout: "tms",
			UseGridNames:    false,
		},
	}
	mapboxService = setting.MapboxService{
		Layers: []setting.MapboxTileLayer{
			{
				Source:   "mvt_cache",
				Name:     "mvt_layer",
				TileJSON: "mvt",
			},
		},
		Styles: []setting.StyleSource{
			{
				Url:              LK_API_URL + "/openplatform/v1/mapStyle/getStyle?styleId=standard&type=1",
				StyleID:          "standard",
				AccessTokenName:  "AK",
				AccessToken:      LK_ACCESSTOKEN,
				Store:            &setting.LocalStore{Directory: "./cache_data/style/"},
				Sprite:           LK_API_URL + "/emg/static/sprites/sprite",
				Glyphs:           LK_API_URL + "/emg/fonts/{fontstack}/{range}.pbf",
				GlyphsStore:      &setting.LocalStore{Directory: "./cache_data/glyphs/"},
				Fonts:            []string{"Noto Sans CJK SC DemiLight"},
				StyleContentAttr: setting.NewString("data.styleContent"),
			},
		},
	}
)

func getProxyDataset() *setting.ProxyDataset {
	pd := setting.NewProxyDataset("mapbox")
	pd.Grids = demo.GridMap

	pd.Sources["mvt"] = &mapboxMVTSource
	pd.Caches["mvt_cache"] = &mapboxMVTCache

	pd.Service = &mapboxService
	return pd
}

func getDataset() *tileproxy.Dataset {
	return tileproxy.NewDataset(getProxyDataset(), "../", &demo.Globals, nil)
}

var dataset *tileproxy.Dataset

func DatasetServer(w http.ResponseWriter, req *http.Request) {
	if dataset == nil {
		dataset = getDataset()
	}
	dataset.Service.ServeHTTP(w, req)
}

//https://api.luokuang.com/view/map/lkstreetv2.json
//https://api.luokuang.com/openplatform/v1/mapStyle/getStyle?styleId=nightblue&type=1&AK={token}
//https://api.luokuang.com/openplatform/v1/mapStyle/getStyle?styleId=standard&type=1&AK={token}
//https://api.luokuang.com/emg/static/sprites/sprite.json?ak={token}
//https://api.luokuang.com/emg/static/sprites/sprite.png?ak={token}
//https://api.luokuang.com/emg/static/sprites/sprite_2020_v2.png?ak={token}
//https://api.luokuang.com/emg/static/sprites/sprite_2020_v2.json?ak={token}
//https://api.luokuang.com/emg/fonts/Noto%20Sans%20CJK%20SC%20DemiLight/0-255.pbf?ak={token}
//https://api.luokuang.com/emg/fonts/Noto%20Sans%20CJK%20SC%20DemiLight,Arial%20Unicode%20MS%20Regular/0-255.pbf?ak={token}
//https://api.luokuang.com/emg/v1/map/tile?format=pbf&layer=basic&style=main&zoom=11&x=1687&y=775&ak={token}
//https://api.luokuang.com/emg/v2/map/tile?format=pbf&layer=basic&style=main&zoom=11&x=1687&y=775&ak={token}

//http://127.0.0.1:8000/v4/mvt_layer.json
//http://127.0.0.1:8000/v4/mvt_layer/11/1687/775.mvt
//http://127.0.0.1:8000/fonts/v1/Noto%20Sans%20CJK%20SC%20DemiLight/0-255.pbf
func main() {
	http.HandleFunc("/", DatasetServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
