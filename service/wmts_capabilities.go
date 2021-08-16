package service

import "github.com/flywave/go-tileproxy/request"

type WMTSCapabilities struct {
	Service     map[string]string
	Layers      []WMTSTileLayer
	MatrixSets  map[string]*TileMatrixSet
	InfoFormats map[string]string
}

func (c *WMTSCapabilities) render(request *request.WMTS100CapabilitiesRequest) []byte {
	return nil
}

func newWMTSCapabilities(md map[string]string, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats map[string]string) *WMTSCapabilities {
	return nil
}

type WMTSRestfulCapabilities struct {
	WMTSCapabilities
}

func newWMTSRestfulCapabilities(md map[string]string, layers []WMTSTileLayer, matrixSets map[string]*TileMatrixSet, infoFormats map[string]string) *WMTSCapabilities {
	return nil
}

func (c *WMTSRestfulCapabilities) render(request *request.WMTS100CapabilitiesRequest) []byte {
	return nil
}
