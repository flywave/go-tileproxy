package service

import (
	"bytes"
	"encoding/xml"
	"html/template"

	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/ogc-specifications/pkg/wmts100"
)

type WMTSCapabilities struct {
	Service          map[string]string
	Layers           []WMTSTileLayer
	MatrixSets       map[string]*TileMatrixSet
	InfoFormats      map[string]string
	Restful          bool
	resourceTemplate string
}

func formatResourceTemplate(layer WMTSTileLayer, tpl string, service map[string]string) string {
	p := map[string]string{"Format": layer.GetFormat(), "Layer": layer.GetName()}

	tmpl, _ := template.New("url").Parse(tpl)

	out := &bytes.Buffer{}

	_ = tmpl.Execute(out, p)

	return service["url"] + string(out.Bytes())
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
	identification.Title = c.Service["title"]
	identification.Abstract = c.Service["abstract"]
	identification.ServiceType = "OGC WMTS"
	identification.ServiceTypeVersion = "1.0.0"
	if f, ok := c.Service["fees"]; ok {
		identification.Fees = f
	} else {
		identification.Fees = "none"
	}
	if f, ok := c.Service["access_constraints"]; ok {
		identification.AccessConstraints = f
	} else {
		identification.AccessConstraints = "none"
	}

	sp := serviceProviderFromMetadata(c.Service)

	if sp != nil {
		serviceProvider := &resp.ServiceProvider
		serviceProvider.ProviderName = sp.ProviderName
		serviceProvider.ProviderSite.Type = sp.ProviderSite.Type
		serviceProvider.ProviderSite.Href = sp.ProviderSite.Href

		serviceProvider.ServiceContact.IndividualName = sp.ServiceContact.IndividualName
		serviceProvider.ServiceContact.PositionName = sp.ServiceContact.PositionName
		serviceProvider.ServiceContact.ContactInfo.Text = sp.ServiceContact.ContactInfo.Text
		serviceProvider.ServiceContact.ContactInfo.Text = sp.ServiceContact.ContactInfo.Text
		serviceProvider.ServiceContact.ContactInfo.Phone.Voice = sp.ServiceContact.ContactInfo.Phone.Voice
		serviceProvider.ServiceContact.ContactInfo.Phone.Facsimile = sp.ServiceContact.ContactInfo.Phone.Facsimile

		serviceProvider.ServiceContact.ContactInfo.Address.DeliveryPoint = sp.ServiceContact.ContactInfo.Address.DeliveryPoint
		serviceProvider.ServiceContact.ContactInfo.Address.City = sp.ServiceContact.ContactInfo.Address.City
		serviceProvider.ServiceContact.ContactInfo.Address.AdministrativeArea = sp.ServiceContact.ContactInfo.Address.AdministrativeArea
		serviceProvider.ServiceContact.ContactInfo.Address.PostalCode = sp.ServiceContact.ContactInfo.Address.PostalCode
		serviceProvider.ServiceContact.ContactInfo.Address.Country = sp.ServiceContact.ContactInfo.Address.Country
		serviceProvider.ServiceContact.ContactInfo.Address.ElectronicMailAddress = sp.ServiceContact.ContactInfo.Address.ElectronicMailAddress

		serviceProvider.ServiceContact.ContactInfo.OnlineResource.Type = sp.ServiceContact.ContactInfo.OnlineResource.Type
		serviceProvider.ServiceContact.ContactInfo.OnlineResource.Href = sp.ServiceContact.ContactInfo.OnlineResource.Href

		serviceProvider.ServiceContact.ContactInfo.HoursOfService = sp.ServiceContact.ContactInfo.HoursOfService
		serviceProvider.ServiceContact.ContactInfo.ContactInstructions = sp.ServiceContact.ContactInfo.ContactInstructions

		serviceProvider.ServiceContact.Role = sp.ServiceContact.Role
	}

	url := c.Service["url"]

	if !c.Restful {
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
	}

	contents := &resp.Contents
	for _, l := range c.Layers {
		layer := wmts100.Layer{}
		layer.Title = l.GetTitle()

		bbox := l.LLBBox()

		layer.WGS84BoundingBox.LowerCorner = [2]float64{bbox.Min[0], bbox.Min[1]}
		layer.WGS84BoundingBox.UpperCorner = [2]float64{bbox.Max[0], bbox.Max[1]}

		layer.Identifier = l.GetName()
		layer.Style = append(layer.Style, wmts100.Style{Identifier: "default"})
		layer.Format = l.GetFormatMimeType()

		for _, g := range l.GetGrids() {
			layer.TileMatrixSetLink = append(layer.TileMatrixSetLink, wmts100.TileMatrixSetLink{TileMatrixSet: g.Name})
		}

		if c.Restful {
			layer.ResourceURL.Format = l.GetFormatMimeType()
			layer.ResourceURL.ResourceType = "tile"
			layer.ResourceURL.Template = formatResourceTemplate(l, c.resourceTemplate, c.Service)
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

	resp.ServiceMetadataURL.Href = url + "/1.0.0/WMTSCapabilities.xml"

	si, _ := xml.MarshalIndent(resp, "", "")

	return si
}

func newWMTSCapabilities(md map[string]string, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats map[string]string) *WMTSCapabilities {
	return &WMTSCapabilities{Service: md, Layers: layers, MatrixSets: matrixSets, InfoFormats: infoFormats, Restful: false}
}

func newWMTSRestCapabilities(md map[string]string, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats map[string]string, resourceTemplate string) *WMTSCapabilities {
	return &WMTSCapabilities{Service: md, Layers: layers, MatrixSets: matrixSets, InfoFormats: infoFormats, Restful: true, resourceTemplate: resourceTemplate}
}

type ServiceProvider struct {
	ProviderName string
	ProviderSite struct {
		Type string
		Href string
	}
	ServiceContact struct {
		IndividualName string
		PositionName   string
		ContactInfo    struct {
			Text  string
			Phone struct {
				Voice     string
				Facsimile string
			}
			Address struct {
				DeliveryPoint         string
				City                  string
				AdministrativeArea    string
				PostalCode            string
				Country               string
				ElectronicMailAddress string
			}
			OnlineResource struct {
				Type string
				Href string
			}
			HoursOfService      string
			ContactInstructions string
		}
		Role string
	}
}

func serviceProviderFromMetadata(metadata map[string]string) *ServiceProvider {
	ret := &ServiceProvider{}
	if l, ok := metadata["serviceprovider.providername"]; ok {
		ret.ProviderName = l
	}
	if l, ok := metadata["serviceprovider.providersite.type"]; ok {
		ret.ProviderSite.Type = l
	}
	if l, ok := metadata["serviceprovider.providersite.href"]; ok {
		ret.ProviderSite.Href = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.individualname"]; ok {
		ret.ServiceContact.IndividualName = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.positionname"]; ok {
		ret.ServiceContact.PositionName = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.text"]; ok {
		ret.ServiceContact.ContactInfo.Text = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.phone.voice"]; ok {
		ret.ServiceContact.ContactInfo.Phone.Voice = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.phone.facsimile"]; ok {
		ret.ServiceContact.ContactInfo.Phone.Facsimile = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.address.deliverypoint"]; ok {
		ret.ServiceContact.ContactInfo.Address.DeliveryPoint = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.address.city"]; ok {
		ret.ServiceContact.ContactInfo.Address.City = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.address.administrativearea"]; ok {
		ret.ServiceContact.ContactInfo.Address.AdministrativeArea = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.address.postalcode"]; ok {
		ret.ServiceContact.ContactInfo.Address.PostalCode = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.address.country"]; ok {
		ret.ServiceContact.ContactInfo.Address.Country = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.address.electronicmailaddress"]; ok {
		ret.ServiceContact.ContactInfo.Address.ElectronicMailAddress = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.onlineresource.type"]; ok {
		ret.ServiceContact.ContactInfo.OnlineResource.Type = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.onlineresource.href"]; ok {
		ret.ServiceContact.ContactInfo.OnlineResource.Href = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.hoursofservice"]; ok {
		ret.ServiceContact.ContactInfo.HoursOfService = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.contactinfo.contactinstructions"]; ok {
		ret.ServiceContact.ContactInfo.ContactInstructions = l
	}
	if l, ok := metadata["serviceprovider.servicecontact.role"]; ok {
		ret.ServiceContact.Role = l
	}
	return ret
}
