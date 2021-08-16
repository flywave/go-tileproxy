package service

import (
	"math"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/ogc-specifications/pkg/wms130"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/request"
)

type WMSCapabilities struct {
	service         map[string]string
	root_layer      *WMSGroupLayer
	tile_layers     []*TileProvider
	imageFormats    []string
	infoFormats     []string
	srs             *geo.SupportedSRS
	srsExtents      map[string]*geo.MapExtent
	inspireMetadata *ExtendedCapabilities
	maxOutputPixels int
	contact         *ContactInformation
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
		return limitLLBBox(llbbox)
	}
	return limitLLBBox(layer.extent.GetLLBBox())
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

func (c *WMSCapabilities) render(req *request.WMSRequest) []byte {
	cam := &wms130.GetCapabilitiesResponse{}
	service := &cam.WMSService
	service.Name = "WMS"
	service.Title = c.service["title"]
	service.Abstract = c.service["abstract"]
	if l, ok := c.service["keyword_list"]; ok {
		service.KeywordList = &wms130.Keywords{}
		keys := strings.Split(l, ",")
		for i := range keys {
			service.KeywordList.Keyword = append(service.KeywordList.Keyword, keys[i])
		}
	}
	url := c.service["url"]

	if l, ok := c.service["online_resource"]; ok {
		service.OnlineResource.Href = &l
	} else {
		service.OnlineResource.Href = &url
	}
	if c.contact != nil {
		service.ContactInformation.ContactPersonPrimary.ContactPerson = c.contact.ContactPersonPrimary.ContactPerson
		service.ContactInformation.ContactPersonPrimary.ContactOrganization = c.contact.ContactPersonPrimary.ContactOrganization
		service.ContactInformation.ContactPosition = c.contact.ContactPosition
		service.ContactInformation.ContactAddress.AddressType = c.contact.ContactAddress.AddressType
		service.ContactInformation.ContactAddress.Address = c.contact.ContactAddress.Address
		service.ContactInformation.ContactAddress.City = c.contact.ContactAddress.City
		service.ContactInformation.ContactAddress.StateOrProvince = c.contact.ContactAddress.StateOrProvince
		service.ContactInformation.ContactAddress.PostCode = c.contact.ContactAddress.PostCode
		service.ContactInformation.ContactAddress.Country = c.contact.ContactAddress.Country
		service.ContactInformation.ContactVoiceTelephone = c.contact.ContactVoiceTelephone
		service.ContactInformation.ContactFacsimileTelephone = c.contact.ContactFacsimileTelephone
		service.ContactInformation.ContactElectronicMailAddress = c.contact.ContactElectronicMailAddress
	}
	if f, ok := c.service["fees"]; ok {
		service.Fees = f
	} else {
		service.Fees = "none"
	}
	if f, ok := c.service["access_constraints"]; ok {
		service.AccessConstraints = f
	} else {
		service.AccessConstraints = "none"
	}

	var max_output_size []int
	if c.maxOutputPixels > 0 {
		output_size := int(math.Sqrt(float64(c.maxOutputPixels)))
		max_output_size = []int{output_size, output_size}
	}

	if max_output_size != nil {
		service.OptionalConstraints.MaxWidth = max_output_size[0]
		service.OptionalConstraints.MaxHeight = max_output_size[1]
	}

	capabilities := &cam.Capabilities.WMSCapabilities

	capabilities.Request.GetCapabilities.Format = []string{"text/xml"}
	capabilities.Request.GetCapabilities.DCPType.HTTP.Get.OnlineResource.Xlink = &url

	if len(c.imageFormats) > 0 {
		capabilities.Request.GetMap.Format = c.imageFormats[:]
	}
	capabilities.Request.GetMap.DCPType.HTTP.Get.OnlineResource.Xlink = &url

	if len(c.infoFormats) > 0 {
		capabilities.Request.GetFeatureInfo.Format = c.infoFormats[:]
	}
	capabilities.Request.GetFeatureInfo.DCPType.HTTP.Get.OnlineResource.Xlink = &url

	capabilities.Exception.Format = []string{"XML", "INIMAGE", "BLANK"}

	if c.inspireMetadata != nil {
		ec := &wms130.ExtendedCapabilities{}
		ec.MetadataURL.Type = c.inspireMetadata.MetadataURL.Type
		ec.MetadataURL.URL = c.inspireMetadata.MetadataURL.URL
		ec.MetadataURL.MediaType = c.inspireMetadata.MetadataURL.MediaType

		ec.SupportedLanguages.DefaultLanguage.Language = c.inspireMetadata.SupportedLanguages.DefaultLanguage.Language

		ec.ResponseLanguage.Language = c.inspireMetadata.ResponseLanguage.Language

		capabilities.ExtendedCapabilities = ec
	}

	for _, l := range c.tile_layers {
		layer := &wms130.Layer{}
		layer.Name = &l.name
		layer.Title = l.title
		metadata := l.metadata
		if ab, ok := metadata["abstract"]; ok {
			layer.Abstract = ab
		}

		if l, ok := metadata["keyword_list"]; ok {
			layer.KeywordList = &wms130.Keywords{}
			keys := strings.Split(l, ",")
			for i := range keys {
				layer.KeywordList.Keyword = append(layer.KeywordList.Keyword, keys[i])
			}
		}

		for i := range c.srs.Srs {
			SrsCode := c.srs.Srs[i].GetSrsCode()
			epsg_num := geo.GetEpsgNum(SrsCode)
			crs := wms130.CRS{Code: epsg_num}
			layer.CRS = append(layer.CRS, crs)
		}

		bbox := c.layerLLBBox(l)
		gbbox := &wms130.EXGeographicBoundingBox{}
		gbbox.WestBoundLongitude = bbox.Min[0]
		gbbox.EastBoundLongitude = bbox.Min[1]
		gbbox.SouthBoundLatitude = bbox.Max[0]
		gbbox.NorthBoundLatitude = bbox.Max[1]

		layer.EXGeographicBoundingBox = gbbox

		bbox1 := &wms130.LayerBoundingBox{}
		bbox1.CRS = "CRS:84"
		bbox1.Minx = bbox.Min[0]
		bbox1.Miny = bbox.Min[1]
		bbox1.Maxx = bbox.Max[0]
		bbox1.Maxy = bbox.Max[1]

		layer.BoundingBox = append(layer.BoundingBox, bbox1)

		srsmap := c.layerSrsBBox(l, true)

		for k := range srsmap {
			bbox1 := &wms130.LayerBoundingBox{}
			bbox1.CRS = k
			bbox1.Minx = srsmap[k].Min[0]
			bbox1.Miny = srsmap[k].Min[1]
			bbox1.Maxx = srsmap[k].Max[0]
			bbox1.Maxy = srsmap[k].Max[1]

			layer.BoundingBox = append(layer.BoundingBox, bbox1)
		}

		tm := TileMetadataFromMetadata(metadata)

		if tm != nil {
			au := &wms130.AuthorityURL{}
			au.Name = tm.AuthorityURL.Name
			au.OnlineResource.Type = tm.AuthorityURL.OnlineResource.Type
			au.OnlineResource.Href = tm.AuthorityURL.OnlineResource.Href
			au.OnlineResource.Xlink = tm.AuthorityURL.OnlineResource.Xlink
			layer.AuthorityURL = au

			id := &wms130.Identifier{}
			id.Authority = tm.Identifier.Authority
			id.Value = tm.Identifier.Value
			layer.Identifier = id

			for i := range tm.MetadataURL {
				u := &wms130.MetadataURL{}
				u.Type = tm.MetadataURL[i].Type
				u.Format = tm.MetadataURL[i].Format
				u.OnlineResource.Xlink = tm.MetadataURL[i].OnlineResource.Xlink
				u.OnlineResource.Type = tm.MetadataURL[i].OnlineResource.Type
				u.OnlineResource.Href = tm.MetadataURL[i].OnlineResource.Href
				layer.MetadataURL = append(layer.MetadataURL, u)
			}
			for i := range tm.Style {
				s := &wms130.Style{}
				s.Name = tm.Style[i].Name
				s.Title = tm.Style[i].Title
				s.LegendURL.Width = tm.Style[i].LegendURL.Width
				s.LegendURL.Height = tm.Style[i].LegendURL.Height
				s.LegendURL.Format = tm.Style[i].LegendURL.Format
				s.LegendURL.OnlineResource.Xlink = tm.Style[i].LegendURL.OnlineResource.Xlink
				s.LegendURL.OnlineResource.Type = tm.Style[i].LegendURL.OnlineResource.Type
				s.LegendURL.OnlineResource.Href = tm.Style[i].LegendURL.OnlineResource.Href
				layer.Style = append(layer.Style, s)
			}
		}
	}

	return cam.ToXML()
}

