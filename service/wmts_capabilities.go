package service

import (
	"bytes"
	"encoding/xml"
	"html/template"
	"strings"

	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/ogc-specifications/pkg/wmts100"
)

type WMTSCapabilities struct {
	Service     *WMTSMetadata
	Layers      []WMTSTileLayer
	MatrixSets  map[string]*TileMatrixSet
	InfoFormats map[string]string
}

func formatResourceTemplate(layer WMTSTileLayer, tpl string, service *WMTSMetadata) string {
	p := map[string]string{"Format": layer.GetFormat(), "Layer": layer.GetName()}

	if strings.Contains(tpl, "{InfoFormat}") {
		tpl = strings.Replace(tpl, "{InfoFormat}", "{{ .InfoFormat }}", 1)
	}

	if strings.Contains(tpl, "{Layer}") {
		tpl = strings.Replace(tpl, "{Layer}", "{{ .Layer }}", 1)
	}

	tmpl, _ := template.New("url").Parse(tpl)

	out := &bytes.Buffer{}

	_ = tmpl.Execute(out, p)

	return service.URL + string(out.Bytes())
}

func (c *WMTSCapabilities) render(request *request.WMTS100CapabilitiesRequest) []byte {
	resp := wmts100.GetCapabilitiesResponse{}
	resp.Namespaces.Xmlns = "http://www.opengis.net/wmts/1.0"
	resp.Namespaces.XmlnsOws = "http://www.opengis.net/ows/1.1"
	resp.Namespaces.XmlnsXlink = "http://www.w3.org/1999/xlink"
	resp.Namespaces.XmlnsXSI = "http://www.w3.org/2001/XMLSchema-instance"
	resp.Namespaces.XmlnsGml = "http://www.opengis.net/gml"
	resp.Namespaces.SchemaLocation = "http://www.opengis.net/wmts/1.0 http://schemas.opengis.net/wmts/1.0/wmtsGetCapabilities_response.xsd"
	resp.Namespaces.Version = "1.0.0"

	identification := &resp.ServiceIdentification
	identification.Title = c.Service.Title
	identification.Abstract = c.Service.Abstract
	identification.ServiceType = "OGC WMTS"
	identification.ServiceTypeVersion = "1.0.0"
	if c.Service.Fees != nil {
		identification.Fees = *c.Service.Fees
	} else {
		identification.Fees = "none"
	}
	if c.Service.AccessConstraints != nil {
		identification.AccessConstraints = *c.Service.AccessConstraints
	} else {
		identification.AccessConstraints = "none"
	}

	if c.Service.Provider != nil {
		resp.ServiceProvider = *c.Service.Provider
	}

	url := c.Service.URL

	if resp.OperationsMetadata == nil {
		resp.OperationsMetadata = &wmts100.OperationsMetadata{}
	}

	op := wmts100.Operation{}
	op.Name = "GetCapabilities"
	get := &wmts100.Get{}
	get.Href = url
	get.Constraint.Name = "GetEncoding"
	get.Constraint.AllowedValues.Value = append(get.Constraint.AllowedValues.Value, "KVP")
	op.DCP.HTTP.Get = get
	resp.OperationsMetadata.Operation = append(resp.OperationsMetadata.Operation, op)

	op = wmts100.Operation{}
	op.Name = "GetTile"
	get = &wmts100.Get{}
	get.Href = url
	get.Constraint.Name = "GetEncoding"
	get.Constraint.AllowedValues.Value = append(get.Constraint.AllowedValues.Value, "KVP")
	op.DCP.HTTP.Get = get
	resp.OperationsMetadata.Operation = append(resp.OperationsMetadata.Operation, op)

	op = wmts100.Operation{}
	op.Name = "GetFeatureInfo"
	get = &wmts100.Get{}
	get.Href = url
	get.Constraint.Name = "GetEncoding"
	get.Constraint.AllowedValues.Value = append(get.Constraint.AllowedValues.Value, "KVP")
	op.DCP.HTTP.Get = get
	resp.OperationsMetadata.Operation = append(resp.OperationsMetadata.Operation, op)

	contents := &resp.Contents
	for _, l := range c.Layers {
		layer := wmts100.Layer{}
		layer.Title = l.GetTitle()

		bbox := l.LLBBox()

		layer.WGS84BoundingBox.LowerCorner = [2]float64{bbox.Min[0], bbox.Min[1]}
		layer.WGS84BoundingBox.UpperCorner = [2]float64{bbox.Max[0], bbox.Max[1]}

		layer.Identifier = l.GetName()
		layer.Style = append(layer.Style, wmts100.Style{Identifier: "default"})
		layer.Format = []string{l.GetFormatMimeType()}

		for _, g := range l.GetGrids() {
			layer.TileMatrixSetLink = append(layer.TileMatrixSetLink, wmts100.TileMatrixSetLink{TileMatrixSet: g.Name})
		}

		contents.Layer = append(contents.Layer, layer)
	}

	for _, tm := range c.MatrixSets {
		tms := wmts100.TileMatrixSet{}
		tms.Identifier = tm.name
		tms.SupportedCRS = tm.srs_name
		for _, t := range tm.GetTileMatrices() {
			tt := wmts100.TileMatrix{}
			tt.Identifier = t["identifier"]
			tt.ScaleDenominator = t["scale_denom"]
			tt.TopLeftCorner = t["topleft"]
			tt.TileWidth = t["tile_width"]
			tt.TileHeight = t["tile_height"]
			tt.MatrixWidth = t["matrix_width"]
			tt.MatrixHeight = t["matrix_width"]

			tms.TileMatrix = append(tms.TileMatrix, tt)
		}
		contents.TileMatrixSet = append(contents.TileMatrixSet, tms)
	}

	if resp.ServiceMetadataURL == nil {
		resp.ServiceMetadataURL = &wmts100.ServiceMetadataURL{}
	}

	resp.ServiceMetadataURL.Href = url + "/1.0.0/WMTSCapabilities.xml"

	si, _ := xml.MarshalIndent(resp, "", "")

	return si
}

