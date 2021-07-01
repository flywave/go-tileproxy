package sources

import (
	"image/color"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
)

type WMSSource struct {
	SupportsMetaTiles         bool
	SupportedSRS              string
	SupportedFormats          string
	ExtReqParams              map[string]string
	TransparentColor          color.Color
	TransparentColorTolerance float64
	Coverage                  geo.Coverage
	ResRange                  geo.ResolutionRange
	Extent                    geo.MapExtent
	ErrorHandler              func(error)
}

func (s *WMSSource) IsOpaque(query layer.MapQuery) bool {
	return false
}

func (s *WMSSource) GetMap(query layer.MapQuery) {

}

func (s *WMSSource) getMap(query layer.MapQuery) {

}

func (s *WMSSource) getSubQuery(query layer.MapQuery, format string) {

}

func (s *WMSSource) getTransformed(query layer.MapQuery, format string) {

}

func (s *WMSSource) isCompatible(other *WMSSource, query layer.MapQuery) {

}

func (s *WMSSource) CombinedLayer(other *WMSSource, query layer.MapQuery) {

}
