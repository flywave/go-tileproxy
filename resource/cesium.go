package resource

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
)

type AvailableBounds struct {
	StartX int `json:"startX"`
	StartY int `json:"startY"`
	EndX   int `json:"endX"`
	EndY   int `json:"endY"`
}

type LayerJson struct {
	Resource             `json:"-"`
	Name                 string              `json:"name"`
	Version              string              `json:"version"`
	Format               string              `json:"format"`
	Description          string              `json:"description,omitempty"`
	Attribution          string              `json:"attribution,omitempty"`
	Available            [][]AvailableBounds `json:"available"`
	MetadataAvailability int                 `json:"metadataAvailability,omitempty"`
	Bounds               [4]float64          `json:"bounds"`
	Extensions           []string            `json:"extensions,omitempty"`
	Minzoom              int                 `json:"minzoom"`
	Maxzoom              int                 `json:"maxzoom"`
	BVHLevels            int                 `json:"bvhlevels"`
	Projection           string              `json:"projection"`
	Scheme               string              `json:"scheme"`
	Tiles                []string            `json:"tiles"` // [ '{z}/{x}/{y}.terrain?v={version}' ]
	Location             string              `json:"-"`
	Stored               bool                `json:"-"`
	StoreID              string              `json:"-"`
}

func (r *LayerJson) GetExtension() string {
	return "json"
}

func (r *LayerJson) IsStored() bool {
	return r.Stored
}

func (r *LayerJson) SetStored() {
	r.Stored = true
}

func (r *LayerJson) GetLocation() string {
	return r.Location
}

func (r *LayerJson) SetLocation(l string) {
	r.Location = l
}

func (r *LayerJson) GetID() string {
	return r.StoreID
}

func (r *LayerJson) SetID(id string) {
	r.StoreID = id
}

func (r *LayerJson) Hash() []byte {
	m := md5.New()
	m.Write([]byte(r.StoreID))
	return m.Sum(nil)
}

func (r *LayerJson) GetData() []byte {
	return r.ToJson()
}

func (r *LayerJson) SetData(content []byte) {
	reader := bytes.NewBuffer(content)
	dec := json.NewDecoder(reader)
	if err := dec.Decode(r); err != nil {
		return
	}
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

type LayerJSONCache struct {
	store Store
}

func (c *LayerJSONCache) Save(r Resource) error {
	return c.store.Save(r)
}

func (c *LayerJSONCache) Load(r Resource) error {
	return c.store.Load(r)
}

func NewLayerJSONCache(store Store) *LayerJSONCache {
	return &LayerJSONCache{store: store}
}
