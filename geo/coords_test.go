package geo

import "testing"

func TestTileCoord_Children(t *testing.T) {
	var tile = &TileCoord{Z: 10, X: 833, Y: 424}
	children := tile.Children()
	t.Logf("%+v", children)
}
