package resource

import (
	"bytes"
	"encoding/json"
)

type AvailableBounds struct {
	StartX int `json:"startX"`
	StartY int `json:"startY"`
	EndX   int `json:"endX"`
	EndY   int `json:"endY"`
}

type LayerJson struct {
	Name                 string              `json:"name"`
	Version              string              `json:"version"`
	Format               string              `json:"format"`
	Description          string              `json:"description,omitempty"`
	Attribution          string              `json:"attribution,omitempty"`
	Available            [][]AvailableBounds `json:"available"`
	MetadataAvailability int                 `json:"metadataAvailability"`
	Bounds               [4]float64          `json:"bounds"`
	Extensions           []string            `json:"extensions"`
	Minzoom              int                 `json:"minzoom"`
	Maxzoom              int                 `json:"maxzoom"`
	BVHLevels            int                 `json:"bvhlevels"`
	Projection           string              `json:"projection"`
	Scheme               string              `json:"scheme"`
	Tiles                []string            `json:"tiles"` // [ '{z}/{x}/{y}.terrain?v={version}' ]
}

func CreateLayerJson(content []byte) *LayerJson {
	tillejson := &LayerJson{}
	reader := bytes.NewBuffer(content)
	dec := json.NewDecoder(reader)
	if err := dec.Decode(tillejson); err != nil {
		return nil
	}
	return tillejson
}

func (att *LayerJson) ToJson() []byte {
	var bt []byte
	wr := bytes.NewBuffer(bt)
	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	enc.Encode(att)
	return wr.Bytes()
}
