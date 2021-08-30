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
	LK_ACCESSTOKEN = "DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8"
)

var (
	mapboxMVTSource = setting.LuokuangTileSource{
		Url:         LK_API_URL,
		TilesetID:   "mapbox.mapbox-streets-v8",
		Version:     "v4",
		AccessToken: LK_ACCESSTOKEN,
		Options:     &setting.VectorOpts{Format: "mvt", Extent: 4096},
		Grid:        "global_webmercator",
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
				Source: "mvt_cache",
				Name:   "mvt_layer",
				TileJSON: setting.TileJSONSource{
					Url:         LK_API_URL + "/view/map/",
					AccessToken: LK_ACCESSTOKEN,
					TilesetID:   "lkstreetv2",
					Store:       &setting.LocalStore{Directory: "./cache_data/mvt/tilejson/"},
					IsLKMode:    true,
				},
			},
		},
		Styles: []setting.StyleSource{
			{
				Url:         LK_API_URL + "/openplatform/v1/mapStyle/getStyle",
				StyleID:     "standard",
				AccessToken: LK_ACCESSTOKEN,
				Store:       &setting.LocalStore{Directory: "./cache_data/style/"},
				IsLKMode:    true,
				Fonts: setting.GlyphsSource{
					Url:         LK_API_URL,
					AccessToken: LK_ACCESSTOKEN,
					Font:        "Arial Unicode MS Regular",
					Store:       &setting.LocalStore{Directory: "./cache_data/glyphs/"},
					IsLKMode:    true,
				},
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
//https://api.luokuang.com/openplatform/v1/mapStyle/getStyle?styleId=standard&type=1&AK=DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8
//https://api.luokuang.com/emg/static/sprites/sprite_2020_v2.png?ak=DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8
//https://api.luokuang.com/emg/static/sprites/sprite_2020_v2.json?ak=DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8
//https://api.luokuang.com/emg/fonts/Noto%20Sans%20CJK%20SC%20DemiLight,Arial%20Unicode%20MS%20Regular/0-255.pbf?ak=DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8
//https://api.luokuang.com/emg/v2/map/tile?format=pbf&layer=basic&style=main&zoom=10&x=844&y=388&ak=DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8

//http://127.0.0.1:8000/v4/pbf_layer.json
//http://127.0.0.1:8000/v4/pbf_layer/1/0/0.mvt
func main() {
	http.HandleFunc("/", DatasetServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
