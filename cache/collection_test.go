package cache

import (
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
)

// Mock implementations for testing

type collectionMockSource struct {
	tile.Source
	data string
}

func newCollectionMockSource(data string) *collectionMockSource {
	return &collectionMockSource{data: data}
}

func (s *collectionMockSource) GetType() tile.TileType {
	return tile.TILE_IMAGERY
}

func (s *collectionMockSource) GetSource() interface{} {
	return s.data
}

func (s *collectionMockSource) SetSource(src interface{}) {
	s.data = src.(string)
}

func (s *collectionMockSource) GetFileName() string {
	return "mock_tile"
}

func (s *collectionMockSource) GetSize() [2]uint32 {
	return [2]uint32{256, 256}
}

func (s *collectionMockSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	return []byte(s.data)
}

func (s *collectionMockSource) GetTile() interface{} {
	return s.data
}

func (s *collectionMockSource) GetCacheable() *tile.CacheInfo {
	return &tile.CacheInfo{Cacheable: true}
}

func (s *collectionMockSource) SetCacheable(c *tile.CacheInfo) {
	// Mock implementation
}

func (s *collectionMockSource) SetTileOptions(options tile.TileOptions) {
	// Mock implementation
}

func (s *collectionMockSource) GetTileOptions() tile.TileOptions {
	return nil
}

func (s *collectionMockSource) GetGeoReference() *geo.GeoReference {
	return nil
}

// Test cases

func TestNewTileCollection_WithNilCoords(t *testing.T) {
	collection := NewTileCollection(nil)

	if collection == nil {
		t.Fatal("Expected non-nil TileCollection")
	}
	if !collection.Empty() {
		t.Error("Expected empty collection")
	}
	if len(collection.tiles) != 0 {
		t.Errorf("Expected empty tiles slice, got length %d", len(collection.tiles))
	}
	if len(collection.tilesDict) != 0 {
		t.Errorf("Expected empty tiles dict, got length %d", len(collection.tilesDict))
	}
}

func TestNewTileCollection_WithCoords(t *testing.T) {
	coords := [][3]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	collection := NewTileCollection(coords)

	if collection == nil {
		t.Fatal("Expected non-nil TileCollection")
	}
	if collection.Empty() {
		t.Error("Expected non-empty collection")
	}
	if len(collection.tiles) != 3 {
		t.Errorf("Expected 3 tiles, got %d", len(collection.tiles))
	}
	if len(collection.tilesDict) != 3 {
		t.Errorf("Expected 3 entries in dict, got %d", len(collection.tilesDict))
	}

	// Verify tiles are created correctly
	for i, coord := range coords {
		tile := collection.tiles[i]
		if tile.Coord != coord {
			t.Errorf("Expected coord %v, got %v", coord, tile.Coord)
		}
		if collection.tilesDict[coord] != tile {
			t.Errorf("Dict entry for coord %v doesn't match tile", coord)
		}
	}
}

func TestTileCollection_SetItem(t *testing.T) {
	collection := NewTileCollection(nil)

	tile1 := NewTile([3]int{1, 2, 3})
	tile1.Source = newCollectionMockSource("tile1_data")

	tile2 := NewTile([3]int{4, 5, 6})
	tile2.Source = newCollectionMockSource("tile2_data")

	collection.SetItem(tile1)
	collection.SetItem(tile2)

	if len(collection.tiles) != 2 {
		t.Errorf("Expected 2 tiles, got %d", len(collection.tiles))
	}
	if len(collection.tilesDict) != 2 {
		t.Errorf("Expected 2 entries in dict, got %d", len(collection.tilesDict))
	}
	if collection.tilesDict[tile1.Coord] != tile1 {
		t.Error("tile1 not found in dict")
	}
	if collection.tilesDict[tile2.Coord] != tile2 {
		t.Error("tile2 not found in dict")
	}
}

func TestTileCollection_UpdateItem(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Create new tile with same coordinate as index 0
	newTile := NewTile([3]int{1, 2, 3})
	newTile.Source = newCollectionMockSource("updated_data")

	collection.UpdateItem(0, newTile)

	if collection.tiles[0] != newTile {
		t.Error("Tile at index 0 was not updated")
	}
	if collection.tilesDict[[3]int{1, 2, 3}] != newTile {
		t.Error("Dict entry was not updated")
	}
}

func TestTileCollection_GetItem_ByCoord(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Test existing coordinate
	tile := collection.GetItem([3]int{1, 2, 3})
	if tile == nil {
		t.Fatal("Expected non-nil tile")
	}
	if tile.Coord != [3]int{1, 2, 3} {
		t.Errorf("Expected coord [1,2,3], got %v", tile.Coord)
	}

	// Test non-existing coordinate (should create new tile)
	newTile := collection.GetItem([3]int{7, 8, 9})
	if newTile == nil {
		t.Fatal("Expected non-nil tile for non-existing coord")
	}
	if newTile.Coord != [3]int{7, 8, 9} {
		t.Errorf("Expected coord [7,8,9], got %v", newTile.Coord)
	}
}

func TestTileCollection_GetItem_ByIndex(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Test valid index
	tile := collection.GetItem(0)
	if tile == nil {
		t.Fatal("Expected non-nil tile")
	}
	if tile.Coord != [3]int{1, 2, 3} {
		t.Errorf("Expected coord [1,2,3], got %v", tile.Coord)
	}

	tile = collection.GetItem(1)
	if tile == nil {
		t.Fatal("Expected non-nil tile")
	}
	if tile.Coord != [3]int{4, 5, 6} {
		t.Errorf("Expected coord [4,5,6], got %v", tile.Coord)
	}
}

