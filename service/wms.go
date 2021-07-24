package service

import (
	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/layer"
	"github.com/flywave/go-tileproxy/tile"
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

type WMSCapabilities struct {
}

type FilteredRootLayer struct {
}

type LayerRenderer struct {
	layers []wmsLayer
	query  *layer.MapQuery
}

type wmsLayer interface {
	layer.Layer
	mapLayersForQuery(query *layer.MapQuery) []wmsLayer
	infoLayersForQuery(query *layer.InfoQuery) []wmsLayer
	legend(query *layer.LegendQuery) []tile.Source
	GetLegendSize() int
	GetName() string
	GetLegendUrl() string
	HasLegend() bool
	Queryable() bool
}

type WMSLayerBase struct {
	wmsLayer
	name       string
	title      string
	isActive   bool
	layers     []wmsLayer
	metadata   map[string]string
	queryable  bool
	hasLegend  bool
	legendUrl  string
	legendSize []int
	resRange   *geo.ResolutionRange
	extent     *geo.MapExtent
}

type WMSLayer struct {
	WMSLayerBase
	mapLayers    []wmsLayer
	infoLayers   []wmsLayer
	legendLayers []wmsLayer
}

func NewWMSLayer(name string, title string, map_layers []wmsLayer, infos []wmsLayer, legends []wmsLayer, res_range *geo.ResolutionRange, md map[string]string) *WMSLayer {
	queryable := false
	if len(infos) > 0 {
		queryable = true
	}

	has_legend := false
	if len(legends) > 0 {
		has_legend = true
	}

	ret := &WMSLayer{WMSLayerBase: WMSLayerBase{name: name, title: title, metadata: md, isActive: false, layers: nil, hasLegend: has_legend, queryable: queryable}, mapLayers: map_layers, infoLayers: infos, legendLayers: legends}

	ret.extent = mergeLayerExtents(map_layers)
	if res_range == nil {
		ret.resRange = mergeLayerResRanges(map_layers)
	} else {
		ret.resRange = res_range
	}

	return ret
}

func (l *WMSLayer) IsOpaque(query *layer.MapQuery) bool {
	for i := range l.mapLayers {
		if l.mapLayers[i].IsOpaque(query) {
			return true
		}
	}
	return false
}

func (l *WMSLayer) rendersQuery(query *layer.MapQuery) bool {
	if l.resRange != nil && !l.resRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}
	return true
}

func (l *WMSLayer) mapLayersForQuery(query *layer.MapQuery) []wmsLayer {
	if l.mapLayers == nil {
		return nil
	}
	return l.mapLayers
}

func (l *WMSLayer) infoLayersForQuery(query *layer.InfoQuery) []wmsLayer {
	if l.infoLayers == nil {
		return nil
	}
	return l.infoLayers
}

func (l *WMSLayer) legend(query *layer.LegendQuery) []tile.Source {
	legend := []tile.Source{}
	for _, lyr := range l.legendLayers {
		legend = append(legend, lyr.legend(query)...)
	}
	return legend
}

type WMSGroupLayer struct {
	WMSLayerBase
	this wmsLayer
}

func mergeLayerResRanges(layers []wmsLayer) *geo.ResolutionRange {
	ranges := []*geo.ResolutionRange{}
	for _, l := range layers {
		ranges = append(ranges, l.GetResolutionRange())
	}
	var ret *geo.ResolutionRange
	if len(ranges) > 0 {
		ret = ranges[0]
		for _, r := range ranges[1:] {
			ret = geo.MergeResolutionRange(ret, r)
		}
	}

	return ret
}

func mergeLayerExtents(layers []wmsLayer) *geo.MapExtent {
	if layers == nil || len(layers) == 0 {
		return geo.MapExtentFromDefault()
	}
	extent := layers[0].GetExtent()
	layers = layers[1:]
	for _, layer := range layers {
		extent = extent.Add(layer.GetExtent())
	}
	return extent
}

func NewWMSGroupLayer(name string, title string, this wmsLayer, layers []wmsLayer, md map[string]string) *WMSGroupLayer {
	is_active := false
	if this != nil {
		is_active = true
	}
	has_legend := false
	if this != nil && this.HasLegend() {
		has_legend = true
	} else {
		for _, l := range layers {
			if l.HasLegend() {
				has_legend = true
			}
		}
	}
	queryable := false
	if this != nil && this.Queryable() {
		queryable = true
	} else {
		for _, l := range layers {
			if l.Queryable() {
				queryable = true
			}
		}
	}

	ret := &WMSGroupLayer{WMSLayerBase: WMSLayerBase{name: name, title: title, metadata: md, isActive: is_active, layers: layers, hasLegend: has_legend, queryable: queryable}, this: this}

	all_layers := append(layers, this)
	ret.extent = mergeLayerExtents(all_layers)
	ret.resRange = mergeLayerResRanges(all_layers)

	return ret
}

func (l *WMSGroupLayer) IsOpaque(query *layer.MapQuery) bool {
	for i := range l.layers {
		if l.layers[i].IsOpaque(query) {
			return true
		}
	}
	return false
}

func (l *WMSGroupLayer) GetLegendSize() int {
	return l.this.GetLegendSize()
}

func (l *WMSGroupLayer) GetLegendUrl() string {
	return l.this.GetLegendUrl()
}

func (l *WMSGroupLayer) rendersQuery(query *layer.MapQuery) bool {
	if l.resRange != nil && !l.resRange.Contains(query.BBox, query.Size, query.Srs) {
		return false
	}
	return true
}

func (l *WMSGroupLayer) mapLayersForQuery(query *layer.MapQuery) []wmsLayer {
	if l.this != nil {
		return l.this.mapLayersForQuery(query)
	} else {
		layers := []wmsLayer{}
		for _, layer := range l.layers {
			layers = append(layers, layer.mapLayersForQuery(query)...)
		}
		return layers
	}
}

func (l *WMSGroupLayer) infoLayersForQuery(query *layer.InfoQuery) []wmsLayer {
	if l.this != nil {
		return l.this.infoLayersForQuery(query)
	} else {
		layers := []wmsLayer{}
		for _, layer := range l.layers {
			layers = append(layers, layer.infoLayersForQuery(query)...)
		}
		return layers
	}
}

func (l *WMSGroupLayer) GetChildLayers() map[string]wmsLayer {
	layers := make(map[string]wmsLayer)
	if l.name != "" {
		layers[l.name] = l
	}
	for _, lyr := range l.layers {
		if iface, ok := lyr.(interface {
			GetChildLayers() map[string]wmsLayer
		}); ok {
			new := iface.GetChildLayers()
			for k, v := range new {
				layers[k] = v
			}
		} else if lyr.GetName() != "" {
			layers[lyr.GetName()] = lyr
		}
	}
	return layers
}

func (l *WMSGroupLayer) legend(query *layer.LegendQuery) []tile.Source {
	panic("not implemented")
}

func CombinedLayers(layers []layer.Layer, query *layer.MapQuery) []layer.Layer {
	if len(layers) <= 1 {
		return layers
	}
	combined_layers := []layer.Layer{layers[0]}
	layers = layers[1:]
	for i := range layers {
		current_layer := layers[i]
		combined := combined_layers[len(combined_layers)-1].CombinedLayer(current_layer, query)
		if combined != nil {
			combined_layers[len(combined_layers)-1] = combined
		} else {
			combined_layers = append(combined_layers, current_layer)
		}
	}
	return combined_layers
}
