package terrain

import (
	"testing"

	"github.com/flywave/go-geo"
)

func TestCaclulateGrid_Raster(t *testing.T) {
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)
	bbox := grid.TileBBox([3]int{1, 1, 1}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)
	georef := geo.NewGeoReference(bbox2, srs4326)

	tests := []struct {
		name   string
		width  int
		height int
		mode   BorderMode
		check  func(*testing.T, *Grid)
	}{
		{
			name:   "basic raster grid",
			width:  10,
			height: 10,
			mode:   BORDER_NONE,
			check: func(t *testing.T, g *Grid) {
				if g.Width != 10 || g.Height != 10 {
					t.Errorf("Expected 10x10 grid, got %dx%d", g.Width, g.Height)
				}
				if len(g.Coordinates) != 100 {
					t.Errorf("Expected 100 coordinates, got %d", len(g.Coordinates))
				}
				if g.Count != 100 {
					t.Errorf("Expected count 100, got %d", g.Count)
				}
			},
		},
		{
			name:   "raster grid with unilateral border",
			width:  10,
			height: 10,
			mode:   BORDER_UNILATERAL,
			check: func(t *testing.T, g *Grid) {
				if g.Width != 11 || g.Height != 11 {
					t.Errorf("Expected 11x11 grid with border, got %dx%d", g.Width, g.Height)
				}
				if len(g.Coordinates) != 121 {
					t.Errorf("Expected 121 coordinates with border, got %d", len(g.Coordinates))
				}
			},
		},
		{
			name:   "raster grid with bilateral border",
			width:  10,
			height: 10,
			mode:   BORDER_BILATERAL,
			check: func(t *testing.T, g *Grid) {
				if g.Width != 12 || g.Height != 12 {
					t.Errorf("Expected 12x12 grid with bilateral border, got %dx%d", g.Width, g.Height)
				}
				if len(g.Coordinates) != 144 {
					t.Errorf("Expected 144 coordinates with bilateral border, got %d", len(g.Coordinates))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &RasterOptions{Format: "raster", Mode: tt.mode}
			grid := CaclulateGrid(tt.width, tt.height, opts, georef)
			if grid == nil {
				t.Fatal("Expected non-nil grid")
			}
			if grid.srs != srs4326 {
				t.Errorf("Expected srs4326, got %v", grid.srs)
			}
			tt.check(t, grid)
		})
	}
}

func TestCaclulateGrid_Terrain(t *testing.T) {
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)
	bbox := grid.TileBBox([3]int{1, 1, 1}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)
	georef := geo.NewGeoReference(bbox2, srs4326)

	opts := &RasterOptions{Format: "terrain", Mode: BORDER_NONE}
	terrainGrid := CaclulateGrid(10, 10, opts, georef)

	if terrainGrid == nil {
		t.Fatal("Expected non-nil terrain grid")
	}
	if terrainGrid.Width != 10 || terrainGrid.Height != 10 {
		t.Errorf("Expected 10x10 terrain grid, got %dx%d", terrainGrid.Width, terrainGrid.Height)
	}
	if len(terrainGrid.Coordinates) != 100 {
		t.Errorf("Expected 100 coordinates for terrain, got %d", len(terrainGrid.Coordinates))
	}
}

func TestGrid_Methods(t *testing.T) {
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)
	bbox := grid.TileBBox([3]int{1, 1, 1}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)
	georef := geo.NewGeoReference(bbox2, srs4326)

	// Create a small grid for testing methods
	smallGrid := CaclulateGrid(3, 3, &RasterOptions{Format: "raster", Mode: BORDER_NONE}, georef)

	t.Run("GetBBox", func(t *testing.T) {
		bbox := smallGrid.GetBBox()
		if bbox.Min[0] >= bbox.Max[0] || bbox.Min[1] >= bbox.Max[1] {
			t.Errorf("Invalid bbox: %v", bbox)
		}
	})

	t.Run("GetRect", func(t *testing.T) {
		rect := smallGrid.GetRect()
		if rect.Min[0] >= rect.Max[0] || rect.Min[1] >= rect.Max[1] {
			t.Errorf("Invalid rect: %v", rect)
		}
	})

	t.Run("GetRay", func(t *testing.T) {
		rays := smallGrid.GetRay()
		if len(rays) != 9 {
			t.Errorf("Expected 9 rays, got %d", len(rays))
		}
		for _, ray := range rays {
			if ray.Direction[2] != -1 {
				t.Errorf("Expected ray direction to be (0,0,-1), got %v", ray.Direction)
			}
		}
	})

	t.Run("Sort", func(t *testing.T) {
		// Add some test data to coordinates
		for i := range smallGrid.Coordinates {
			smallGrid.Coordinates[i][2] = float64(i)
		}
		smallGrid.Sort()
		if smallGrid.Coordinates[0][2] != 0 || smallGrid.Coordinates[8][2] != 8 {
			t.Error("Sort did not work correctly")
		}
	})

	t.Run("Value", func(t *testing.T) {
		// Test Value method with known values
		for i := range smallGrid.Coordinates {
			smallGrid.Coordinates[i][2] = float64(i)
		}
		val := smallGrid.Value(1, 1)
		if val != 4 {
			t.Errorf("Expected value at (1,1) to be 4, got %f", val)
		}
	})

	t.Run("GetRange", func(t *testing.T) {
		smallGrid.Minimum = 100
		smallGrid.Maximum = 200
		rangeVal := smallGrid.GetRange()
		if rangeVal != 100 {
			t.Errorf("Expected range 100, got %f", rangeVal)
		}
	})
}

