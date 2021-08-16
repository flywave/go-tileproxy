package service

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/request"
)

type WMSCapabilities struct {
	service         map[string]string
	root_layer      *WMSGroupLayer
	tile_layers     []*TileProvider
	imageFormats    map[string]*imagery.ImageOptions
	info_formats    []string
	srs             *geo.SupportedSRS
	srsExtents      map[string]*geo.MapExtent
	inspireMetadata map[string]string
	maxOutputPixels int
}

func (c *WMSCapabilities) layerSrsBBox(layer *TileProvider, epsg_axis_order bool) map[string]vec2d.Rect {
	ret := make(map[string]vec2d.Rect)
	for srs, extent := range c.srsExtents {
		if !geo.SrcInProj(srs, c.srs.Srs) {
			continue
		}
		var bbox vec2d.Rect
		if extent.IsDefault() {
			bbox = layer.extent.BBoxFor(geo.NewSRSProj4(srs))
		} else if layer.extent.IsDefault() {
			bbox = extent.BBoxFor(geo.NewSRSProj4(srs))
		} else {
			a := extent.Transform(geo.NewSRSProj4("EPSG:4326"))
			b := layer.extent.Transform(geo.NewSRSProj4("EPSG:4326"))
			bbox = a.Intersection(b).BBoxFor(geo.NewSRSProj4(srs))
		}

		if epsg_axis_order {
			bbox = request.SwitchBBoxEpsgAxisOrder(bbox, srs)
		}

		ret[srs] = bbox
	}

	layer_srs_code := layer.extent.Srs.GetSrsCode()
	if _, ok := c.srsExtents[layer_srs_code]; ok {
		bbox := layer.extent.BBox
		if epsg_axis_order {
			bbox = request.SwitchBBoxEpsgAxisOrder(bbox, layer_srs_code)
		}
		if !geo.SrcInProj(layer_srs_code, c.srs.Srs) {
			ret[layer_srs_code] = bbox
		}
	}
	return ret
}

func (c *WMSCapabilities) layerLLBBox(layer *TileProvider) vec2d.Rect {
	if srs, ok := c.srsExtents["EPSG:4326"]; ok {
		llbbox := srs.Intersection(layer.extent).GetLLBBox()
		return limit_llbbox(llbbox)
	}
	return limit_llbbox(layer.extent.GetLLBBox())
}

func limit_llbbox(bbox vec2d.Rect) vec2d.Rect {
	minx, miny, maxx, maxy := bbox.Min[0], bbox.Min[1], bbox.Max[0], bbox.Max[1]
	minx = math.Max(-180, minx)
	miny = math.Max(-89.999999, miny)
	maxx = math.Min(180, maxx)
	maxy = math.Min(89.999999, maxy)
	return vec2d.Rect{Min: vec2d.T{minx, miny}, Max: vec2d.T{maxx, maxy}}
}

var (
	DEFAULT_EXTENTS = map[string]*geo.MapExtent{
		"EPSG:3857":   geo.MapExtentFromDefault(),
		"EPSG:4326":   geo.MapExtentFromDefault(),
		"EPSG:900913": geo.MapExtentFromDefault(),
	}
)

func copyMapExtents(extent map[string]*geo.MapExtent) map[string]*geo.MapExtent {
	ret := make(map[string]*geo.MapExtent)
	for k := range extent {
		ret[k] = extent[k]
	}
	return ret
}

func limit_srs_extents(srs_extents map[string]*geo.MapExtent, supported_srs *geo.SupportedSRS) map[string]*geo.MapExtent {
	if srs_extents != nil {
		srs_extents = copyMapExtents(srs_extents)
	} else {
		srs_extents = copyMapExtents(DEFAULT_EXTENTS)
	}

	for srs := range srs_extents {
		notin := true
		for i := range supported_srs.Srs {
			if supported_srs.Srs[i].GetSrsCode() == srs {
				notin = false
			}
		}
		if !notin {
			delete(srs_extents, srs)
		}
	}

	return srs_extents
}
