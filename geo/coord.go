package geo

import "fmt"

type Tile struct {
	X, Y, Z int
}

func NewTile(coords ...int) *Tile {
	if len(coords) < 3 {
		return nil
	}
	return &Tile{X: coords[0], Y: coords[1], Z: coords[2]}
}

func (t *Tile) Coords() [3]int {
	return [3]int{t.X, t.Y, t.Z}
}

func (t *Tile) TopLeftChild() *Tile {
	return &Tile{Z: t.Z + 1, X: t.X << 1, Y: t.Y << 1}
}

func (t *Tile) TopRightChild() *Tile {
	return &Tile{Z: t.Z + 1, X: (t.X << 1) + 1, Y: t.Y << 1}
}

func (t *Tile) BottomLeftChild() *Tile {
	return &Tile{Z: t.Z + 1, X: t.X << 1, Y: (t.Y << 1) + 1}
}

func (t *Tile) BottomRightChild() *Tile {
	return &Tile{Z: t.Z + 1, X: (t.X << 1) + 1, Y: (t.Y << 1) + 1}
}

func (t *Tile) Children() []Tile {
	z, x, y := t.Z+1, t.X<<1, t.Y<<1
	return []Tile{
		{Z: z, X: x, Y: y},
		{Z: z, X: x + 1, Y: y},
		{Z: z, X: x, Y: y + 1},
		{Z: z, X: x + 1, Y: y + 1},
	}
}

func (t *Tile) String() string {
	return fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)
}
