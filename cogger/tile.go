package cogger

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"regexp"
	"strconv"

	"github.com/lanrat/extsort"
)

func TileLatitude(zoom uint8, y uint32) float64 {
	yf := 1.0 - 2.0*float64(y)/float64(uint32(1)<<zoom)
	return math.Atan(math.Sinh(math.Pi * yf))
}

func TileArea(zoom uint8, y uint32) float64 {
	earthSurface := 510065623.0 // in km²
	latFraction := (TileLatitude(zoom, y) - TileLatitude(zoom, y+1)) / math.Pi
	return earthSurface * latFraction / float64(uint32(1)<<zoom)
}

type TileKey uint64

const WorldTile = TileKey(0)

const NoTile = TileKey(^uint64(0x1f))

func MakeTileKey(zoom uint8, x, y uint32) TileKey {
	val := uint64(zoom)
	shift := uint8(64 - 2*zoom)
	for bit := uint8(0); bit < zoom; bit++ {
		xm := uint64((x>>bit)&1) << shift
		ym := uint64((y>>bit)&1) << (shift + 1)
		val |= xm | ym
		shift += 2
	}
	return TileKey(val)
}

func (t TileKey) Contains(other TileKey) bool {
	zoom := t.Zoom()
	otherZoom := other.Zoom()
	if otherZoom > zoom {
		return t == other.ToZoom(zoom)
	} else {
		return false
	}
}

func (t TileKey) Zoom() uint8 {
	return uint8(t) & 0x1f
}

func (t TileKey) ZoomXY() (zoom uint8, x, y uint32) {
	val := uint64(t)
	zoom = uint8(val) & 0x1f
	shift := uint8(64 - 2*zoom)
	for bit := uint8(0); bit < zoom; bit++ {
		x |= (uint32(val>>shift) & 1) << bit
		y |= (uint32(val>>(shift+1)) & 1) << bit
		shift += 2
	}
	return zoom, x, y
}

func (t TileKey) ToZoom(z uint8) TileKey {
	val := uint64(t)
	shift := uint8(64 - 2*z)
	return TileKey(((val >> shift) << shift) | uint64(z))
}

func (t TileKey) String() string {
	if t == NoTile {
		return "NoTile"
	}

	zoom, x, y := t.ZoomXY()
	return fmt.Sprintf("%d/%d/%d", zoom, x, y)
}

func (t TileKey) Next(maxZoom uint8) TileKey {
	zoom := uint8(t) & 0x1f

	if zoom < maxZoom {
		return TileKey(uint64(t) & ^uint64(0x1f) | uint64(zoom+1))
	}

	shift := uint8(64 - 2*maxZoom)
	val := uint64(t) >> shift

	if bits.OnesCount64(val) == int(2*maxZoom) {
		return NoTile
	}

	val = val + 1
	newZoom := maxZoom - uint8(bits.TrailingZeros64(val)/2)
	return TileKey(val<<shift | uint64(newZoom))
}

type TileCount struct {
	Key   TileKey
	Count uint64
}

var tileLogRegexp = regexp.MustCompile(`^(\d+)/(\d+)/(\d+)\s+(\d+)$`)

func ParseTileCount(s string) TileCount {
	match := tileLogRegexp.FindStringSubmatch(s)
	if match == nil || len(match) != 5 {
		return TileCount{NoTile, 0}
	}
	zoom, _ := strconv.Atoi(match[1])
	if zoom < 0 || zoom > 24 {
		return TileCount{NoTile, 0}
	}
	x, _ := strconv.ParseUint(match[2], 10, 32)
	y, _ := strconv.ParseUint(match[3], 10, 32)
	if x >= 1<<zoom || y >= 1<<zoom {
		return TileCount{NoTile, 0}
	}
	count, _ := strconv.ParseUint(match[4], 10, 64)
	key := MakeTileKey(uint8(zoom), uint32(x), uint32(y))
	return TileCount{Key: key, Count: count}
}

func (c TileCount) ToBytes() []byte {
	zoom, x, y := c.Key.ZoomXY()
	var buf [binary.MaxVarintLen32*2 + binary.MaxVarintLen64 + 1]byte
	pos := binary.PutUvarint(buf[:], uint64(x))
	pos += binary.PutUvarint(buf[pos:], uint64(y))
	pos += binary.PutUvarint(buf[pos:], c.Count)
	buf[pos] = zoom
	pos += 1
	return buf[0:pos]
}

func TileCountFromBytes(b []byte) extsort.SortType {
	x, pos := binary.Uvarint(b)
	y, len := binary.Uvarint(b[pos:])
	pos += len
	count, len := binary.Uvarint(b[pos:])
	pos += len
	zoom := b[pos]
	key := MakeTileKey(zoom, uint32(x), uint32(y))
	return TileCount{Key: key, Count: count}
}

func TileCountLess(a, b extsort.SortType) bool {
	aa := a.(TileCount)
	bb := b.(TileCount)
	if aa.Key != bb.Key {
		return aa.Key < bb.Key
	} else {
		return aa.Count < bb.Count
	}
}
