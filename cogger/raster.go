package cogger

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
)

type RasterData struct {
	tile   TileKey
	pixels []float64
}

func NewRasterData(tile TileKey, pixels []float64) *RasterData {
	return &RasterData{tile: tile, pixels: pixels}
}

type RasterWriter struct {
	path         string
	tempFile     *os.File
	tempFileSize uint64
	dataSize     uint64
	zoom         uint8
	maxValue     float64
	tileSize     uint32

	tileOffsets    [][]uint32
	tileByteCounts [][]uint32
	uniformTiles   []map[uint32]int

	ifdPos            []int64
	nextIFDPos        []int64
	tileOffsetsPos    []int64
	tileByteCountsPos []int64
}

func NewRasterWriter(path string, zoom uint8, tileSize uint32) (*RasterWriter, error) {
	tempFile, err := os.CreateTemp("", "*.tmp")
	if err != nil {
		return nil, err
	}

	r := &RasterWriter{
		path:              path,
		tempFile:          tempFile,
		zoom:              zoom,
		tileSize:          tileSize,
		tileOffsets:       make([][]uint32, zoom+1),
		tileByteCounts:    make([][]uint32, zoom+1),
		uniformTiles:      make([]map[uint32]int, zoom+1),
		ifdPos:            make([]int64, zoom+1),
		nextIFDPos:        make([]int64, zoom+1),
		tileOffsetsPos:    make([]int64, zoom+1),
		tileByteCountsPos: make([]int64, zoom+1),
	}
	for z := uint8(0); z <= zoom; z++ {
		r.tileOffsets[z] = make([]uint32, 1<<(2*z))
		r.tileByteCounts[z] = make([]uint32, 1<<(2*z))
		r.uniformTiles[z] = make(map[uint32]int, 16)
	}
	return r, nil
}

func (w *RasterWriter) Write(r *RasterData) error {
	uniform := true
	color := uint32(r.pixels[0] + 0.5)
	for i := 0; i < len(r.pixels); i++ {
		col := r.pixels[i]
		if uint32(col+0.5) != color {
			uniform = false
			break
		}
		if col > w.maxValue {
			w.maxValue = r.pixels[i]
		}
	}
	if uniform {
		return w.WriteUniform(r.tile, color)
	}

	offset, size, err := w.compress(r.tile, r.pixels[:])
	if err != nil {
		return err
	}

	zoom, x, y := r.tile.ZoomXY()
	tileIndex := (1<<zoom)*y + x
	w.tileOffsets[zoom][tileIndex] = uint32(offset)
	w.tileByteCounts[zoom][tileIndex] = size

	return nil
}

func (w *RasterWriter) WriteUniform(tile TileKey, color uint32) error {
	zoom, x, y := tile.ZoomXY()
	tileIndex := (1<<zoom)*y + x
	if same, exists := w.uniformTiles[zoom][color]; exists {
		w.tileOffsets[zoom][tileIndex] = w.tileOffsets[zoom][same]
		w.tileByteCounts[zoom][tileIndex] = w.tileByteCounts[zoom][same]
		return nil
	}
	col := float64(color)
	if col > w.maxValue {
		w.maxValue = col
	}
	pixels := make([]float64, w.tileSize*w.tileSize)
	for i := 0; i < len(pixels); i++ {
		pixels[i] = col
	}
	offset, size, err := w.compress(tile, pixels[:])
	if err != nil {
		return err
	}
	w.tileOffsets[zoom][tileIndex] = uint32(offset)
	w.tileByteCounts[zoom][tileIndex] = size
	w.uniformTiles[zoom][color] = int(tileIndex)
	return nil
}

func (w *RasterWriter) compress(tile TileKey, pixels []float64) (offset uint64, size uint32, err error) {
	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return 0, 0, err
	}

	if err := binary.Write(writer, binary.LittleEndian, pixels); err != nil {
		return 0, 0, err
	}

	if err := writer.Close(); err != nil {
		return 0, 0, err
	}

	n, err := compressed.WriteTo(w.tempFile)
	if err != nil {
		return 0, 0, err
	}

	offset = w.tempFileSize
	w.tempFileSize += uint64(n)
	return offset, uint32(n), nil
}

func (w *RasterWriter) Close() error {
	out, err := os.Create(w.path + ".tmp")
	if err != nil {
		return err
	}
	if err := w.writeTiff(out); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Rename(w.path+".tmp", w.path); err != nil {
		return err
	}

	tempFileName := w.tempFile.Name()
	if err := w.tempFile.Close(); err != nil {
		return err
	}
	if err := os.Remove(tempFileName); err != nil {
		return err
	}

	return nil
}

