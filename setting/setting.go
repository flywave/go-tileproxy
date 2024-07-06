package setting

import (
	"bytes"
	"encoding/json"
)

type ProxyService struct {
	Id        string                 `json:"id,omitempty"`
	Service   interface{}            `json:"service,omitempty"`
	Coverages map[string]Coverage    `json:"coverages,omitempty"`
	Grids     map[string]GridOpts    `json:"grids,omitempty"`
	Sources   map[string]interface{} `json:"sources,omitempty"`
	Caches    map[string]interface{} `json:"caches,omitempty"`
}

func NewProxyService(id string) *ProxyService {
	return &ProxyService{
		Id:        id,
		Coverages: make(map[string]Coverage),
		Sources:   make(map[string]interface{}),
		Caches:    make(map[string]interface{}),
		Grids:     make(map[string]GridOpts),
	}
}

func CreateProxyServiceFromJSON(content []byte) *ProxyService {
	set := &ProxyService{}
	reader := bytes.NewBuffer(content)
	dec := json.NewDecoder(reader)
	if err := dec.Decode(set); err != nil {
		return nil
	}
	set.Service = covertService(set.Service)
	return set
}

func covertService(ser interface{}) interface{} {
	if s, ok := ser.(map[string]interface{}); ok {
		ty := s["type"].(string)
		data, _ := json.Marshal(ser)
		switch ty {
		case string(CESIUM_SERVICE):
			sv := &CesiumService{}
			err := json.Unmarshal(data, sv)
			if err != nil {
				return ser
			}
			return sv
		case string(MAPBOX_SERVICE):
			sv := &MapboxService{}
			err := json.Unmarshal(data, sv)
			if err != nil {
				return ser
			}
			return sv
		case string(WMS_SERVICE):
			sv := &WMSService{}
			err := json.Unmarshal(data, sv)
			if err != nil {
				return ser
			}
			return sv
		case string(WMTS_SERVICE):
			sv := &WMTSService{}
			err := json.Unmarshal(data, sv)
			if err != nil {
				return ser
			}
			return sv
		case string(TMS_SERVICE):
			sv := &TMSService{}
			err := json.Unmarshal(data, sv)
			if err != nil {
				return ser
			}
			return sv
		}
	}
	return ser
}

func (gs *ProxyService) ToJSON() []byte {
	var bt []byte
	wr := bytes.NewBuffer(bt)
	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	enc.Encode(gs)
	return wr.Bytes()
}
