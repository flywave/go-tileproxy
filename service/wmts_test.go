package service

import (
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/cache"
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
)

func TestWMTSServiceGetCapabilities(t *testing.T) {

}

func TestWMTSServiceGetTile(t *testing.T) {

}

func TestWMTSServiceGetFeatureInfo(t *testing.T) {

}

func TestWMTSRestService(t *testing.T) {

}

func TestWMTSCapabilities(t *testing.T) {

	service := make(map[string]string)
	service["url"] = "http://flywave.net"
	service["title"] = "flywave"
	service["abstract"] = ""

	service["serviceprovider.providername"] = "flywave"
	service["serviceprovider.providersite.type"] = "wms"
	service["serviceprovider.providersite.href"] = "http://flywave.net"
	service["serviceprovider.servicecontact.individualname"] = "test"
	service["serviceprovider.servicecontact.positionname"] = "test"
	service["serviceprovider.servicecontact.contactinfo.text"] = "test"
	service["serviceprovider.servicecontact.contactinfo.phone.voice"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.phone.facsimile"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.deliverypoint"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.city"] = "test"
	service["serviceprovider.servicecontact.contactinfo.address.administrativearea"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.postalcode"] = "1288766553"
	service["serviceprovider.servicecontact.contactinfo.address.country"] = "test"
	service["serviceprovider.servicecontact.contactinfo.address.electronicmailaddress"] = "test"
	service["serviceprovider.servicecontact.contactinfo.onlineresource.type"] = "test"
	service["serviceprovider.servicecontact.contactinfo.onlineresource.href"] = "http://flywave.net"
	service["serviceprovider.servicecontact.contactinfo.hoursofservice"] = "test"
	service["serviceprovider.servicecontact.contactinfo.contactinstructions"] = "test"
	service["serviceprovider.servicecontact.role"] = "test"

	opts := geo.DefaultTileGridOptions()
	opts[geo.TILEGRID_SRS] = "EPSG:4326"
	opts[geo.TILEGRID_BBOX] = vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}
	grid := geo.NewTileGrid(opts)

	tileset := NewTileMatrixSet(grid)

	layerMetadata := map[string]string{"name_path": "test"}

	info := []layer.Layer{}

	dimensions := make(utils.Dimensions)
	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)

	ccreater := func(location string) tile.Source {
		data, _ := os.ReadFile(location)
		s := imagery.CreateImageSourceFromBufer(data, imageopts)
		return s
	}
	locker := &cache.DummyTileLocker{}
	c := cache.NewLocalCache("./test_cache", "png", "quadkey", ccreater)

	manager := cache.NewTileManager([]layer.Layer{source}, grid, c, locker, "test", "png", imageopts, false, false, nil, -1, false, 0, [2]uint32{1, 1})

	tp := NewTileProvider("test", "test", layerMetadata, manager, info, dimensions)

	if tp == nil {
		t.FailNow()
	}

	layers := []WMTSTileLayer{NewWMTSTileLayer([]Provider{tp})}

	capabilities := newWMTSCapabilities(service, layers, map[string]*TileMatrixSet{"EPSG:4326": tileset}, nil)

	xml := capabilities.render(nil)

	f, _ := os.Create("./test.xml")

	f.Write(xml)

	f.Close()

	os.Remove("./test.xml")

	if xml == nil {
		t.FailNow()
	}
}