func (w *RasterWriter) writeTiff(out *os.File) error {
	magic := []byte{'I', 'I', 42, 0}
	if _, err := out.Write(magic); err != nil {
		return err
	}

	if err := binary.Write(out, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}

	smd := `LAYOUT=IFDS_BEFORE_DATA
BLOCK_LEADER=SIZE_AS_UINT4
BLOCK_TRAILER=LAST_4_BYTES_REPEATED
KNOWN_INCOMPATIBLE_EDITION=NO 
`
	if !strings.Contains(smd, "=NO \n") {
		panic("missing space after NO") // as per GDAL documentation
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("GDAL_STRUCTURAL_METADATA_SIZE=%06d bytes\n", len(smd)))
	buf.WriteString(smd)
	if err := addPadding(&buf); err != nil {
		return err
	}

	if _, err := out.Write(buf.Bytes()); err != nil {
		return err
	}

	for zoom := int(w.zoom); zoom >= 0; zoom-- {
		if err := w.writeIFD(uint8(zoom), out); err != nil {
			return err
		}
	}

	if err := w.writeIFDList(out); err != nil {
		return err
	}

	for zoom := uint8(0); zoom <= w.zoom; zoom++ {
		if err := w.writeTiles(zoom, out); err != nil {
			return err
		}
	}

	for zoom := uint8(0); zoom <= w.zoom; zoom++ {
		if err := w.writeTileByteCounts(zoom, out); err != nil {
			return err
		}
	}

	return nil
}

