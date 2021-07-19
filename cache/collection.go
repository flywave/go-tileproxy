package cache

type TileCollection struct {
	tiles     []*Tile
	tilesDict map[[3]int]*Tile
}

func NewTileCollection(tileCoords [][3]int) *TileCollection {
	ret := &TileCollection{}
	ret.tiles = make([]*Tile, len(tileCoords))
	ret.tilesDict = make(map[[3]int]*Tile)
	for i := range tileCoords {
		ret.tiles[i] = NewTile(tileCoords[i])
		ret.tilesDict[tileCoords[i]] = ret.tiles[i]
	}
	return ret
}

func (c *TileCollection) GetItem(idx_or_coord interface{}) *Tile {
	switch v := idx_or_coord.(type) {
	case [3]int:
		if t, ok := c.tilesDict[v]; ok {
			return t
		} else {
			return NewTile(v)
		}
	case int:
		return c.tiles[v]
	}
	return nil
}

func (c *TileCollection) Contains(idx_or_coord interface{}) bool {
	switch v := idx_or_coord.(type) {
	case [3]int:
		_, ok := c.tilesDict[v]
		return ok
	case int:
		return v < len(c.tiles)
	}
	return false
}

func (c *TileCollection) GetSlice() []*Tile {
	return c.tiles
}

func (c *TileCollection) GetMap() map[[3]int]*Tile {
	return c.tilesDict
}
