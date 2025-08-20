package service

import (
	"encoding/xml"
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/ogc-osgeo/pkg/wms130"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/request"
)

type WMSCapabilities struct {
	service         *WMSMetadata
	rootLayer       *WMSGroupLayer
	imageFormats    []string
	infoFormats     []string
	srs             *geo.SupportedSRS
	srsExtents      map[string]*geo.MapExtent
	maxOutputPixels int
}

func newCapabilities(service *WMSMetadata, root_layer *WMSGroupLayer, imageFormats []string, info_formats []string, srs *geo.SupportedSRS, srsExtents map[string]*geo.MapExtent, maxOutputPixels int) *WMSCapabilities {
	return &WMSCapabilities{
		service:         service,
		rootLayer:       root_layer,
		imageFormats:    imageFormats,
		infoFormats:     info_formats,
		srs:             srs,
		srsExtents:      srsExtents,
		maxOutputPixels: maxOutputPixels,
	}
}

func (c *WMSCapabilities) layerSrsBBox(layer WMSLayer, epsgAxisOrder bool) map[string]vec2d.Rect {
	ret := make(map[string]vec2d.Rect)
	for srs, extent := range c.srsExtents {
		if !geo.SrcInProj(srs, c.srs.Srs) {
			continue
		}
		var bbox vec2d.Rect
		if extent.IsDefault() {
			bbox = layer.GetExtent().BBoxFor(geo.NewProj(srs))
		} else if layer.GetExtent().IsDefault() {
			bbox = extent.BBoxFor(geo.NewProj(srs))
		} else {
			a := extent.Transform(geo.NewProj(4326))
			b := layer.GetExtent().Transform(geo.NewProj(4326))
			bbox = a.Intersection(b).BBoxFor(geo.NewProj(srs))
		}

		if epsgAxisOrder {
			bbox = request.SwitchBBoxEpsgAxisOrder(bbox, srs)
		}

		ret[srs] = bbox
	}

	layer_srs_code := layer.GetExtent().Srs.GetSrsCode()
	if _, ok := c.srsExtents[layer_srs_code]; ok {
		bbox := layer.GetExtent().BBox
		if epsgAxisOrder {
			bbox = request.SwitchBBoxEpsgAxisOrder(bbox, layer_srs_code)
		}
		if !geo.SrcInProj(layer_srs_code, c.srs.Srs) {
			ret[layer_srs_code] = bbox
		}
	}
	return ret
}

func (c *WMSCapabilities) layerLLBBox(layer WMSLayer) vec2d.Rect {
	if srs, ok := c.srsExtents["EPSG:4326"]; ok {
		llbbox := srs.Intersection(layer.GetExtent()).GetLLBBox()
		return limitLLBBox(llbbox)
	}
	return limitLLBBox(layer.GetExtent().GetLLBBox())
}

