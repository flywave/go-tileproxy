package client

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/tile"
)

type TileClient struct {
	BaseClient
	Grid     *geo.TileGrid
	Template *TileURLTemplate
}

func NewTileClient(grid *geo.TileGrid, tpl *TileURLTemplate) *TileClient {
	return &TileClient{Grid: grid, Template: tpl}
}

func (c *TileClient) GetTile(tile_coord [3]int, format *tile.TileFormat) []byte {
	url := c.Template.substitute(tile_coord, format, c.Grid)
	return c.Get(url).Body
}

func tilecachePath(tile_coord [3]int) string {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]
	parts := []string{fmt.Sprintf("%02d", z),
		fmt.Sprintf("%03d", int(x/1000000)),
		fmt.Sprintf("%03d", (int(x/1000) % 1000)),
		fmt.Sprintf("%03d", (int(x) % 1000)),
		fmt.Sprintf("%03d", int(y/1000000)),
		fmt.Sprintf("%03d", (int(y/1000) % 1000)),
		fmt.Sprintf("%03d", (int(y) % 1000))}
	return strings.Join(parts, "/")
}

func quadKey(tile_coord [3]int) string {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]
	quadKey := ""
	for _, i := range [3]int{z, 0, -1} {
		digit := 0
		mask := 1 << (i - 1)
		if (x & mask) != 0 {
			digit += 1
		}
		if (y & mask) != 0 {
			digit += 2
		}
		quadKey += strconv.FormatInt(int64(digit), 10)
	}
	return quadKey
}

func tmsPath(tile_coord [3]int) string {
	return fmt.Sprintf("%d/%d/%d", tile_coord[2], tile_coord[0], tile_coord[1])
}

func arcgisCachePath(tile_coord [3]int) string {
	return fmt.Sprintf("L%02d/R%08x/C%08x", tile_coord[2], tile_coord[1], tile_coord[0])
}

func bbox(tile_coord [3]int, grid *geo.TileGrid) string {
	rect := grid.TileBBox(tile_coord, false)
	return fmt.Sprintf("%.8f,%.8f,%.8f,%.8f", rect.Min[0], rect.Min[1], rect.Max[0], rect.Max[1])
}

type TileURLTemplate struct {
	Template            string
	Format              string
	WithQuadkey         bool
	WithTCPath          bool
	WithTMSPath         bool
	WithArcgisCachePath bool
	WithBBox            bool
}

func NewURLTemplate(template string, format string) *TileURLTemplate {
	rt := &TileURLTemplate{Template: template, Format: format}

	if strings.Contains(template, "{{ .quadkey }}") {
		rt.WithQuadkey = true
	} else {
		rt.WithQuadkey = false
	}

	if strings.Contains(template, "{{ .tc_path }}") {
		rt.WithTCPath = true
	} else {
		rt.WithTCPath = false
	}

	if strings.Contains(template, "{{ .tms_path }}") {
		rt.WithTMSPath = true
	} else {
		rt.WithTMSPath = false
	}

	if strings.Contains(template, "{{ .arcgiscache_path }}") {
		rt.WithArcgisCachePath = true
	} else {
		rt.WithArcgisCachePath = false
	}

	if strings.Contains(template, "{{ .bbox }}") {
		rt.WithBBox = true
	} else {
		rt.WithBBox = false
	}

	return rt
}

func (t *TileURLTemplate) substitute(tile_coord [3]int, format *tile.TileFormat, grid *geo.TileGrid) string {
	x, y, z := tile_coord[0], tile_coord[1], tile_coord[2]
	data := map[string]string{"x": strconv.FormatInt(int64(x), 10), "y": strconv.FormatInt(int64(y), 10), "z": strconv.FormatInt(int64(z), 10)}
	if format != nil {
		data["format"] = format.Extension()
	} else {
		data["format"] = t.Format
	}
	if t.WithQuadkey {
		data["quadkey"] = quadKey(tile_coord)
	}
	if t.WithTCPath {
		data["tc_path"] = tilecachePath(tile_coord)
	}
	if t.WithTMSPath {
		data["tms_path"] = tmsPath(tile_coord)
	}
	if t.WithArcgisCachePath {
		data["arcgiscache_path"] = arcgisCachePath(tile_coord)
	}
	if t.WithBBox {
		data["bbox"] = bbox(tile_coord, grid)
	}

	tmpl, err := template.New("test").Parse(t.Template)
	if err != nil {
		return ""
	}
	wr := &bytes.Buffer{}
	tmpl.Execute(wr, data)

	return string(wr.Bytes())
}

func (t *TileURLTemplate) ToString() string {
	return fmt.Sprintf("(%s, format=%s)", t.Template, t.Format)
}
