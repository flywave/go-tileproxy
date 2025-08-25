package cache

import (
	"os"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/imagery"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/imaging"
)

func TestWatermark(t *testing.T) {
	imagery.SetFontPath("../fonts/")

	tc := NewTile([3]int{0, 0, 0})
	filter := NewWatermark("flywave.net", nil, nil, nil, nil)

	img_opts := &imagery.ImageOptions{Format: tile.TileFormat("image/png")}
	img_opts.Transparent = geo.NewBool(true)
	img := imagery.CreateImageSource([2]uint32{256, 256}, img_opts)
	tc.Source = img

	tc, _ = filter.Apply(tc)

	imaging.Save(tc.GetSourceImage(), "./test.png")

	defer os.Remove("./test.png")
}
