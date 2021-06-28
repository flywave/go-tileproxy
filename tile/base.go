package tile

import (
	"fmt"
	"math"

	"github.com/flywave/go-geom/generic"
)

var UnknownConversionError = fmt.Errorf("do not know how to convert value to requested value")

type BaseTile struct {
	Tile
	Z         uint
	X         uint
	Y         uint
	Lat       float64
	Long      float64
	Tolerance float64
	Extent    float64
	Buffer    float64
	cached    bool
	xspan     float64
	yspan     float64
	extent    *generic.Extent
	bufpext   *generic.Extent
}

func NewTileWithZXY(z, x, y uint) (t *BaseTile) {
	t = &BaseTile{
		Z:         z,
		X:         x,
		Y:         y,
		Buffer:    DefaultTileBuffer,
		Extent:    DefaultExtent,
		Tolerance: DefaultEpislon,
	}
	t.Lat, t.Long = t.Num2Deg()
	t.Init()
	return t
}

func NewTileLatLong(z uint, lat, lon float64) (t *BaseTile) {
	t = &BaseTile{
		Z:         z,
		Lat:       lat,
		Long:      lon,
		Buffer:    DefaultTileBuffer,
		Extent:    DefaultExtent,
		Tolerance: DefaultEpislon,
	}
	x, y := t.Deg2Num()
	t.X, t.Y = uint(x), uint(y)
	t.Init()
	return t
}

func (t *BaseTile) Init() {
	max := 20037508.34

	res := (max * 2) / math.Exp2(float64(t.Z))
	t.cached = true
	t.extent = &generic.Extent{
		-max + (float64(t.X) * res),       // MinX
		max - (float64(t.Y) * res),        // Miny
		-max + (float64(t.X) * res) + res, // MaxX
		max - (float64(t.Y) * res) - res,  // MaxY
	}
	t.xspan = t.extent.MaxX() - t.extent.MinX()
	t.yspan = t.extent.MaxY() - t.extent.MinY()

	t.bufpext = &generic.Extent{
		0 - t.Buffer, 0 - t.Buffer,
		t.Extent + t.Buffer, t.Extent + t.Buffer,
	}
}

func (t *BaseTile) Deg2Num() (x, y int) {
	x = int(math.Floor((t.Long + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))))
	y = int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))

	return x, y
}

func (t *BaseTile) Num2Deg() (lat, lng float64) {
	lat = Tile2Lat(uint64(t.Y), uint64(t.Z))
	lng = Tile2Lon(uint64(t.X), uint64(t.Z))
	return lat, lng
}

func Tile2Lon(x, z uint64) float64 { return float64(x)/math.Exp2(float64(z))*360.0 - 180.0 }

func Tile2Lat(y, z uint64) float64 {
	var n float64 = math.Pi
	if y != 0 {
		n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(z))
	}

	return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
}

func (t *BaseTile) Bounds() [4]float64 {
	north := Tile2Lon(uint64(t.X), uint64(t.Z))
	east := Tile2Lat(uint64(t.Y), uint64(t.Z))
	south := Tile2Lon(uint64(t.X+1), uint64(t.Z))
	west := Tile2Lat(uint64(t.Y+1), uint64(t.Z))
	return [4]float64{north, east, south, west}
}

func (t *BaseTile) PixelBufferedBounds() (bounds [4]float64, err error) {
	return t.bufpext.Extent(), nil
}

func (t *BaseTile) ZLevel() uint {
	return t.Z
}

func (t *BaseTile) ZRes() float64 {
	return 40075016.6855785 / (t.Extent * math.Exp2(float64(t.Z)))
}

func (t *BaseTile) ZEpislon() float64 {
	if t.Z == MaxZ {
		return 0
	}
	epi := t.Tolerance
	if epi <= 0 {
		return 0
	}
	ext := t.Extent

	denom := (math.Exp2(float64(t.Z)) * ext)

	e := epi / denom
	return e
}