func TestGrid_GetTileDate(t *testing.T) {
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)
	bbox := grid.TileBBox([3]int{1, 1, 1}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)
	georef := geo.NewGeoReference(bbox2, srs4326)

	tests := []struct {
		name  string
		mode  BorderMode
		check func(*testing.T, *TileData)
	}{
		{
			name: "no border",
			mode: BORDER_NONE,
			check: func(t *testing.T, td *TileData) {
				if td == nil {
					t.Fatal("Expected non-nil TileData")
				}
				if td.Size[0] != 3 || td.Size[1] != 3 {
					t.Errorf("Expected 3x3 TileData, got %dx%d", td.Size[0], td.Size[1])
				}
			},
		},
		{
			name: "unilateral border",
			mode: BORDER_UNILATERAL,
			check: func(t *testing.T, td *TileData) {
				if td == nil {
					t.Fatal("Expected non-nil TileData")
				}
				if td.Size[0] != 3 || td.Size[1] != 3 {
					t.Errorf("Expected 3x3 TileData, got %dx%d", td.Size[0], td.Size[1])
				}
			},
		},
		{
			name: "bilateral border",
			mode: BORDER_BILATERAL,
			check: func(t *testing.T, td *TileData) {
				if td == nil {
					t.Fatal("Expected non-nil TileData")
				}
				if td.Size[0] != 3 || td.Size[1] != 3 {
					t.Errorf("Expected 3x3 TileData, got %dx%d", td.Size[0], td.Size[1])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := CaclulateGrid(3, 3, &RasterOptions{Format: "raster", Mode: tt.mode}, georef)
			tileData := grid.GetTileDate(tt.mode)
			tt.check(t, tileData)
		})
	}
}

func TestGrid_EdgeCases(t *testing.T) {
	srs900913 := geo.NewProj(900913)
	srs4326 := geo.NewProj(4326)

	conf := geo.DefaultTileGridOptions()
	conf[geo.TILEGRID_SRS] = srs900913
	conf[geo.TILEGRID_RES_FACTOR] = 2.0
	conf[geo.TILEGRID_TILE_SIZE] = []uint32{256, 256}
	conf[geo.TILEGRID_ORIGIN] = geo.ORIGIN_UL

	grid := geo.NewTileGrid(conf)
	bbox := grid.TileBBox([3]int{1, 1, 1}, false)
	bbox2 := srs900913.TransformRectTo(srs4326, bbox, 16)
	georef := geo.NewGeoReference(bbox2, srs4326)

	t.Run("empty grid", func(t *testing.T) {
		grid := CaclulateGrid(0, 0, &RasterOptions{Format: "raster", Mode: BORDER_NONE}, georef)
		if grid == nil {
			t.Fatal("Expected non-nil empty grid")
		}
		if grid.Count != 0 {
			t.Errorf("Expected count 0 for empty grid, got %d", grid.Count)
		}
	})

	t.Run("single pixel grid", func(t *testing.T) {
		grid := CaclulateGrid(1, 1, &RasterOptions{Format: "raster", Mode: BORDER_NONE}, georef)
		if grid == nil {
			t.Fatal("Expected non-nil single pixel grid")
		}
		if grid.Count != 1 {
			t.Errorf("Expected count 1 for single pixel grid, got %d", grid.Count)
		}
	})
}
