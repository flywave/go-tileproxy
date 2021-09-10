package service

import (
	"os"
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/flywave/ogc-specifications/pkg/wms130"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/sources"
	"github.com/flywave/go-tileproxy/tile"
)

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

	newString := func(s string) *string {
		return &s
	}

	layerMetadata := &WMSLayerMetadata{}
	layerMetadata.AuthorityURL.Name = "flywave"
	layerMetadata.AuthorityURL.OnlineResource.Xlink = newString("http://flywave.net")
	layerMetadata.AuthorityURL.OnlineResource.Type = newString("simple")
	layerMetadata.AuthorityURL.OnlineResource.Href = newString("http://flywave.net")
	layerMetadata.Identifier.Authority = "flywave"
	layerMetadata.Identifier.Value = "flywave"
	layerMetadata.Style = append(layerMetadata.Style, &wms130.Style{Name: "flywave"})

	imageopts := &imagery.ImageOptions{Format: tile.TileFormat("png"), Resampling: "nearest"}

	source := sources.NewWMSSource(nil, imageopts, nil, nil, nil, nil, nil, nil, nil)

	testLayer := NewWMSNodeLayer("test", "hhh", map[string]layer.Layer{"yy": source}, nil, nil, nil, layerMetadata)

	rootLayer := NewWMSGroupLayer("test", "hhh", testLayer, nil, layerMetadata)

	srs := &geo.SupportedSRS{Srs: []geo.Proj{geo.NewProj(4326)}}

	srsExtents := make(map[string]*geo.MapExtent)

	srsExtents["EPSG:4326"] = &geo.MapExtent{Srs: geo.NewProj(4326), BBox: vec2d.Rect{Min: vec2d.T{-180, -90}, Max: vec2d.T{180, 90}}}

	capabilities := newCapabilities(service, rootLayer, []string{"image/png"}, []string{"image/png"}, srs, srsExtents, 2)

	xml := capabilities.render(nil)

	f, _ := os.Create("./test.xml")

	f.Write(xml)

	f.Close()

	if xml == nil {
		t.FailNow()
	}
}
