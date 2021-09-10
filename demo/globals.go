package demo

import "github.com/flywave/go-tileproxy/setting"

var (
	Globals = setting.GlobalsSetting{}
)

func init() {
	Globals.Srs.ProjDataDir = "../../proj_data"
	Globals.Geoid.GeoidDataDir = "../../geoid_data"

	Globals.Cache.BaseDir = "./cache_data"
	Globals.Cache.MetaSize = []uint32{4, 4}
	Globals.Cache.MetaBuffer = 80

	Globals.Grid.TileSize = []uint32{256, 256}

	Globals.Image.ResamplingMethod = "bicubic"
	Globals.Image.MaxStretchFactor = setting.NewFloat64(1.15)
	Globals.Image.MaxShrinkFactor = setting.NewFloat64(4.0)
	Globals.Image.Paletted = setting.NewBool(false)
	Globals.Image.FontDir = setting.NewString("../../imagery/fonts/")
	Globals.Image.Formats = map[string]setting.ImageOpts{
		"custom_format": {Format: "image/png", Mode: "rgba", Transparent: setting.NewBool(true)},
		"image/jpeg":    {Format: "image/jpeg", Mode: "rgb", Transparent: setting.NewBool(false), EncodingOptions: map[string]interface{}{"jpeg_quality": 90}},
	}
}
