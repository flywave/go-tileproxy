package imports

import "testing"

func TestArchiveImport(t *testing.T) {
	import_ := NewArchiveImport("../data/test_import.tar.gz", nil)

	err := import_.Open()

	if err != nil {
		t.FailNow()
	}

	ext := import_.GetExtension()

	if ext != "png" {
		t.FailNow()
	}

	tile, err := import_.LoadTileCoord([3]int{1, 1, 1})

	if err != nil {
		t.FailNow()
	}

	if tile.Source == nil {
		t.FailNow()
	}

	import_.Close()
}

func TestGeoPackageImport(t *testing.T) {
	import_ := NewGeoPackageImport("../data/test_import.gpkg", nil)

	err := import_.Open()

	if err != nil {
		t.FailNow()
	}

	tile, err := import_.LoadTileCoord([3]int{1, 1, 1})

	if err != nil {
		t.FailNow()
	}

	if tile.Source == nil {
		t.FailNow()
	}

	import_.Close()
}

func TestMBTilesImport(t *testing.T) {
	import_ := NewMBTilesImport("../data/test_import.mbtils", nil)

	err := import_.Open()

	if err != nil {
		t.FailNow()
	}

	tile, err := import_.LoadTileCoord([3]int{1, 1, 1})

	if err != nil {
		t.FailNow()
	}

	if tile.Source == nil {
		t.FailNow()
	}

	import_.Close()
}