func (w *RasterWriter) writeIFD(zoom uint8, f *os.File) error {
	const (
		newSubfileType   = 254
		imageWidth       = 256
		imageHeight      = 257
		bitsPerSample    = 258
		compression      = 259
		photometric      = 262
		imageDescription = 270
		samplesPerPixel  = 277
		planarConfig     = 284
		software         = 305
		tileWidth        = 322
		tileLength       = 323
		tileOffsets      = 324
		tileByteCounts   = 325
		sampleFormat     = 339
		sMinSampleValue  = 340
		sMaxSampleValue  = 341

		modelPixelScale = 33550
		modelTiepoint   = 33922
		geoKeyDirectory = 34735
		geoAsciiParams  = 34737

		asciiFormat  = 2
		shortFormat  = 3
		longFormat   = 4
		floatFormat  = 11
		doubleFormat = 12
	)

	fileSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	w.ifdPos[zoom] = fileSize

	geoAscii := "WGS 84 / Pseudo-Mercator|WGS 84|\u0000"
	geoKeys := []uint16{
		1, 1, 0, // Version: 1.1.0
		7,             // NumberOfKeys: 7
		1024, 0, 1, 1, // GTModelType: 2D projected
		1025, 0, 1, 1, // GTRasterTyp: PixelIsArea
		1026, geoAsciiParams, 25, 0, // GTCitation: "WGS 84 / Pseudo-Mercator"
		2049, geoAsciiParams, 7, 25, // GeodeticCitation: "WGS 84"
		2054, 0, 1, 9102, // GeogAngularUnits: degree [EPSG unit 9102]
		3072, 0, 1, 3857, // ProjectedCRS: Web Mercator [epsg.io/3857]
		3076, 0, 1, 9001, // ProjLinearUnits: meter [EPSG unit 9001]
	}

	// As per WGS 84, the circumference of the Earth at the equator is
	// defined to be 40075017 meters.
	const earthCircumference = 40075017.0
	metersPerPixel := earthCircumference / float64(uint64(1<<(w.zoom+8))) // at equator
	geoModelPixelScale := []float64{metersPerPixel, metersPerPixel, 0}
	geoModelTiepoints := []float64{0, 0, 0, -20037508.34, 20037508.34, 0}

	numTiles := uint32(1 << (zoom * 2))
	type ifdEntry struct {
		tag uint16
		val uint32
	}
	ifd := []ifdEntry{
		{imageWidth, 1 << (zoom + 8)},
		{imageHeight, 1 << (zoom + 8)},
		{bitsPerSample, 64},
		{compression, 8}, // 1 = no compression; 8 = zlib/flate
		{photometric, 0}, // 0 = WhiteIsZero
		{samplesPerPixel, 1},
		{planarConfig, 1},
		{tileWidth, w.tileSize},
		{tileLength, w.tileSize},
		{tileOffsets, 0},
		{tileByteCounts, 0},
		{sampleFormat, 3}, // 3 = IEEE floating point, TIFF spec page 80
	}

	if zoom == w.zoom {
		ifd = append(ifd, ifdEntry{imageDescription, 0})
		ifd = append(ifd, ifdEntry{software, 0})
		ifd = append(ifd, ifdEntry{modelPixelScale, 0})
		ifd = append(ifd, ifdEntry{modelTiepoint, 0})
		ifd = append(ifd, ifdEntry{geoKeyDirectory, 0})
		ifd = append(ifd, ifdEntry{geoAsciiParams, 0})
		ifd = append(ifd, ifdEntry{sMinSampleValue, 0})
		ifd = append(ifd, ifdEntry{sMaxSampleValue, 0})
	} else {
		ifd = append(ifd, ifdEntry{newSubfileType, 1})
	}

	sort.Slice(ifd, func(i, j int) bool { return ifd[i].tag < ifd[j].tag })

	extraPos := fileSize + int64(2+len(ifd)*12+4)

	var buf, extraBuf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, uint16(len(ifd))); err != nil {
		return err
	}

	lastTag := uint16(0)
	for i, e := range ifd {
		ifdEntryPos := fileSize + int64(2+i*12)

		if e.tag <= lastTag {
			panic("TIFF tags must be in increasing order")
		}
		lastTag = e.tag

		if err := binary.Write(&buf, binary.LittleEndian, e.tag); err != nil {
			return err
		}
		var typ uint16
		var count, value uint32
		switch e.tag {
		case newSubfileType:
			typ, count, value = longFormat, 1, e.val

		case imageDescription:
			s := []byte("OpenStreetMap view density, in weekly user views per km2\u0000")
			typ, count, value = asciiFormat, uint32(len(s)), uint32(extraPos)+uint32(extraBuf.Len())
			if _, err := extraBuf.Write(s); err != nil {
				return err
			}

		case software:
			s := []byte("TileRank\u0000")
			typ, count, value = asciiFormat, uint32(len(s)), uint32(extraPos)+uint32(extraBuf.Len())
			extraBuf.Write(s)

		case sMinSampleValue:
			typ, count, value = floatFormat, 1, math.Float32bits(0)

		case sMaxSampleValue:
			typ, count, value = floatFormat, 1, math.Float32bits(float32(w.maxValue))

		case geoKeyDirectory:
			typ, count, value = shortFormat, uint32(len(geoKeys)), uint32(extraPos)+uint32(extraBuf.Len())
			if err := binary.Write(&extraBuf, binary.LittleEndian, geoKeys); err != nil {
				return err
			}

		case modelPixelScale:
			typ, count, value = doubleFormat, uint32(len(geoModelPixelScale)), uint32(extraPos)+uint32(extraBuf.Len())
			if err := binary.Write(&extraBuf, binary.LittleEndian, geoModelPixelScale); err != nil {
				return err
			}

		case modelTiepoint:
			typ, count, value = doubleFormat, uint32(len(geoModelTiepoints)), uint32(extraPos)+uint32(extraBuf.Len())
			if err := binary.Write(&extraBuf, binary.LittleEndian, geoModelTiepoints); err != nil {
				return err
			}

		case geoAsciiParams:
			s := []byte(geoAscii)
			typ, count, value = asciiFormat, uint32(len(s)), uint32(extraPos)+uint32(extraBuf.Len())
			if _, err := extraBuf.Write(s); err != nil {
				return err
			}

		case tileOffsets:
			typ, count, value = longFormat, numTiles, 0xdeadbeef
			w.tileOffsetsPos[zoom] = ifdEntryPos + 8

		case tileByteCounts:
			typ, count, value = longFormat, numTiles, 0xdeadbeef
			w.tileByteCountsPos[zoom] = ifdEntryPos + 8

		default:
			typ, count, value = longFormat, uint32(1), e.val
			if e.val <= 0xffff {
				typ = shortFormat
			}
		}
		if err := binary.Write(&buf, binary.LittleEndian, typ); err != nil {
			return err
		}
		if err := binary.Write(&buf, binary.LittleEndian, count); err != nil {
			return err
		}
		if err := binary.Write(&buf, binary.LittleEndian, value); err != nil {
			return err
		}
	}

	nextIFD := uint32(0)
	nextIFDPos := fileSize + int64(buf.Len())
	w.nextIFDPos[zoom] = nextIFDPos
	if err := binary.Write(&buf, binary.LittleEndian, nextIFD); err != nil {
		return err
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		return err
	}
	fileSize += int64(buf.Len())

	if err := addPadding(&extraBuf); err != nil {
		return err
	}
	if fileSize != extraPos {
		panic("fileSize != extraPos")
	}

	fileSize += int64(extraBuf.Len())
	if _, err := extraBuf.WriteTo(f); err != nil {
		return err
	}

	return nil
}