type ContactInformation struct {
	ContactPersonPrimary struct {
		ContactPerson       string
		ContactOrganization string
	}
	ContactPosition string
	ContactAddress  struct {
		AddressType     string
		Address         string
		City            string
		StateOrProvince string
		PostCode        string
		Country         string
	}
	ContactVoiceTelephone        string
	ContactFacsimileTelephone    string
	ContactElectronicMailAddress string
}

type ExtendedCapabilities struct {
	MetadataURL struct {
		Type      string
		URL       string
		MediaType string
	}
	SupportedLanguages struct {
		DefaultLanguage struct {
			Language string
		}
	}
	ResponseLanguage struct {
		Language string
	}
}

type TileMetadata struct {
	AuthorityURL *AuthorityURL
	Identifier   *Identifier
	MetadataURL  []*MetadataURL
	Style        []*Style
}

func TileMetadataFromMetadata(data map[string]string) *TileMetadata {
	return nil
}

type OnlineResource struct {
	Xlink *string
	Type  *string
	Href  *string
}

type Identifier struct {
	Authority string
	Value     string
}

type MetadataURL struct {
	Type           *string
	Format         *string
	OnlineResource OnlineResource
}

type AuthorityURL struct {
	Name           string
	OnlineResource OnlineResource
}

type Style struct {
	Name      string
	Title     string
	LegendURL struct {
		Width          int
		Height         int
		Format         string
		OnlineResource OnlineResource
	}
}