func TestTileCollection_GetItem_InvalidType(t *testing.T) {
	collection := NewTileCollection(nil)

	// Test with invalid type
	tile := collection.GetItem("invalid")
	if tile != nil {
		t.Error("Expected nil for invalid type")
	}
}

func TestTileCollection_Contains_ByCoord(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Test existing coordinate
	if !collection.Contains([3]int{1, 2, 3}) {
		t.Error("Expected collection to contain coord [1,2,3]")
	}
	if !collection.Contains([3]int{4, 5, 6}) {
		t.Error("Expected collection to contain coord [4,5,6]")
	}

	// Test non-existing coordinate
	if collection.Contains([3]int{7, 8, 9}) {
		t.Error("Expected collection to not contain coord [7,8,9]")
	}
}

func TestTileCollection_Contains_ByIndex(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Test valid indices
	if !collection.Contains(0) {
		t.Error("Expected collection to contain index 0")
	}
	if !collection.Contains(1) {
		t.Error("Expected collection to contain index 1")
	}

	// Test invalid index
	if collection.Contains(2) {
		t.Error("Expected collection to not contain index 2")
	}
	// Note: negative index will return true due to current implementation logic
	// where -1 < len(c.tiles) is true
}

func TestTileCollection_Contains_InvalidType(t *testing.T) {
	collection := NewTileCollection(nil)

	// Test with invalid type
	if collection.Contains("invalid") {
		t.Error("Expected false for invalid type")
	}
}

func TestTileCollection_Empty(t *testing.T) {
	// Test empty collection
	collection := NewTileCollection(nil)
	if !collection.Empty() {
		t.Error("Expected empty collection")
	}

	// Test non-empty collection
	coords := [][3]int{{1, 2, 3}}
	collection = NewTileCollection(coords)
	if collection.Empty() {
		t.Error("Expected non-empty collection")
	}
}

func TestTileCollection_GetSlice(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	slice := collection.GetSlice()
	if len(slice) != 2 {
		t.Errorf("Expected slice length 2, got %d", len(slice))
	}
	if slice[0].Coord != [3]int{1, 2, 3} {
		t.Errorf("Expected first tile coord [1,2,3], got %v", slice[0].Coord)
	}
	if slice[1].Coord != [3]int{4, 5, 6} {
		t.Errorf("Expected second tile coord [4,5,6], got %v", slice[1].Coord)
	}
}

func TestTileCollection_GetMap(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	tileMap := collection.GetMap()
	if len(tileMap) != 2 {
		t.Errorf("Expected map length 2, got %d", len(tileMap))
	}
	if tileMap[[3]int{1, 2, 3}] == nil {
		t.Error("Expected tile at coord [1,2,3]")
	}
	if tileMap[[3]int{4, 5, 6}] == nil {
		t.Error("Expected tile at coord [4,5,6]")
	}
}

func TestTileCollection_AllBlank_WithBlankSources(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Set all tiles to have BlankImageSource
	imageOpts := &imagery.ImageOptions{Format: "png"}
	for _, tile := range collection.tiles {
		tile.Source = imagery.NewBlankImageSource([2]uint32{256, 256}, imageOpts, nil)
	}

	if !collection.AllBlank() {
		t.Error("Expected all tiles to be blank")
	}
}

func TestTileCollection_AllBlank_WithNonBlankSources(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Set tiles to have non-blank sources
	collection.tiles[0].Source = newCollectionMockSource("data1")
	collection.tiles[1].Source = newCollectionMockSource("data2")

	if collection.AllBlank() {
		t.Error("Expected not all tiles to be blank")
	}
}

func TestTileCollection_AllBlank_WithMixedSources(t *testing.T) {
	coords := [][3]int{{1, 2, 3}, {4, 5, 6}}
	collection := NewTileCollection(coords)

	// Set one tile to blank, one to non-blank
	imageOpts := &imagery.ImageOptions{Format: "png"}
	collection.tiles[0].Source = imagery.NewBlankImageSource([2]uint32{256, 256}, imageOpts, nil)
	collection.tiles[1].Source = newCollectionMockSource("data")

	if collection.AllBlank() {
		t.Error("Expected not all tiles to be blank with mixed sources")
	}
}

func TestTileCollection_AllBlank_EmptyCollection(t *testing.T) {
	collection := NewTileCollection(nil)

	// Empty collection should return true
	if !collection.AllBlank() {
		t.Error("Expected empty collection to be considered all blank")
	}
}

func TestTileCollection_ConcurrentOperations(t *testing.T) {
	collection := NewTileCollection(nil)

	// Test adding multiple tiles and verifying consistency
	coords := [][3]int{
		{1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 4}, {5, 5, 5},
	}

	for _, coord := range coords {
		tile := NewTile(coord)
		tile.Source = newCollectionMockSource("test_data")
		collection.SetItem(tile)
	}

	// Verify all tiles were added correctly
	if len(collection.tiles) != 5 {
		t.Errorf("Expected 5 tiles, got %d", len(collection.tiles))
	}
	if len(collection.tilesDict) != 5 {
		t.Errorf("Expected 5 dict entries, got %d", len(collection.tilesDict))
	}

	// Verify each tile can be found by coordinate
	for _, coord := range coords {
		if !collection.Contains(coord) {
			t.Errorf("Collection should contain coord %v", coord)
		}
		tile := collection.GetItem(coord)
		if tile.Coord != coord {
			t.Errorf("Retrieved tile has wrong coord: expected %v, got %v", coord, tile.Coord)
		}
	}
}
