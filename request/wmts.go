package request

type WMTSTileRequestParams struct{}

type WMTSRequest struct {
}

type WMTS100TileRequest struct{}

type WMTSFeatureInfoRequestParams struct {
	WMTSTileRequestParams
}

type WMTS100FeatureInfoRequest struct {
	WMTS100TileRequest
}

type WMTS100CapabilitiesRequest struct {
	WMTSRequest
}

type URLTemplateConverter struct {
}

type FeatureInfoURLTemplateConverter struct {
	URLTemplateConverter
}

type WMTS100RestTileRequest struct {
	TileRequest
}

type WMTS100RestFeatureInfoRequest struct {
	TileRequest
}

type WMTS100RestCapabilitiesRequest struct{}