func (w *RasterWriter) writeIFDList(f io.WriteSeeker) error {
	pos := int64(4)
	for zoom := int(w.zoom); zoom >= 0; zoom-- {
		if w.ifdPos[zoom] != 0 {
			if err := patchOffset(f, pos, w.ifdPos[zoom]); err != nil {
				return err
			}
			pos = w.nextIFDPos[zoom]
		}
	}
	if err := patchOffset(f, pos, 0); err != nil {
		return err
	}
	return nil
}

func (w *RasterWriter) writeTiles(zoom uint8, f *os.File) error {
	if w.tileByteCountsPos[zoom] == 0 {
		return nil
	}

	fileSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	if fileSize&3 != 0 {
		padding := []byte{0, 0, 0}[fileSize&3-1:]
		if n, err := f.Write(padding); err == nil {
			fileSize += int64(n)
		} else {
			return err
		}
	}

	numTiles := uint32(1 << (zoom * 2))

	tileOffsetsPos := fileSize
	numRows := 1 << zoom
	emptyRow := make([]byte, numRows*4)
	for y := 0; y < numRows; y++ {
		if _, err := f.Write(emptyRow); err != nil {
			return err
		}
	}
	fileSize += int64(numTiles * 4)

	uniform := make(map[uint32]bool, len(w.uniformTiles[zoom]))
	uniformPos := make(map[uint32]uint32, len(w.uniformTiles[zoom]))
	for _, t := range w.uniformTiles[zoom] {
		uniform[w.tileOffsets[zoom][t]] = true
	}

	finalTileOffsets := make([]uint32, numTiles)
	for tile := uint32(0); tile < numTiles; tile++ {
		tileOffset := w.tileOffsets[zoom][tile]
		if unipos, exists := uniformPos[tileOffset]; !exists {
			tileSize := w.tileByteCounts[zoom][tile]
			data := make([]byte, tileSize+8)
			payload := data[4 : 4+tileSize]
			if _, err := w.tempFile.ReadAt(payload, int64(tileOffset)); err != nil {
				return err
			}

			var leader bytes.Buffer
			if err := binary.Write(&leader, binary.LittleEndian, uint32(tileSize)); err != nil {
				return err
			}
			copy(data[0:4], leader.Bytes())

			copy(data[len(data)-4:], payload[len(payload)-4:])

			finalTileOffset := uint32(fileSize) + 4
			finalTileOffsets[tile] = finalTileOffset
			if uniform[tileOffset] {
				uniformPos[tileOffset] = finalTileOffset
			}
			if _, err := f.Write(data); err != nil {
				return err
			}
			fileSize += int64(len(data))
		} else {
			finalTileOffsets[tile] = unipos
		}
	}

	if len(finalTileOffsets) == 1 {
		if _, err := f.Seek(w.tileOffsetsPos[zoom], io.SeekStart); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, finalTileOffsets[0]); err != nil {
			return err
		}

		return nil
	}

	if _, err := f.Seek(tileOffsetsPos, io.SeekStart); err != nil {
		return err
	}

	if err := binary.Write(f, binary.LittleEndian, finalTileOffsets); err != nil {
		return err
	}

	if err := patchOffset(f, w.tileOffsetsPos[zoom], tileOffsetsPos); err != nil {
		return err
	}

	return nil
}

func (w *RasterWriter) writeTileByteCounts(zoom uint8, f io.WriteSeeker) error {
	pos, sizes := w.tileByteCountsPos[zoom], w.tileByteCounts[zoom]

	if pos == 0 {
		return nil
	}

	if len(sizes) == 1 {
		if _, err := f.Seek(pos, io.SeekStart); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, sizes); err != nil {
			return err
		}
		return nil
	}

	arrayPos, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	if arrayPos&3 != 0 {
		padding := []byte{0, 0, 0}[arrayPos&3-1:]
		if n, err := f.Write(padding); err == nil {
			arrayPos += int64(n)
		} else {
			return err
		}
	}

	if err := binary.Write(f, binary.LittleEndian, sizes); err != nil {
		return err
	}

	if err := patchOffset(f, pos, arrayPos); err != nil {
		return err
	}

	return nil
}

func addPadding(buf *bytes.Buffer) error {
	if buf.Len()&1 != 0 {
		if err := buf.WriteByte(0); err != nil {
			return err
		}
	}
	return nil
}

func patchOffset(f io.WriteSeeker, pos int64, value int64) error {
	if value < 0 || value > 0xffffffff {
		panic("offset value out of range")
	}

	if _, err := f.Seek(pos, io.SeekStart); err != nil {
		return err
	}

	if err := binary.Write(f, binary.LittleEndian, uint32(value)); err != nil {
		return err
	}

	return nil
}
