package sources

import (
	"image/color"

	"github.com/flywave/go-tileproxy/maths"
)

type WMSSource struct {
	SupportsMetaTiles         bool
	SupportedSRS              string
	SupportedFormats          string
	ExtReqParams              map[string]string
	TransparentColor          color.Color
	TransparentColorTolerance float64
	Coverage                  maths.Coverage
	ResRange                  maths.ResolutionRange
	Extent                    maths.MapExtent
	ErrorHandler              func(error)
}

func (s *WMSSource) IsOpaque(query MapQuery) bool {
	return false
}

func (s *WMSSource) GetMap(query MapQuery) {

}

func (s *WMSSource) getMap(query MapQuery) {

}

func (s *WMSSource) getSubQuery(query MapQuery, format string) {

}

func (s *WMSSource) getTransformed(query MapQuery, format string) {

}

func (s *WMSSource) isCompatible(other *WMSSource, query MapQuery) {

}

func (s *WMSSource) CombinedLayer(other *WMSSource, query MapQuery) {

}
