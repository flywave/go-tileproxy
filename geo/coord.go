package geo

import "fmt"

type TileCoord struct {
	X, Y, Z int
}

func NewTileCoord(coords ...int) *TileCoord {
	if len(coords) < 3 {
		return nil
	}
	return &TileCoord{X: coords[0], Y: coords[1], Z: coords[2]}
}

func (t *TileCoord) Coords() [3]int {
	return [3]int{t.X, t.Y, t.Z}
}

func (t *TileCoord) TopLeftChild() *TileCoord {
	return &TileCoord{Z: t.Z + 1, X: t.X << 1, Y: t.Y << 1}
}

func (t *TileCoord) TopRightChild() *TileCoord {
	return &TileCoord{Z: t.Z + 1, X: (t.X << 1) + 1, Y: t.Y << 1}
}

func (t *TileCoord) BottomLeftChild() *TileCoord {
	return &TileCoord{Z: t.Z + 1, X: t.X << 1, Y: (t.Y << 1) + 1}
}

func (t *TileCoord) BottomRightChild() *TileCoord {
	return &TileCoord{Z: t.Z + 1, X: (t.X << 1) + 1, Y: (t.Y << 1) + 1}
}

func (t *TileCoord) Children() []TileCoord {
	z, x, y := t.Z+1, t.X<<1, t.Y<<1
	return []TileCoord{
		{Z: z, X: x, Y: y},
		{Z: z, X: x + 1, Y: y},
		{Z: z, X: x, Y: y + 1},
		{Z: z, X: x + 1, Y: y + 1},
	}
}

func (t *TileCoord) String() string {
	return fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
}
