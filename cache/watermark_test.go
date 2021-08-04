package cache

import (
	"os"
	"testing"

	"github.com/flywave/go-tileproxy/geo"
	"github.com/flywave/go-tileproxy/images"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/imaging"
)

func TestWatermark(t *testing.T) {
	images.SetFontPath("../images/fonts/")

	tc := NewTile([3]int{0, 0, 0})
	filter := NewWatermark("test", nil, nil, nil, nil)

	img_opts := &images.ImageOptions{Format: tile.TileFormat("image/png")}
	img_opts.Transparent = geo.NewBool(true)
	img := images.CreateImageSource([2]uint32{100, 100}, img_opts)
	tc.Source = img

	tc = filter.Apply(tc)

	imaging.Save(tc.GetSourceImage(), "./test.png")

	defer os.Remove("./test.png")
}
