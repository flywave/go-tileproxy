package layer

import (
	"errors"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/resource"
)

type Layer interface {
	GetMap(query *MapQuery) images.Source
}

type MapLayer struct {
	SupportMetaTiles bool
	ResRange         *geo.ResolutionRange
	Coverage         geo.Coverage
	Extent           *geo.MapExtent
	Options          *images.ImageOptions
}

func NewMapLayer(opts *images.ImageOptions) *MapLayer {
	return &MapLayer{Options: opts}
}

func (l *MapLayer) GetOpacity() float64 {
	return *l.Options.Opacity
}

func (l *MapLayer) SetOpacity(value float64) {
	l.Options.Opacity = geo.NewFloat64(value)
}

func (l *MapLayer) IsOpaque(query *MapQuery) bool {
	return false
}

func (l *MapLayer) CheckResRange(query *MapQuery) error {
	if l.ResRange != nil &&
		!l.ResRange.Contains(query.BBox, query.Size, query.Srs) {
		return errors.New("BlankImage")
	}
	return nil
}

func (l *MapLayer) CombinedLayer(other Layer, query *MapQuery) Layer {
	return nil
}

type LimitedLayer struct {
	layer    Layer
	coverage geo.Coverage
}

func (l *LimitedLayer) CombinedLayer(other Layer, query *MapQuery) Layer {
	return nil
}

func (l *LimitedLayer) GetInfo(query *InfoQuery) *resource.FeatureInfo {
	return nil
}

type DirectMapLayer struct {
}

type CacheMapLayer struct {
}