func newWMTSCapabilities(md *WMTSMetadata, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats map[string]string) *WMTSCapabilities {
	return &WMTSCapabilities{Service: md, Layers: layers, MatrixSets: matrixSets, InfoFormats: infoFormats}
}

type RestfulCapabilities struct {
	WMTSCapabilities
	urlConverter     *request.URLTemplateConverter
	infoUrlConverter *request.URLTemplateConverter
	resourceTemplate string
}

func (c *RestfulCapabilities) render(request *request.WMTS100CapabilitiesRequest) []byte {
	resp := wmts100.GetCapabilitiesResponse{}
	resp.Namespaces.Xmlns = "http://www.opengis.net/wmts/1.0"
	resp.Namespaces.XmlnsOws = "http://www.opengis.net/ows/1.1"
	resp.Namespaces.XmlnsXlink = "http://www.w3.org/1999/xlink"
	resp.Namespaces.XmlnsXSI = "http://www.w3.org/2001/XMLSchema-instance"
	resp.Namespaces.XmlnsGml = "http://www.opengis.net/gml"
	resp.Namespaces.SchemaLocation = "http://www.opengis.net/wmts/1.0 http://schemas.opengis.net/wmts/1.0/wmtsGetCapabilities_response.xsd"
	resp.Namespaces.Version = "1.0.0"

	identification := &resp.ServiceIdentification
	identification.Title = c.Service.Title
	identification.Abstract = c.Service.Abstract
	identification.ServiceType = "OGC WMTS"
	identification.ServiceTypeVersion = "1.0.0"
	if c.Service.Fees != nil {
		identification.Fees = *c.Service.Fees
	} else {
		identification.Fees = "none"
	}
	if c.Service.AccessConstraints != nil {
		identification.AccessConstraints = *c.Service.AccessConstraints
	} else {
		identification.AccessConstraints = "none"
	}

	if c.Service.Provider != nil {
		resp.ServiceProvider = *c.Service.Provider
	}

	url := c.Service.URL

	contents := &resp.Contents
	for _, l := range c.Layers {
		layer := wmts100.Layer{}
		layer.Title = l.GetTitle()

		bbox := l.LLBBox()

		layer.WGS84BoundingBox.LowerCorner = [2]float64{bbox.Min[0], bbox.Min[1]}
		layer.WGS84BoundingBox.UpperCorner = [2]float64{bbox.Max[0], bbox.Max[1]}

		layer.Identifier = l.GetName()
		layer.Style = append(layer.Style, wmts100.Style{Identifier: "default"})
		layer.Format = []string{l.GetFormatMimeType()}

		for i := range c.InfoFormats {
			layer.InfoFormat = append(layer.InfoFormat, c.InfoFormats[i])
		}

		for _, g := range l.GetGrids() {
			layer.TileMatrixSetLink = append(layer.TileMatrixSetLink, wmts100.TileMatrixSetLink{TileMatrixSet: g.Name})
		}

		layer.ResourceURL.Format = l.GetFormatMimeType()
		layer.ResourceURL.ResourceType = "tile"
		layer.ResourceURL.Template = formatResourceTemplate(l, c.resourceTemplate, c.Service)

		contents.Layer = append(contents.Layer, layer)
	}

	for _, tm := range c.MatrixSets {
		tms := wmts100.TileMatrixSet{}
		tms.Identifier = tm.name
		tms.SupportedCRS = tm.srs_name
		for _, t := range tm.GetTileMatrices() {
			tt := wmts100.TileMatrix{}
			tt.Identifier = t["identifier"]
			tt.ScaleDenominator = t["scale_denom"]
			tt.TopLeftCorner = t["topleft"]
			tt.TileWidth = t["tile_width"]
			tt.TileHeight = t["tile_height"]
			tt.MatrixWidth = t["matrix_width"]
			tt.MatrixHeight = t["matrix_width"]

			tms.TileMatrix = append(tms.TileMatrix, tt)
		}
		contents.TileMatrixSet = append(contents.TileMatrixSet, tms)
	}

	resp.ServiceMetadataURL.Href = url + "/1.0.0/WMTSCapabilities.xml"

	si, _ := xml.MarshalIndent(resp, "", "")

	return si
}

func newWMTSRestCapabilities(md *WMTSMetadata, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, urlConverter *request.URLTemplateConverter, infoUrlConverter *request.URLTemplateConverter, infoFormats map[string]string, resourceTemplate string) *RestfulCapabilities {
	return &RestfulCapabilities{WMTSCapabilities: WMTSCapabilities{Service: md, Layers: layers, MatrixSets: matrixSets, InfoFormats: infoFormats}, resourceTemplate: resourceTemplate, urlConverter: urlConverter, infoUrlConverter: infoUrlConverter}
}
