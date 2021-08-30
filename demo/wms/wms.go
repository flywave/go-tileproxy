package wms

var (
	base_url  = "https://wms.geo.admin.ch"
	srs       = "EPSG:21781"
	center    = []float64{8.23, 46.86}
	centersrs = "EPSG:4326"
	extent    = []float64{420000, 30000, 900000, 350000}
	extentsrs = "EPSG:21781"
	layers    = []string{
		"ch.swisstopo.pixelkarte-farbe-pk1000.noscale",
		"ch.bafu.hydroweb-warnkarte_national",
	}
)

//https://wms.geo.admin.ch/?SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&FORMAT=image%2Fjpeg&TRANSPARENT=true&LAYERS=ch.swisstopo.pixelkarte-farbe-pk1000.noscale&WIDTH=256&HEIGHT=256&CRS=EPSG%3A21781&STYLES=&BBOX=705373.9428000001%2C124338.29039999997%2C749274.8168%2C168239.16439999998
//https://wms.geo.admin.ch/?SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&FORMAT=image%2Fpng&TRANSPARENT=true&LAYERS=ch.bafu.hydroweb-warnkarte_national&WIDTH=256&HEIGHT=256&CRS=EPSG%3A21781&STYLES=&BBOX=793175.6908%2C212140.03839999996%2C837076.5647999999%2C256040.91239999997
