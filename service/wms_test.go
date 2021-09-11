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
	service := &WMSMetadata{}
	service.URL = "http://flywave.net"
	service.Title = "flywave"
	service.Abstract = ""
	service.KeywordList = []string{"helll", "testhnh"}
	service.OnlineResource.Href = &service.URL

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
