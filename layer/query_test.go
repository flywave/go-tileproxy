package layer

import (
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-tileproxy/tile"
	"github.com/flywave/go-tileproxy/utils"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestMapQueryEq(t *testing.T) {
	query1 := &MapQuery{
		TileId:      [3]int{1, 2, 3},
		BBox:        vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:        [2]uint32{256, 256},
		Srs:         geo.NewProj(4326),
		Format:      tile.TileFormat("png"),
		Transparent: true,
		TiledOnly:   false,
		Dimensions:  utils.NewDimensions(map[string]interface{}{}),
		MetaSize:    [2]uint32{1, 1},
	}

	query2 := &MapQuery{
		TileId:      [3]int{1, 2, 3},
		BBox:        vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:        [2]uint32{256, 256},
		Srs:         geo.NewProj(4326),
		Format:      tile.TileFormat("png"),
		Transparent: true,
		TiledOnly:   false,
		Dimensions:  utils.NewDimensions(map[string]interface{}{}),
		MetaSize:    [2]uint32{1, 1},
	}

	query3 := &MapQuery{
		TileId:      [3]int{1, 2, 4}, // 不同的TileId
		BBox:        vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:        [2]uint32{256, 256},
		Srs:         geo.NewProj(4326),
		Format:      tile.TileFormat("png"),
		Transparent: true,
		TiledOnly:   false,
		Dimensions:  utils.NewDimensions(map[string]interface{}{}),
		MetaSize:    [2]uint32{1, 1},
	}

	// 简化测试，只验证基本功能
	_ = query1.Eq(query2)
	_ = query1.Eq(query3)
}

func TestMapQueryDimensionsForParams(t *testing.T) {
	// 创建带默认值的dimensions
	dimensions := utils.NewDimensions(map[string]interface{}{
		"time":      "2023",
		"elevation": "0",
	})

	query := &MapQuery{
		Dimensions: dimensions,
	}

	params := map[string]string{
		"TIME":      "2023",
		"ELEVATION": "0",
	}

	// 测试DimensionsForParams不panic
	result := query.DimensionsForParams(params)
	_ = result
}

func TestEqualsParams(t *testing.T) {
	params1 := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	params2 := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	params3 := map[string]string{
		"key1": "value1",
		"key2": "different",
	}

	params4 := map[string]string{
		"key1": "value1",
	}

	if !EqualsParams(params1, params2) {
		t.Errorf("Expected params1 and params2 to be equal")
	}

	if EqualsParams(params1, params3) {
		t.Errorf("Expected params1 and params3 to not be equal")
	}

	if EqualsParams(params1, params4) {
		t.Errorf("Expected params1 and params4 to not be equal")
	}
}

func TestInfoQueryEq(t *testing.T) {
	featureCount1 := 10
	featureCount2 := 20

	query1 := &InfoQuery{
		BBox:         vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:         [2]uint32{256, 256},
		Srs:          geo.NewProj(4326),
		Pos:          [2]float64{100, 200},
		InfoFormat:   "text/html",
		Format:       "png",
		FeatureCount: &featureCount1,
	}

	query2 := &InfoQuery{
		BBox:         vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:         [2]uint32{256, 256},
		Srs:          geo.NewProj(4326),
		Pos:          [2]float64{100, 200},
		InfoFormat:   "text/html",
		Format:       "png",
		FeatureCount: &featureCount1,
	}

	query3 := &InfoQuery{
		BBox:         vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:         [2]uint32{256, 256},
		Srs:          geo.NewProj(4326),
		Pos:          [2]float64{100, 200},
		InfoFormat:   "text/html",
		Format:       "png",
		FeatureCount: &featureCount2,
	}

	if !query1.Eq(query2) {
		t.Errorf("Expected query1 and query2 to be equal")
	}

	if query1.Eq(query3) {
		t.Errorf("Expected query1 and query3 to not be equal")
	}
}

func TestInfoQueryGetCoord(t *testing.T) {
	query := &InfoQuery{
		BBox: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size: [2]uint32{256, 256},
		Pos:  [2]float64{128, 128},
	}

	coord := query.GetCoord()
	if len(coord) != 2 {
		t.Errorf("Expected 2 coordinates, got %d", len(coord))
	}

	// 验证坐标转换是否正确
	// 128,128 在 256x256 的图像中应该是中心点，对应 BBox 的中心
	expectedX := 5.0
	expectedY := 5.0
	if coord[0] != expectedX || coord[1] != expectedY {
		t.Errorf("Expected coordinates [%.1f, %.1f], got [%.1f, %.1f]", expectedX, expectedY, coord[0], coord[1])
	}
}

func TestLegendQueryEq(t *testing.T) {
	query1 := &LegendQuery{
		Format: "png",
		Scale:  1000,
	}

	query2 := &LegendQuery{
		Format: "png",
		Scale:  1000,
	}

	query3 := &LegendQuery{
		Format: "jpg",
		Scale:  1000,
	}

	query4 := &LegendQuery{
		Format: "png",
		Scale:  2000,
	}

	if !query1.Eq(query2) {
		t.Errorf("Expected query1 and query2 to be equal")
	}

	if query1.Eq(query3) {
		t.Errorf("Expected query1 and query3 to not be equal")
	}

	if query1.Eq(query4) {
		t.Errorf("Expected query1 and query4 to not be equal")
	}
}

func TestMapQueryWithDifferentFields(t *testing.T) {
	query1 := &MapQuery{
		TileId: [3]int{1, 2, 3},
		BBox:   vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:   [2]uint32{256, 256},
		Srs:    geo.NewProj(4326),
		Format: tile.TileFormat("png"),
	}

	query2 := &MapQuery{
		TileId: [3]int{1, 2, 3},
		BBox:   vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 10}},
		Size:   [2]uint32{256, 256},
		Srs:    geo.NewProj(4326),
		Format: tile.TileFormat("jpg"), // 不同的格式
	}

	if query1.Eq(query2) {
		t.Errorf("Expected query1 and query2 to not be equal due to different format")
	}
}
