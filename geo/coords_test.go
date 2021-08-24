package geo

import "testing"

func TestTileChildren(t *testing.T) {
	var tile = &Tile{Z: 10, X: 833, Y: 424}
	children := tile.Children()
	t.Logf("%+v", children)
}
