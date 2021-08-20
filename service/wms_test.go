package service

import (
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

func TestWMSLayer(t *testing.T) {

}

func TestWMSGroupLayer(t *testing.T) {

}

func TestWMSServiceGetMap(t *testing.T) {

}

func TestWMSServiceGetFeatureInfo(t *testing.T) {

}

func TestWMSServiceLegendgraphic(t *testing.T) {

}

func TestCombinedLayers(t *testing.T) {

}

func TestWMSServiceGetCapabilities(t *testing.T) {
	service := make(map[string]string)
	service["url"] = "http://flywave.net"
	service["title"] = "flywave"
	service["abstract"] = ""
	service["keyword_list"] = "helll,testhnh"
	service["online_resource"] = "http://flywave.net"

	service["contactinformation.contactpersonprimary.contactperson"] = "dd"
	service["contactinformation.contactpersonprimary.contactorganization"] = "flywave"
	service["contactinformation.contactposition"] = "flywave"
	service["contactinformation.contactaddress.addresstype"] = "flywave"
	service["contactinformation.contactaddress.address"] = "flywave"
	service["contactinformation.contactaddress.city"] = "flywave"
	service["contactinformation.contactaddress.stateorprovince"] = "flywave"
	service["contactinformation.contactaddress.postcode"] = "flywave"
	service["contactinformation.contactaddress.country"] = "flywave"
	service["contactinformation.contactvoicetelephone"] = "flywave"
	service["contactinformation.contactfacsimiletelephone"] = "flywave"
	service["contactinformation.contactelectronicmailaddress"] = "flywave"

	service["extendedcapabilities.metadataurl.type"] = "simple"
	service["extendedcapabilities.metadataurl.url"] = "http://flywave.net"
	service["extendedcapabilities.metadataurl.mediatype"] = "simple"
	service["supportedlanguages.defaultlanguage.language"] = "simple"
	service["responselanguage.language"] = "simple"

	layerMetadata := map[string]string{}
	layerMetadata["tilemetadata.authorityurl.name"] = "flywave"
	layerMetadata["tilemetadata.authorityurl.onlineresource.xlink"] = "http://flywave.net"
	layerMetadata["tilemetadata.authorityurl.onlineresource.type"] = "simple"
	layerMetadata["tilemetadata.authorityurl.onlineresource.href"] = "http://flywave.net"
	layerMetadata["tilemetadata.identifier.authority"] = "flywave"
	layerMetadata["tilemetadata.identifier.value"] = "flywave"
	layerMetadata["tilemetadata.style.name"] = "flywave"
	layerMetadata["tilemetadata.style.legend.width"] = "256"
	layerMetadata["tilemetadata.style.legend.height"] = "256"
	layerMetadata["tilemetadata.style.legend.url"] = "http://flywave.net"

	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)

	testLayer := NewWMSLayer("test", "hhh", map[string]layer.Layer{"yy": source}, nil, nil, nil, layerMetadata)

	rootLayer := NewWMSGroupLayer("test", "hhh", testLayer, nil, layerMetadata)

	srs := &geo.SupportedSRS{Srs: []geo.Proj{geo.NewSRSProj4("EPSG:4326")}}

	srsExtents := make(map[string]*geo.MapExtent)

	srsExtents["EPSG:4326"] = &geo.MapExtent{Srs: geo.NewSRSProj4("EPSG:4326"), BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}}

	capabilities := newCapabilities(service, rootLayer, []string{"image/png"}, []string{"image/png"}, srs, srsExtents, 2)

	xml := capabilities.render(nil)

	f, _ := os.Create("./test.xml")

	f.Write(xml)

	f.Close()

	if xml == nil {
		t.FailNow()
	}
}
