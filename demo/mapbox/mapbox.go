package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

const (
	MAPBOX_API_URL     = "https://api.mapbox.com"
	MAPBOX_USERNAME    = "examples"
	MAPBOX_ACCESSTOKEN = "pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ"
)

var (
	mapboxMVTSource = setting.MapboxTileSource{
		Url:         MAPBOX_API_URL,
		TilesetID:   "mapbox.mapbox-streets-v8",
		Version:     "v4",
		UserName:    MAPBOX_USERNAME,
		AccessToken: MAPBOX_ACCESSTOKEN,
		Options:     &setting.VectorOpts{Format: "mvt", Extent: 4096},
		Grid:        "global_webmercator",
	}
	mapboxRasterSource = setting.MapboxTileSource{
		Url:         MAPBOX_API_URL,
		TilesetID:   "mapbox.satellite",
		Version:     "v4",
		UserName:    MAPBOX_USERNAME,
		AccessToken: MAPBOX_ACCESSTOKEN,
		Options:     &setting.ImageOpts{Format: "png"},
		Grid:        "global_webmercator",
	}
	mapboxRasterDemSource = setting.MapboxTileSource{
		Url:         MAPBOX_API_URL,
		Version:     "v1",
		Layer:       setting.NewString("raster"),
		TilesetID:   "mapbox.mapbox-terrain-dem-v1",
		UserName:    MAPBOX_USERNAME,
		AccessToken: MAPBOX_ACCESSTOKEN,
		Options:     &setting.ImageOpts{Format: "webp"},
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
	mapboxRasterCache = setting.Caches{
		Sources:       []string{"raster"},
		Name:          "raster_cache",
		Grid:          "global_webmercator",
		Format:        "png",
		RequestFormat: "png",
		CacheInfo: &setting.LocalCache{
			Directory:       "./cache_data/raster/",
			DirectoryLayout: "tms",
			UseGridNames:    false,
		},
	}
	mapboxRasterDemCache = setting.Caches{
		Sources:       []string{"rasterdem"},
		Name:          "rasterdem_cache",
		Grid:          "global_webmercator",
		Format:        "webp",
		RequestFormat: "webp",
		CacheInfo: &setting.LocalCache{
			Directory:       "./cache_data/rasterdem/",
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
					Url:         MAPBOX_API_URL,
					Version:     "v4",
					UserName:    MAPBOX_USERNAME,
					AccessToken: MAPBOX_ACCESSTOKEN,
					TilesetID:   "mapbox.mapbox-streets-v8",
					Store:       &setting.LocalStore{Directory: "./cache_data/mvt/tilejson/"},
				},
			},
			{
				Source: "raster_cache",
				Name:   "raster_layer",
				TileJSON: setting.TileJSONSource{
					Url:         MAPBOX_API_URL,
					Version:     "v4",
					UserName:    MAPBOX_USERNAME,
					AccessToken: MAPBOX_ACCESSTOKEN,
					TilesetID:   "mapbox.satellite",
					Store:       &setting.LocalStore{Directory: "./cache_data/raster/tilejson/"},
				},
			},
			{
				Source: "rasterdem_cache",
				Name:   "rasterdem_layer",
				TileJSON: setting.TileJSONSource{
					Url:         MAPBOX_API_URL,
					Version:     "v4",
					UserName:    MAPBOX_USERNAME,
					AccessToken: MAPBOX_ACCESSTOKEN,
					TilesetID:   "mapbox.mapbox-terrain-dem-v1",
					Store:       &setting.LocalStore{Directory: "./cache_data/rasterdem/tilejson/"},
				},
			},
		},
		Styles: []setting.StyleSource{
			{
				Url:         MAPBOX_API_URL,
				UserName:    MAPBOX_USERNAME,
				Version:     "v1",
				StyleID:     "cjikt35x83t1z2rnxpdmjs7y7",
				AccessToken: MAPBOX_ACCESSTOKEN,
				Store:       &setting.LocalStore{Directory: "./cache_data/style/"},
			},
		},
		Fonts: []setting.GlyphsSource{
			{
				Url:         MAPBOX_API_URL,
				UserName:    MAPBOX_USERNAME,
				Version:     "v1",
				AccessToken: MAPBOX_ACCESSTOKEN,
				Font:        "Arial Unicode MS Regular",
				Store:       &setting.LocalStore{Directory: "./cache_data/glyphs/"},
			},
		},
	}
)

func getProxyDataset() *setting.ProxyDataset {
	pd := setting.NewProxyDataset("mapbox")
	pd.Grids = demo.GridMap

	pd.Sources["mvt"] = &mapboxMVTSource
	pd.Sources["raster"] = &mapboxRasterSource
	pd.Sources["rasterdem"] = &mapboxRasterDemSource

	pd.Caches["mvt_cache"] = &mapboxMVTCache
	pd.Caches["raster_cache"] = &mapboxRasterCache
	pd.Caches["rasterdem_cache"] = &mapboxRasterDemCache

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

//http://127.0.0.1:8000/v4/mvt_layer.json
//http://127.0.0.1:8000/v4/mvt_layer/1/0/0.mvt
//http://127.0.0.1:8000/v4/raster_layer/1/0/0.png
//http://127.0.0.1:8000/v4/rasterdem_layer/14/13733/6366.webp
//http://127.0.0.1:8000/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7
//http://127.0.0.1:8000/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7/sprite@3x
//http://127.0.0.1:8000/styles/v1/examples/cjikt35x83t1z2rnxpdmjs7y7/sprite.png
//http://127.0.0.1:8000/fonts/v1/examples/Arial%20Unicode%20MS%20Regular/0-255.pbf
func main() {
	http.HandleFunc("/", DatasetServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
