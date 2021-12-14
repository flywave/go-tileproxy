package resource

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
)

const (
	RASTER_DEM = "raster-dem"
	RASTER     = "raster"
	VECTOR     = "vector"
)

type VectorLayer struct {
	Id          string            `json:"id"`
	Description string            `json:"description"`
	Maxzoom     uint32            `json:"maxzoom"`
	Minzoom     uint32            `json:"minzoom"`
	Fileds      map[string]string `json:"fields"`
	Source      string            `json:"source"`
	SourceName  string            `json:"source_name"`
}

func NewVectorLayer() *VectorLayer {
	return &VectorLayer{Fileds: make(map[string]string)}
}

type TileJSON struct {
	Resource        `json:"-"`
	Attribution     string         `json:"attribution"`
	Description     string         `json:"description"`
	Bounds          [4]float32     `json:"bounds"`
	Center          [3]float32     `json:"center"`
	Created         uint64         `json:"created"`
	FileSize        uint64         `json:"filesize"`
	FillZoom        uint32         `json:"fillzoom"`
	Format          string         `json:"format"`
	ID              string         `json:"id"`
	MaxZoom         uint32         `json:"maxzoom"`
	MinZoom         uint32         `json:"minzoom"`
	Modified        uint64         `json:"modified"`
	Name            string         `json:"name"`
	Scheme          string         `json:"scheme"`
	TilejsonVersion string         `json:"tilejson"`
	Version         string         `json:"version"`
	VectorLayers    []*VectorLayer `json:"vector_layers,omitempty"`
	Tiles           []string       `json:"tiles,omitempty"`
	Data            []string       `json:"data,omitempty"`
	Template        *string        `json:"template,omitempty"`
	Legend          *string        `json:"legend,omitempty"`
	Grids           []string       `json:"grids,omitempty"`
	Webpage         string         `json:"webpage,omitempty"`
	Location        string         `json:"-"`
	Stored          bool           `json:"-"`
	StoreID         string         `json:"-"`
}

func (r *TileJSON) GetExtension() string {
	return "json"
}

func (r *TileJSON) IsStored() bool {
	return r.Stored
}

func (r *TileJSON) SetStored() {
	r.Stored = true
}

func (r *TileJSON) GetLocation() string {
	return r.Location
}

func (r *TileJSON) SetLocation(l string) {
	r.Location = l
}

func (r *TileJSON) GetID() string {
	return r.StoreID
}

func (r *TileJSON) SetID(id string) {
	r.StoreID = id
}

func (r *TileJSON) Hash() []byte {
	m := md5.New()
	m.Write([]byte(r.StoreID))
	return m.Sum(nil)
}

func (r *TileJSON) GetData() []byte {
	return r.ToJson()
}

func (r *TileJSON) SetData(content []byte) {
	reader := bytes.NewBuffer(content)
	dec := json.NewDecoder(reader)
	if err := dec.Decode(r); err != nil {
		return
	}
}

func NewTileJSON(id, name string) *TileJSON {
	att := &TileJSON{
		StoreID: id,
		Name:    name,
		Bounds:  [4]float32{-180, -85, 180, 85},
		Center:  [3]float32{0, 0, 0},
		Scheme:  "xyz",
	}
	return att
}

func CreateTileJSON(content []byte) *TileJSON {
	tillejson := &TileJSON{}
	reader := bytes.NewBuffer(content)
	dec := json.NewDecoder(reader)
	if err := dec.Decode(tillejson); err != nil {
		return nil
	}
	return tillejson
}

func (att *TileJSON) ToJson() []byte {
	var bt []byte
	wr := bytes.NewBuffer(bt)
	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	enc.Encode(att)
	return wr.Bytes()
}

type TileJSONCache struct {
	store Store
}

func (c *TileJSONCache) Save(r Resource) error {
	return c.store.Save(r)
}

func (c *TileJSONCache) Load(r Resource) error {
	return c.store.Load(r)
}

func NewTileJSONCache(store Store) *TileJSONCache {
	return &TileJSONCache{store: store}
}
