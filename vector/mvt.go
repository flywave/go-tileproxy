package vector

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/flywave/go-geom"
	"github.com/flywave/go-mapbox/mvt"
	"github.com/flywave/go-mapbox/tileid"
	"github.com/flywave/go-tileproxy/tile"
)

const (
	PBF_PTOTO_MAPBOX   mvt.ProtoType = mvt.PROTO_MAPBOX
	PBF_PTOTO_LUOKUANG mvt.ProtoType = mvt.PROTO_LK
	PBF_MIME_MAPBOX                  = "application/vnd.mapbox-vector-tile"
	PBF_MIME_LUOKUANG                = "application/x-protobuf"
)

type MVTSource struct {
	VectorSource
	Proto mvt.ProtoType
}

func NewMVTSource(tile [3]int, proto mvt.ProtoType, options tile.TileOptions) *MVTSource {
	src := &MVTSource{Proto: proto, VectorSource: VectorSource{tile: tile, Options: options}}
	src.io = &PBFIO{tile: tile, proto: proto}
	return src
}

func NewEmptyMVTSource(proto mvt.ProtoType, options tile.TileOptions) *MVTSource {
	src := &MVTSource{Proto: proto, VectorSource: VectorSource{tile: [3]int{0, 0, 0}, Options: options}}
	src.io = &PBFIO{tile: [3]int{0, 0, 0}, proto: proto}
	src.SetSource(Vector{})
	return src
}

type PBFIO struct {
	VectorIO
	tile  [3]int
	proto mvt.ProtoType
}

func (i *PBFIO) Decode(r io.Reader) (interface{}, error) {
	PBF := LoadPBF(r, i.tile, i.proto)
	return PBF, nil
}

func (i *PBFIO) Encode(data interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := SavePBF(buf, i.tile, i.proto, data.(Vector))
	return buf.Bytes(), err
}

func ConvertPBFToGeom(p Vector) *geom.FeatureCollection {
	fc := &geom.FeatureCollection{}
	for l, fs := range p {
		for i := range fs {
			if _, ok := fs[i].Properties["layer"]; !ok {
				fs[i].Properties["layer"] = l
			}
			fc.Features = append(fc.Features, fs[i])
		}
	}
	return fc
}

func LoadPBF(r io.Reader, coord [3]int, proto mvt.ProtoType) Vector {
	pbf := make(Vector)
	pbfdata, _ := ioutil.ReadAll(r)
	tileid := tileid.TileID{X: int64(coord[0]), Y: int64(coord[1]), Z: uint64(coord[2])}
	tile, _ := mvt.NewTile(pbfdata, mvt.ProtoType(proto))

	for _, layer := range tile.LayerMap {
		feats := []*geom.Feature{}
		for layer.Next() {
			feat, _ := layer.Feature()
			featg, _ := feat.ToGeoJSON(tileid)
			feats = append(feats, featg)
		}
		pbf[layer.Name] = feats
	}

	return pbf
}

func SavePBF(w io.Writer, coord [3]int, proto mvt.ProtoType, vts Vector) error {
	data := []byte{}
	tileid := tileid.TileID{X: int64(coord[0]), Y: int64(coord[1]), Z: uint64(coord[2])}

	for layer, feats := range vts {
		conf := mvt.NewConfig(layer, tileid, mvt.ProtoType(proto))
		conf.ExtentBool = false
		data = append(data, mvt.WriteLayer(feats, conf)...)
	}

	_, err := w.Write(data)
	return err
}
