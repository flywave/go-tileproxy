package terrain

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"unsafe"

	qmt "github.com/flywave/go-quantized-mesh"
	tin "github.com/flywave/go-tin"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geoid"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
	"github.com/flywave/go-tileproxy/utils/ray"
)

type TerrainSource struct {
	tile.Source
	data      *qmt.QuantizedMeshTile
	buf       []byte
	fname     string
	size      [2]uint32
	cacheable *tile.CacheInfo
	georef    *geo.GeoReference
	Options   tile.TileOptions
}

func NewTerrainSource(options tile.TileOptions) *TerrainSource {
	src := &TerrainSource{Options: options}
	return src
}

func (s *TerrainSource) GetType() tile.TileType {
	return tile.TILE_DEM
}

func (s *TerrainSource) GetCacheable() *tile.CacheInfo {
	if s.cacheable == nil {
		s.cacheable = &tile.CacheInfo{Cacheable: false}
	}
	return s.cacheable
}

func (s *TerrainSource) SetCacheable(c *tile.CacheInfo) {
	s.cacheable = c
}

func (s *TerrainSource) GetFileName() string {
	return s.fname
}

func (s *TerrainSource) GetTile() interface{} {
	return s.GetQuantizedMeshTile()
}

func (s *TerrainSource) GetRasterOptions() *RasterOptions {
	return s.Options.(*RasterOptions)
}

func (s *TerrainSource) Decode(r io.Reader) (*qmt.QuantizedMeshTile, error) {
	if reader, ok := r.(io.ReadSeeker); !ok {
		return nil, errors.New("reader is must ReadSeeker")
	} else {
		mesh := &qmt.QuantizedMeshTile{}
		err := mesh.Read(reader, qmt.Ext_None)
		return mesh, err
	}
}

func (s *TerrainSource) Encode(tile *qmt.QuantizedMeshTile) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := tile.Write(buf)
	return buf.Bytes(), err
}

func (s *TerrainSource) GetQuantizedMeshTile() *qmt.QuantizedMeshTile {
	if s.data == nil {
		if s.buf == nil {
			f, err := os.Open(s.fname)
			if err != nil {
				return nil
			}
			s.buf, err = ioutil.ReadAll(f)
			if err != nil {
				return nil
			}
		}
		r := utils.NewMemFile(s.buf)
		var err error
		s.data, err = s.Decode(r)
		if err != nil {
			return nil
		}
	}
	return s.data
}

func (s *TerrainSource) GetGeoReference() *geo.GeoReference {
	return s.georef
}

func (s *TerrainSource) GetSize() [2]uint32 {
	return s.size
}

func (s *TerrainSource) GetSource() interface{} {
	if s.data != nil {
		return s.data
	} else if len(s.fname) > 0 {
		return s.fname
	}
	return nil
}

func (s *TerrainSource) SetSource(src interface{}) {
	s.data = nil
	s.buf = nil
	switch ss := src.(type) {
	case io.ReadSeeker:
		s.data, _ = s.Decode(ss)
	case string:
		s.fname = ss
	case *qmt.QuantizedMeshTile:
		s.data = ss
	}
}

func (s *TerrainSource) GetBuffer(format *tile.TileFormat, in_tile_opts tile.TileOptions) []byte {
	if s.buf == nil {
		var err error
		s.buf, err = s.Encode(s.data)
		if err != nil {
			return nil
		}
	}
	return s.buf
}

func (s *TerrainSource) SetTileOptions(options tile.TileOptions) {
	s.Options = options
}

func (s *TerrainSource) GetTileOptions() tile.TileOptions {
	return s.Options
}

func GenTerrainSource(data *TileData, options *RasterOptions) (*TerrainSource, error) {
	if !data.IsUnilateral() {
		return nil, errors.New("error")
	}

	if data.NoDataValue() == 0 {
		data.NoData = -9999
	}

	raw, si, tsf := data.GetExtend()
	xsize, ysize := int(si[0]), int(si[0])

	ypos := tsf[3] + tsf[5]*float64(ysize)

	rd := tin.NewRasterDoubleWithData(xsize, ysize, raw)
	rd.NoData = data.NoDataValue()

	rd.SetTransform(func(v *tin.Vertex) tin.Vertex {
		z := v[2]
		if options.HeightModel == geoid.HAE {
			z += options.HeightOffset
		} else {
			gid := geoid.NewGeoid(options.HeightModel, false)
			z = gid.ConvertHeight(v[0], v[1], z, geoid.GEOIDTOELLIPSOID)
		}
		return tin.Vertex{v[0], v[1], z}
	})

	rd.SetXYPos(tsf[0], ypos, tsf[1])

	mesh1 := tin.GenerateTinMesh(rd, options.MaxError)
	mk := tin.NewTileMaker(mesh1)

	mesh, _ := mk.GenTile()

	qdt := &qmt.MeshData{}
	qdt.Vertices = *(*[][3]float64)(unsafe.Pointer(&mesh.Vertices))
	qdt.Faces = *(*[][3]int)(unsafe.Pointer(&mesh.Faces))
	qdt.BBox = mesh.BBox

	qmesh := &qmt.QuantizedMeshTile{}
	qmesh.SetMesh(qdt, false)

	source := NewTerrainSource(options)

	source.SetSource(qmesh)

	return source, nil
}

func newRayMesh(qmesh *qmt.QuantizedMeshTile) *ray.RayMesh {
	mdata, err := qmesh.GetMesh()
	if err != nil {
		return nil
	}
	return ray.NewRayMesh(mdata.Vertices, mdata.Faces, mdata.BBox)
}

func (s *TerrainSource) Resample(georef *geo.GeoReference, grid *Grid) error {
	if georef == nil && s.georef != nil {
		georef = s.georef
	}
	if georef == nil {
		return errors.New("source georef is nil")
	}
	bbox := grid.GetRect()
	if !grid.srs.Eq(georef.GetSrs()) {
		bbox = grid.srs.TransformRectTo(georef.GetSrs(), bbox, 16)
	}
	if !geo.BBoxContains(georef.GetBBox(), bbox) {
		return errors.New("not Contains target grid")
	}

	rayMesh := newRayMesh(s.GetQuantizedMeshTile())

	if rayMesh == nil {
		return errors.New("ray mesh error")
	}

	rays := grid.GetRay()

	grid.Coordinates = make(Coordinates, len(rays))

	for i, ray := range rays {
		lon, lat := ray.Start[0], ray.Start[1]
		if !grid.srs.Eq(georef.GetSrs()) {
			d := grid.srs.TransformTo(georef.GetSrs(), []vec2d.T{{lon, lat}})
			lon, lat = d[0][0], d[0][1]
		}
		intersection := rayMesh.Intersect(&ray)
		grid.Coordinates[i][0] = lon
		grid.Coordinates[i][1] = lat
		grid.Coordinates[i][2] = intersection.Point[2]
	}
	return nil
}
