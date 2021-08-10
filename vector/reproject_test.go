package vector

import (
	"testing"

	"github.com/flywave/go-tileproxy/geo"
)

func TestVectorTransformer(t *testing.T) {
	srs4326 := geo.NewSRSProj4("EPSG:4326")
	pgcj02 := geo.NewGCJ02Proj(true)

	tran := NewVectorTransformer(pgcj02, srs4326)

	source := NewMVTSource([3]int{1686, 776, 11}, PBF_PTOTO_LUOKUANG, &VectorOptions{Format: PBF_MIME_LUOKUANG})

	source.SetSource("../data/tile.pbf")
	tile := source.GetTile()

	feats := tile.(PBF)

	newfeats := make(PBF)

	for k, f := range feats {
		newfeats[k] = tran.Apply(f)
	}

	if newfeats == nil {
		t.FailNow()
	}
}
