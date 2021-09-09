package ray

import (
	"encoding/json"
	"testing"
	"unsafe"

	vec3d "github.com/flywave/go3d/float64/vec3"
	"github.com/stretchr/testify/assert"
)

func assertEqualVectors(t *testing.T, expected vec3d.T, v vec3d.T) {
	assert := assert.New(t)
	assert.InDelta(expected[0], v[0], Epsilon)
	assert.InDelta(expected[1], v[1], Epsilon)
	assert.InDelta(expected[2], v[2], Epsilon)
}

func TestGetBBox(t *testing.T) {
	meshData := []byte(`{
		"vertices": [
			[-1, -2, 0],
			[1, 5, -8],
			[1, 2, 0],
			[1, 1, 0]
		],
		"faces": [
			[0, 1, 2],
			[1, 2, 3]
		]
	}`)
	mesh := &RayMesh{}
	err := json.Unmarshal(meshData, &mesh)
	if err != nil {
		t.Fatalf("Error reading json: %s\n", err)
		return
	}

	mesh.Init()

	bbox := mesh.GetBBox()

	assertEqualVectors(t, vec3d.T{1, 5, 0}, vec3d.T(bbox.MaxVolume))
	assertEqualVectors(t, vec3d.T{-1, -2, -8}, vec3d.T(bbox.MinVolume))
}

func TestIntersectTwoTriangles(t *testing.T) {
	meshData := []byte(`{
		"vertices": [
			[0, 0, 0],
			[1, 0, 0],
			[0, 1, 0],
			[0, 0, 4],
			[1, 0, 4],
			[0, 1, 4]
		],
		"faces": [
			[3, 4, 5],
			[0, 1, 2]
		]
	}`)
	mesh := &RayMesh{}
	err := json.Unmarshal(meshData, &mesh)
	if err != nil {
		t.Fatalf("Error reading json: %s\n", err)
	}

	mesh.Init()

	testRay := &Ray{
		Start:     vec3d.T{0.15, 0.11, 1},
		Direction: vec3d.T{0, 0, -1},
	}

	var intersection *Intersection
	intersection = mesh.Intersect(testRay)
	if intersection == nil {
		t.Fatal("Intersection shouldn't be nil")
	}

	testRay.Direction[2] = 1

	intersection = mesh.Intersect(testRay)
	if intersection == nil {
		t.Fatal("Intersection shouldn't be nil")
	}
}

func TestConvert(t *testing.T) {
	vertices := [][3]float64{{0, 0, 0}}
	vvecs := *(*[]vec3d.T)(unsafe.Pointer(&vertices))

	if vvecs != nil {
		t.FailNow()
	}
}
