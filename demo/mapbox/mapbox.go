package mapbox

import (
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
	MapboxMVTSource = setting.MapboxTileSource{
		Url:         MAPBOX_API_URL,
		TilesetID:   "mapbox.mapbox-streets-v8",
		Version:     "v4",
		UserName:    MAPBOX_USERNAME,
		AccessToken: MAPBOX_ACCESSTOKEN,
		Options:     &setting.VectorOpts{Format: "mvt", Extent: 4096},
		Grid:        "global_webmercator",
	}
	MapboxRasterSource = setting.MapboxTileSource{
		Url:         MAPBOX_API_URL,
		TilesetID:   "mapbox.satellite",
		Version:     "v4",
		UserName:    MAPBOX_USERNAME,
		AccessToken: MAPBOX_ACCESSTOKEN,
		Options:     &setting.ImageOpts{Format: "png"},
		Grid:        "global_webmercator",
	}
	MapboxRasterDemSource = setting.MapboxTileSource{
		Url:         MAPBOX_API_URL,
		Version:     "v1",
		TilesetID:   "mapbox.mapbox-terrain-dem-v1",
		UserName:    MAPBOX_USERNAME,
		AccessToken: MAPBOX_ACCESSTOKEN,
		Options:     &setting.ImageOpts{Format: "webp"},
		Grid:        "global_webmercator",
	}

	MapboxMVTCache = setting.Caches{
		Sources: []string{"mvt"},
		Name:    "mvt_cache",
		Grid:    "global_webmercator",
		CacheInfo: &setting.LocalCache{
			Directory:       "./cache_data/mvt",
			DirectoryLayout: "tms",
			UseGridNames:    false,
		},
	}
	MapboxRasterCache = setting.Caches{
		Sources: []string{"raster"},
		Name:    "raster_cache",
		Grid:    "global_webmercator",
		CacheInfo: &setting.LocalCache{
			Directory:       "./cache_data/raster/",
			DirectoryLayout: "tms",
			UseGridNames:    false,
		},
	}
	MapboxRasterDemCache = setting.Caches{
		Sources: []string{"rasterdem"},
		Name:    "rasterdem_cache",
		Grid:    "global_webmercator",
		CacheInfo: &setting.LocalCache{
			Directory:       "./cache_data/rasterdem/",
			DirectoryLayout: "tms",
			UseGridNames:    false,
		},
	}

	MapboxService = setting.MapboxService{
		Layers: []setting.MapboxTileLayer{
			{
				Source: "mvt_cache",
				Name:   "mvt_layer",
				TileJSON: setting.TileJSONSource{
					Url:         MAPBOX_API_URL,
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
				StyleID:     "cjikt35x83t1z2rnxpdmjs7y7",
				AccessToken: MAPBOX_ACCESSTOKEN,
				Store:       &setting.LocalStore{Directory: "./cache_data/style/"},
			},
		},
		Fonts: []setting.GlyphsSource{
			{
				Url:         MAPBOX_API_URL,
				UserName:    MAPBOX_USERNAME,
				AccessToken: MAPBOX_ACCESSTOKEN,
				Font:        "Arial Unicode MS Regular",
				Store:       &setting.LocalStore{Directory: "./cache_data/glyphs/"},
			},
		},
	}
)

func GetProxyDataset() *setting.ProxyDataset {
	pd := setting.NewProxyDataset("mapbox")
	pd.Grids = demo.GridMap
	pd.Sources["mvt"] = &MapboxMVTSource
	pd.Sources["raster"] = &MapboxRasterSource
	pd.Sources["rasterdem"] = &MapboxRasterDemSource

	pd.Caches["mvt_cache"] = &MapboxMVTCache
	pd.Caches["raster_cache"] = &MapboxRasterCache
	pd.Caches["rasterdem_cache"] = &MapboxRasterDemCache

	pd.Service = &MapboxService
	return pd
}

func GetDataset() *tileproxy.Dataset {
	return tileproxy.NewDataset(GetProxyDataset(), "../", &demo.Globals, nil)
}