func limitLLBBox(bbox vec2d.Rect) vec2d.Rect {
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

func limitSrsExtents(srs_extents map[string]*geo.MapExtent, supported_srs *geo.SupportedSRS) map[string]*geo.MapExtent {
	if srs_extents != nil {
		srs_extents = copyMapExtents(srs_extents)
	} else {
		srs_extents = copyMapExtents(DEFAULT_EXTENTS)
	}

	for srs := range srs_extents {
		in := false
		for i := range supported_srs.Srs {
			if supported_srs.Srs[i].GetSrsCode() == srs {
				in = true
				break
			}
		}
		if !in {
			delete(srs_extents, srs)
		}
	}

	return srs_extents
}

func (c *WMSCapabilities) render(req *request.WMSRequest) []byte {
	cam := &wms130.GetCapabilitiesResponse{}
	cam.Namespaces.XmlnsWMS = "http://www.opengis.net/wms"
	cam.Namespaces.XmlnsSLD = "http://www.opengis.net/sld"
	cam.Namespaces.XmlnsXlink = "http://www.w3.org/1999/xlink"
	cam.Namespaces.XmlnsXSI = "http://www.w3.org/2001/XMLSchema-instance"
	cam.Namespaces.XmlnsInspireCommon = "http://inspire.ec.europa.eu/schemas/common/1.0"
	cam.Namespaces.XmlnsInspireVs = "http://inspire.ec.europa.eu/schemas/inspire_vs/1.0"
	cam.Namespaces.SchemaLocation = "http://www.opengis.net/wms http://schemas.opengis.net/wms/1.3.0/capabilities_1_3_0.xsd http://www.opengis.net/sld http://schemas.opengis.net/sld/1.1.0/sld_capabilities.xsd"
	cam.Namespaces.Version = "1.0.0"

	service := &cam.WMSService
	service.Name = "WMS"
	service.Title = c.service.Title
	service.Abstract = c.service.Abstract

	// 初始化KeywordList
	if len(c.service.KeywordList) > 0 {
		service.KeywordList = &wms130.Keywords{}
		service.KeywordList.Keyword = append(service.KeywordList.Keyword, c.service.KeywordList...)
	}

	url := c.service.URL

	// 初始化OnlineResource
	service.OnlineResource = wms130.OnlineResource{}
	if c.service.OnlineResource.Href != nil {
		service.OnlineResource.Href = c.service.OnlineResource.Href
	} else {
		service.OnlineResource.Href = &url
	}
	if c.service.OnlineResource.Type != nil {
		service.OnlineResource.Type = c.service.OnlineResource.Type
	}
	if c.service.OnlineResource.Xlink != nil {
		service.OnlineResource.Xlink = c.service.OnlineResource.Xlink
	}

	// 初始化ContactInformation
	if c.service.Contact != nil {
		service.ContactInformation = *c.service.Contact
	} else {
		service.ContactInformation = wms130.ContactInformation{}
	}

	// 设置费用和约束
	if c.service.Fees != nil {
		service.Fees = *c.service.Fees
	} else {
		service.Fees = "none"
	}
	if c.service.AccessConstraints != nil {
		service.AccessConstraints = *c.service.AccessConstraints
	} else {
		service.AccessConstraints = "none"
	}

	// 初始化可选约束
	service.OptionalConstraints = wms130.OptionalConstraints{}
	if c.maxOutputPixels > 0 {
		output_size := int(math.Sqrt(float64(c.maxOutputPixels)))
		service.OptionalConstraints.MaxWidth = output_size
		service.OptionalConstraints.MaxHeight = output_size
	}

	capabilities := &cam.Capabilities.WMSCapabilities

	// 初始化请求结构体
	capabilities.Request = wms130.Request{}
	capabilities.Request.GetCapabilities = wms130.RequestType{}
	capabilities.Request.GetCapabilities.Format = []string{"text/xml"}
	capabilities.Request.GetCapabilities.DCPType = &wms130.DCPType{}
	capabilities.Request.GetCapabilities.DCPType.HTTP = struct {
		Get  *wms130.Method `xml:"Get" yaml:"get"`
		Post *wms130.Method `xml:"Post" yaml:"post"`
	}{
		Get:  &wms130.Method{},
		Post: &wms130.Method{},
	}
	capabilities.Request.GetCapabilities.DCPType.HTTP.Get = &wms130.Method{}
	capabilities.Request.GetCapabilities.DCPType.HTTP.Get.OnlineResource = wms130.OnlineResource{}
	capabilities.Request.GetCapabilities.DCPType.HTTP.Get.OnlineResource.Xlink = &url

	capabilities.Request.GetMap = wms130.RequestType{}
	if len(c.imageFormats) > 0 {
		capabilities.Request.GetMap.Format = c.imageFormats[:]
	}
	capabilities.Request.GetMap.DCPType = &wms130.DCPType{}
	capabilities.Request.GetMap.DCPType.HTTP = struct {
		Get  *wms130.Method `xml:"Get" yaml:"get"`
		Post *wms130.Method `xml:"Post" yaml:"post"`
	}{
		Get:  &wms130.Method{},
		Post: &wms130.Method{},
	}
	capabilities.Request.GetMap.DCPType.HTTP.Get = &wms130.Method{}
	capabilities.Request.GetMap.DCPType.HTTP.Get.OnlineResource = wms130.OnlineResource{}
	capabilities.Request.GetMap.DCPType.HTTP.Get.OnlineResource.Xlink = &url

	if len(c.infoFormats) > 0 {
		capabilities.Request.GetFeatureInfo = &wms130.RequestType{}
		capabilities.Request.GetFeatureInfo.Format = c.infoFormats[:]
	} else {
		capabilities.Request.GetFeatureInfo = &wms130.RequestType{}
	}
	capabilities.Request.GetFeatureInfo.DCPType = &wms130.DCPType{}
	capabilities.Request.GetFeatureInfo.DCPType.HTTP = struct {
		Get  *wms130.Method `xml:"Get" yaml:"get"`
		Post *wms130.Method `xml:"Post" yaml:"post"`
	}{
		Get:  &wms130.Method{},
		Post: &wms130.Method{},
	}
	capabilities.Request.GetFeatureInfo.DCPType.HTTP.Get = &wms130.Method{}
	capabilities.Request.GetFeatureInfo.DCPType.HTTP.Get.OnlineResource = wms130.OnlineResource{}
	capabilities.Request.GetFeatureInfo.DCPType.HTTP.Get.OnlineResource.Xlink = &url

	capabilities.Exception = wms130.ExceptionType{}
	capabilities.Exception.Format = []string{"XML", "INIMAGE", "BLANK"}

	if c.service.Extended != nil {
		capabilities.ExtendedCapabilities = c.service.Extended
	}

	if c.rootLayer != nil {
		// 收集所有要处理的图层
		layersToProcess := []WMSLayer{}

		// 如果rootLayer有This字段，优先处理
		if c.rootLayer.this != nil {
			layersToProcess = append(layersToProcess, c.rootLayer.this)
		}

		// 处理Layers映射中的图层
		if c.rootLayer.layers != nil {
			for _, l := range c.rootLayer.layers {
				if l != nil {
					// 避免重复添加this图层
					if c.rootLayer.this == nil || l.GetName() != c.rootLayer.this.GetName() {
						layersToProcess = append(layersToProcess, l)
					}
				}
			}
		}

		// 处理所有收集到的图层
		for _, l := range layersToProcess {
			layer := wms130.Layer{}
			name := l.GetName()
			if name != "" {
				layer.Name = &name
			}
			layer.Title = l.GetTitle()
			layer.Queryable = geo.NewInt(1)

			metadata := l.GetMetadata()
			if metadata != nil {
				if metadata.Abstract != "" {
					layer.Abstract = metadata.Abstract
				}
				if metadata.KeywordList != nil {
					layer.KeywordList = metadata.KeywordList
				}
			}

			if c.srs != nil && len(c.srs.Srs) > 0 {
				for i := range c.srs.Srs {
					if c.srs.Srs[i] == nil {
						continue
					}
					SrsCode := c.srs.Srs[i].GetSrsCode()
					epsg_num := geo.GetEpsgNum(SrsCode)
					if epsg_num > 0 {
						crs := wms130.CRS{Code: epsg_num}
						layer.CRS = append(layer.CRS, crs)
					}
				}
			}

			bbox := c.layerLLBBox(l)
			if len(bbox.Min) >= 2 && len(bbox.Max) >= 2 {
				gbbox := &wms130.EXGeographicBoundingBox{}
				gbbox.WestBoundLongitude = bbox.Min[0]
				gbbox.EastBoundLongitude = bbox.Max[0]
				gbbox.SouthBoundLatitude = bbox.Min[1]
				gbbox.NorthBoundLatitude = bbox.Max[1]
				layer.EXGeographicBoundingBox = gbbox

				bbox1 := &wms130.LayerBoundingBox{}
				bbox1.CRS = "CRS:84"
				bbox1.Minx = bbox.Min[0]
				bbox1.Miny = bbox.Min[1]
				bbox1.Maxx = bbox.Max[0]
				bbox1.Maxy = bbox.Max[1]
				layer.BoundingBox = append(layer.BoundingBox, bbox1)
			}

			srsmap := c.layerSrsBBox(l, true)
			for k, v := range srsmap {
				if len(v.Min) >= 2 && len(v.Max) >= 2 {
					bbox1 := &wms130.LayerBoundingBox{}
					bbox1.CRS = k
					bbox1.Minx = v.Min[0]
					bbox1.Miny = v.Min[1]
					bbox1.Maxx = v.Max[0]
					bbox1.Maxy = v.Max[1]
					layer.BoundingBox = append(layer.BoundingBox, bbox1)
				}
			}

			if metadata != nil {
				if metadata.AuthorityURL != nil {
					layer.AuthorityURL = metadata.AuthorityURL
				}

				if metadata.Identifier != nil {
					layer.Identifier = metadata.Identifier
				}

				if len(metadata.MetadataURL) > 0 {
					layer.MetadataURL = append(layer.MetadataURL, metadata.MetadataURL...)
				}

				if len(metadata.Style) > 0 {
					layer.Style = append(layer.Style, metadata.Style...)
				}
			}

			cam.Capabilities.Layer = append(cam.Capabilities.Layer, layer)
		}
	}

	si, _ := xml.MarshalIndent(cam, "", "")

	return si
}
