package cache

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func level_location(level int, cache_dir string) string {
	return path.Join(cache_dir, fmt.Sprintf("%02d", level))
}

func no_level_location(level int, cache_dir string) string {
	panic("cache does not have any level location")
}

func level_location_arcgiscache(z int, cache_dir string) string {
	level := fmt.Sprintf("L%02d", z)
	return path.Join(cache_dir, level)
}

func level_part(level interface{}) string {
	switch t := level.(type) {
	case string:
		return t
	case int:
		return fmt.Sprintf("%02d", t)
	}
	return ""
}

func validateCoordinates(x, y, z int) error {
	const maxZoom = 30
	if z < 0 || z > maxZoom {
		return fmt.Errorf("invalid zoom level: %d", z)
	}
	if x < 0 || y < 0 {
		return fmt.Errorf("negative coordinates not allowed: x=%d, y=%d", x, y)
	}
	return nil
}

func ensure_directory(location string) error {
	if err := os.MkdirAll(location, 0755); err != nil {
		return err
	}
	return nil
}

func tile_location_tc(tile *Tile, cache_dir string, file_ext string, create_dir bool) (string, error) {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		if err := validateCoordinates(x, y, z); err != nil {
			return "", err
		}
		parts := []string{cache_dir,
			level_part(z),
			fmt.Sprintf("%03d", int(x/1000000)),
			fmt.Sprintf("%03d", (int(x/1000) % 1000)),
			fmt.Sprintf("%03d", (int(x) % 1000)),
			fmt.Sprintf("%03d", int(y/1000000)),
			fmt.Sprintf("%03d", (int(y/1000) % 1000)),
			fmt.Sprintf("%03d.%s", (int(y) % 1000), file_ext)}
		tile.Location = path.Join(parts...)
		tile.Location = filepath.Clean(tile.Location)
	}
	if create_dir {
		dir := path.Dir(tile.Location)
		if err := ensure_directory(dir); err != nil {
			return "", err
		}
	}
	return tile.Location, nil
}

func tile_location_mp(tile *Tile, cache_dir string, file_ext string, create_dir bool) (string, error) {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		if err := validateCoordinates(x, y, z); err != nil {
			return "", err
		}
		parts := []string{cache_dir,
			level_part(z),
			fmt.Sprintf("%04d", int(x/10000)),
			fmt.Sprintf("%04d", (int(x) % 10000)),
			fmt.Sprintf("%04d", int(y/10000)),
			fmt.Sprintf("%04d.%s", (int(y) % 10000), file_ext)}
		tile.Location = path.Join(parts...)
		tile.Location = filepath.Clean(tile.Location)
	}
	if create_dir {
		dir := path.Dir(tile.Location)
		if err := ensure_directory(dir); err != nil {
			return "", err
		}
	}
	return tile.Location, nil
}

func tile_location_tms(tile *Tile, cache_dir string, file_ext string, create_dir bool) (string, error) {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		if err := validateCoordinates(x, y, z); err != nil {
			return "", err
		}
		tile.Location = path.Join(
			cache_dir, level_part(strconv.Itoa(z)),
			strconv.Itoa(x), strconv.Itoa(y)+"."+file_ext)
		tile.Location = filepath.Clean(tile.Location)
	}
	if create_dir {
		dir := path.Dir(tile.Location)
		if err := ensure_directory(dir); err != nil {
			return "", err
		}
	}
	return tile.Location, nil
}

func tile_location_reverse_tms(tile *Tile, cache_dir string, file_ext string, create_dir bool) (string, error) {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		if err := validateCoordinates(x, y, z); err != nil {
			return "", err
		}
		tile.Location = path.Join(cache_dir, strconv.Itoa(y), strconv.Itoa(x), strconv.Itoa(z)+"."+file_ext)
		tile.Location = filepath.Clean(tile.Location)
	}
	if create_dir {
		dir := path.Dir(tile.Location)
		if err := ensure_directory(dir); err != nil {
			return "", err
		}
	}
	return tile.Location, nil
}

func tile_location_quadkey(tile *Tile, cache_dir string, file_ext string, create_dir bool) (string, error) {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		if err := validateCoordinates(x, y, z); err != nil {
			return "", err
		}
		quadKey := ""
		for i := z; i > 0; i-- {
			if i > 62 {
				return "", fmt.Errorf("zoom level too large for quadkey: %d", z)
			}
			digit := 0
			mask := 1 << (i - 1)
			if (x & mask) != 0 {
				digit += 1
			}
			if (y & mask) != 0 {
				digit += 2
			}
			quadKey += strconv.Itoa(digit)
		}
		tile.Location = path.Join(
			cache_dir, quadKey+"."+file_ext)
		tile.Location = filepath.Clean(tile.Location)
	}
	if create_dir {
		dir := path.Dir(tile.Location)
		if err := ensure_directory(dir); err != nil {
			return "", err
		}
	}
	return tile.Location, nil
}

func tile_location_arcgiscache(tile *Tile, cache_dir string, file_ext string, create_dir bool) (string, error) {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		if err := validateCoordinates(x, y, z); err != nil {
			return "", err
		}
		parts := []string{cache_dir, fmt.Sprintf("L%02d", z), fmt.Sprintf("R%08x", y), fmt.Sprintf("C%08x.%s", x, file_ext)}
		tile.Location = path.Join(parts...)
		tile.Location = filepath.Clean(tile.Location)
	}
	if create_dir {
		dir := path.Dir(tile.Location)
		if err := ensure_directory(dir); err != nil {
			return "", err
		}
	}
	return tile.Location, nil
}

type TileLocationFunc func(*Tile, string, string, bool) (string, error)

func LocationPaths(layout string) (TileLocationFunc, func(int, string) string, error) {
	switch layout {
	case "tc":
		return tile_location_tc, level_location, nil
	case "mp":
		return tile_location_mp, level_location, nil
	case "tms":
		return tile_location_tms, level_location, nil
	case "reverse_tms":
		return tile_location_reverse_tms, nil, nil
	case "quadkey":
		return tile_location_quadkey, no_level_location, nil
	case "arcgis":
		return tile_location_arcgiscache, level_location_arcgiscache, nil
	}
	return nil, nil, fmt.Errorf("unknown directory_layout \"%s\"", layout)
}
