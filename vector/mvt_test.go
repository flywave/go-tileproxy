package vector

import (
	"io/ioutil"
	"testing"

	"github.com/flywave/go-mapbox/mvt"
	m "github.com/flywave/go-mapbox/tileid"
)

var bytevals1, _ = ioutil.ReadFile("../data/3194.mvt")
var tileid1 = m.TileID{X: 13515, Y: 6392, Z: 14}
var bytevals2, _ = ioutil.ReadFile("../data/tile.pbf")
var tileid2 = m.TileID{X: 1686, Y: 776, Z: 11}

func TestMVTSource(t *testing.T) {
	source := NewMVTSource([3]int{13515, 6392, 14}, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})

	source.SetSource("../data/3194.mvt")
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}

func TestMVTSourceBuffer(t *testing.T) {
	source := NewMVTSource([3]int{13515, 6392, 14}, PBF_PTOTO_MAPBOX, &VectorOptions{Format: MVT_MIME})

	source.SetSource(bytevals1)
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}

func TestPBFSource(t *testing.T) {
	source := NewMVTSource([3]int{1686, 776, 11}, PBF_PTOTO_LUOKUANG, &VectorOptions{Format: PBF_MIME, Proto: int(mvt.PROTO_LK)})

	source.SetSource("../data/tile.pbf")
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}

func TestPBFSourceBuffer(t *testing.T) {
	source := NewMVTSource([3]int{1686, 776, 11}, PBF_PTOTO_LUOKUANG, &VectorOptions{Format: PBF_MIME, Proto: int(mvt.PROTO_LK)})

	source.SetSource(bytevals2)
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}
