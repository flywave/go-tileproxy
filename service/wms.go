package service

import (
	"github.com/flywave/go-tileproxy/layer"
	_ "github.com/flywave/ogc-specifications/pkg/wms130"
)

type WMSService struct {
	RootLayer WMSGroupLayer
}

func (s *WMSService) GetMap() {

}

func (s *WMSService) GetCapabilities() {

}

func (s *WMSService) GetFeatureInfo() {

}

func (s *WMSService) checkMapRequest() {

}

func (s *WMSService) checkFeatureinfoRequest() {

}

func (s *WMSService) validateLayers() {

}

func (s *WMSService) Legendgraphic() {

}

func (s *WMSService) authorizedTileLayers() {

}

func (s *WMSService) filterActualLayers() {

}

func (s *WMSService) authorizedCapabilityLayers() {

}

type FilteredRootLayer struct {
}

type WMSLayerBase struct{}

type WMSLayer struct {
	WMSLayerBase
}

type WMSGroupLayer struct {
	WMSLayerBase
}

type LayerRenderer struct {
}

func CombinedLayers(layers []*WMSLayerBase, query *layer.MapQuery) {

}
