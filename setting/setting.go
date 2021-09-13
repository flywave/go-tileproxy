package setting

import (
	"bytes"
	"encoding/json"
)

type ProxyDataset struct {
	Identifier string                 `json:"id,omitempty"`
	Service    interface{}            `json:"service,omitempty"`
	Coverages  map[string]Coverage    `json:"coverages,omitempty"`
	Grids      map[string]GridOpts    `json:"grids,omitempty"`
	Sources    map[string]interface{} `json:"sources,omitempty"`
	Caches     map[string]interface{} `json:"caches,omitempty"`
}

func NewProxyDataset(name string) *ProxyDataset {
	return &ProxyDataset{
		Identifier: name,
		Coverages:  make(map[string]Coverage),
		Sources:    make(map[string]interface{}),
		Caches:     make(map[string]interface{}),
		Grids:      make(map[string]GridOpts),
	}
}

func CreateProxyDatasetFromJSON(content []byte) *ProxyDataset {
	set := &ProxyDataset{}
	reader := bytes.NewBuffer(content)
	dec := json.NewDecoder(reader)
	if err := dec.Decode(set); err != nil {
		return nil
	}
	return set
}

func (gs *ProxyDataset) ToJSON() []byte {
	var bt []byte
	wr := bytes.NewBuffer(bt)
	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	enc.Encode(gs)
	return wr.Bytes()
}
