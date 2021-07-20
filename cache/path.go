package cache

import (
	"errors"
	"fmt"
	"os"
	"path"
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

func ensure_directory(location string) {
	os.MkdirAll(location, os.ModePerm)
}

func tile_location_tc(tile *Tile, cache_dir string, file_ext string, create_dir bool) string {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		parts := []string{cache_dir,
			level_part(z),
			fmt.Sprintf("%03d", int(x/1000000)),
			fmt.Sprintf("%03d", (int(x/1000) % 1000)),
			fmt.Sprintf("%03d", (int(x) % 1000)),
			fmt.Sprintf("%03d", int(y/1000000)),
			fmt.Sprintf("%03d", (int(y/1000) % 1000)),
			fmt.Sprintf("%03d.%s", (int(y) % 1000), file_ext)}
		tile.Location = path.Join(parts...)
	}
	if create_dir {
		ensure_directory(tile.Location)
	}
	return tile.Location
}

func tile_location_mp(tile *Tile, cache_dir string, file_ext string, create_dir bool) string {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		parts := []string{cache_dir,
			level_part(z),
			fmt.Sprintf("%04d", int(x/10000)),
			fmt.Sprintf("%04d", (int(x) % 10000)),
			fmt.Sprintf("%04d", int(y/10000)),
			fmt.Sprintf("%04d.%s", (int(y) % 10000), file_ext)}
		tile.Location = path.Join(parts...)
	}
	if create_dir {
		ensure_directory(tile.Location)
	}
	return tile.Location
}

func tile_location_tms(tile *Tile, cache_dir string, file_ext string, create_dir bool) string {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		tile.Location = path.Join(
			cache_dir, level_part(strconv.Itoa(z)),
			strconv.Itoa(x), strconv.Itoa(y)+"."+file_ext)
	}
	if create_dir {
		ensure_directory(tile.Location)
	}
	return tile.Location
}

func tile_location_reverse_tms(tile *Tile, cache_dir string, file_ext string, create_dir bool) string {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		tile.Location = path.Join(
			cache_dir, strconv.Itoa(y), strconv.Itoa(x), strconv.Itoa(z)+"."+file_ext)
	}
	if create_dir {
		ensure_directory(tile.Location)
	}
	return tile.Location
}

func tile_location_quadkey(tile *Tile, cache_dir string, file_ext string, create_dir bool) string {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		quadKey := ""
		for i := range []int{z, 0, -1} {
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
	}
	if create_dir {
		ensure_directory(tile.Location)
	}
	return tile.Location
}

func tile_location_arcgiscache(tile *Tile, cache_dir string, file_ext string, create_dir bool) string {
	if tile.Location == "" {
		x, y, z := tile.Coord[0], tile.Coord[1], tile.Coord[2]
		parts := []string{cache_dir, fmt.Sprintf("L%02d", z), fmt.Sprintf("R%08x", y), fmt.Sprintf("C%08x.%s", x, file_ext)}
		tile.Location = path.Join(parts...)
	}
	if create_dir {
		ensure_directory(tile.Location)
	}
	return tile.Location
}

func LocationPaths(layout string) (func(*Tile, string, string, bool) string, func(int, string) string, error) {
	if layout == "tc" {
		return tile_location_tc, level_location, nil
	} else if layout == "mp" {
		return tile_location_mp, level_location, nil
	} else if layout == "tms" {
		return tile_location_tms, level_location, nil
	} else if layout == "reverse_tms" {
		return tile_location_reverse_tms, nil, nil
	} else if layout == "quadkey" {
		return tile_location_quadkey, no_level_location, nil
	} else if layout == "arcgis" {
		return tile_location_arcgiscache, level_location_arcgiscache, nil
	}
	return nil, nil, errors.New(fmt.Sprintf("unknown directory_layout \"%s\"", layout))
}
