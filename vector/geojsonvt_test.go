package vector

import "testing"

func TestGeojsonVTSource(t *testing.T) {
	source := NewGeoJSONVTSource([3]int{13515, 6392, 14}, &GeoJSONVTOptions{})

	source.SetSource("../data/us-states-tiles.json")
	tile := source.GetTile()

	if tile == nil {
		t.FailNow()
	}
}
