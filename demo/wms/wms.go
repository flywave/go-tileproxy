package main

import (
	"log"
	"net/http"

	"github.com/flywave/go-tileproxy"
	"github.com/flywave/go-tileproxy/demo"
	"github.com/flywave/go-tileproxy/setting"
)

var (
	wmsTMSSource = setting.WMSSource{
		Opts: setting.WMSSourceOpts{
			Version: "1.1.1",
			Map:     setting.NewBool(true),
		},
		Url:          "https://maps.omniscale.net/v2/demo/style.default/map?",
		Layers:       []string{"osm"},
		SupportedSrs: []string{"EPSG:31467", "EPSG:4326"},
	}

	wmsService = setting.WMSService{
		Srs:                []string{"EPSG:4326", "EPSG:900913", "EPSG:25832"},
		BBoxSrs:            []setting.BBoxSrs{{Srs: "EPSG:4326"}},
		ImageFormats:       []string{"image/jpeg", "image/png", "image/gif", "image/GeoTIFF", "image/tiff"},
		Layers:             []setting.WMSLayer{{MapSources: []string{"osm_wms"}, Name: "osm", Title: "Omniscale OSM WMS - osm.omniscale.net"}},
		MaxOutputPixels:    setting.NewInt(2000 * 2000),
		Strict:             setting.NewBool(true),
		FeatureinfoFormats: []setting.FeatureinfoFormat{{Suffix: "text", MimeType: "text/plain"}, {Suffix: "html", MimeType: "text/html"}, {Suffix: "xml", MimeType: "text/xml"}},
	}
)

func getProxyService() *setting.ProxyService {
	pd := setting.NewProxyService("wms")
	pd.Grids = demo.GridMap

	pd.Sources["osm_wms"] = &wmsTMSSource

	pd.Service = &wmsService
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

// https://maps.omniscale.net/v2/demo/style.default/map?LAYERS=osm&FORMAT=image/png&SERVICE=WMS&VERSION=1.1.1&SRS=EPSG:31467&REQUEST=GetMap&BBOX=3648808.05872,5599105.55872,3669808.05872,5621605.55872&WIDTH=420&HEIGHT=450
// http://127.0.0.1:8000/?SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&FORMAT=image%2Fjpeg&TRANSPARENT=true&LAYERS=osm&WIDTH=256&HEIGHT=256&CRS=EPSG%3A21781&STYLES=&BBOX=705373.9428000001%2C124338.29039999997%2C749274.8168%2C168239.16439999998

func main() {
	http.HandleFunc("/", ProxyServer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
