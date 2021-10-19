package setting

import (
	"bytes"
	"encoding/json"
)

type ProxyService struct {
	UUID      string                 `json:"uuid,omitempty"`
	Service   interface{}            `json:"service,omitempty"`
	Coverages map[string]Coverage    `json:"coverages,omitempty"`
	Grids     map[string]GridOpts    `json:"grids,omitempty"`
	Sources   map[string]interface{} `json:"sources,omitempty"`
	Caches    map[string]interface{} `json:"caches,omitempty"`
}

func NewProxyService(uuid string) *ProxyService {
	return &ProxyService{
		UUID:      uuid,
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
	return set
}

func (gs *ProxyService) ToJSON() []byte {
	var bt []byte
	wr := bytes.NewBuffer(bt)
	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	enc.Encode(gs)
	return wr.Bytes()
}
