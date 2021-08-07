package vector

import (
	"io"
	"io/ioutil"

	"github.com/flywave/go-geom"
	"github.com/flywave/go-mapbox/mvt"
	"github.com/flywave/go-mapbox/tileid"
	"github.com/flywave/go-tileproxy/tile"
)

type PBFProto mvt.ProtoType

const (
	PBF_PTOTO_MAPBOX   mvt.ProtoType = mvt.PROTO_MAPBOX
	PBF_PTOTO_LUOKUANG mvt.ProtoType = mvt.PROTO_LK
)

type PBF map[string][]*geom.Feature

type MVTSource struct {
	VectorSource
	Proto PBFProto
}

func NewMVTSource(tile [3]int, proto PBFProto, options tile.TileOptions) *MVTSource {
	src := &MVTSource{Proto: proto, VectorSource: VectorSource{tile: tile, Options: options}}
	src.decodeFunc = func(r io.Reader) (interface{}, error) {
		PBF := LoadPBF(r, tile, src.Proto)
		return PBF, nil
	}
	src.encodeFunc = func(data interface{}) ([]byte, error) {
		return nil, nil
	}
	return src
}

func ConvertPBFToGeom(p PBF) *geom.FeatureCollection {
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

func LoadPBF(r io.Reader, coord [3]int, proto PBFProto) PBF {
	pbf := make(PBF)
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
