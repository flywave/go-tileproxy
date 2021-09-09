package ray

import (
	"encoding/json"
	"fmt"
	"testing"

	vec3d "github.com/flywave/go3d/float64/vec3"
	"github.com/stretchr/testify/assert"
)

func assertEqualVectors(t *testing.T, expected vec3d.T, v vec3d.T) {
	assert := assert.New(t)
	assert.InDelta(expected[0], v[0], Epsilon)
	assert.InDelta(expected[1], v[1], Epsilon)
	assert.InDelta(expected[2], v[2], Epsilon)
}

func TestGetBoundingBox(t *testing.T) {
	meshData := []byte(`{
		"vertices": [
			{
				"normal": [0, 0, 1],
				"coordinates": [-1, -2, 0]
			},
			{
				"normal": [0, 0, 1],
				"coordinates": [1, 5, -8]
			},
			{
				"normal": [0, 0, 1],
				"coordinates": [1, 2, 0]
			},
			{
				"normal": [0, 0, 1],
				"coordinates": [1, 1, 0]
			}
		],
		"faces": [
			{
				"vertices": [0, 1, 2],
				"material": 42
			},
			{
				"vertices": [1, 2, 3],
				"material": 42
			}
		]
	}`)
	mesh := &Mesh{}
	err := json.Unmarshal(meshData, &mesh)
	if err != nil {
		t.Fatalf("Error reading json: %s\n", err)
		return
	}

	mesh.Init()

	bbox := mesh.GetBoundingBox()

	assertEqualVectors(t, vec3d.T{1, 5, 0}, vec3d.T(bbox.MaxVolume))
	assertEqualVectors(t, vec3d.T{-1, -2, -8}, vec3d.T(bbox.MinVolume))
}

func ExampleMesh_Intersect() {
	meshData := []byte(`{
		"vertices": [
			{
				"normal": [0, 0, 1],
				"coordinates": [0, 0, 0],
				"uv": [0, 0]
			},
			{
				"normal": [0, 0, 1],
				"coordinates": [1, 0, 0],
				"uv": [1, 0]
			},
			{
				"normal": [0, 0, 1],
				"coordinates": [0, 1, 0],
				"uv": [0, 1]
			}
		],
		"faces": [
			{
				"vertices": [0, 1, 2],
				"material": 42
			}
		]
	}`)
	mesh := &Mesh{}
	err := json.Unmarshal(meshData, &mesh)
	if err != nil {
		fmt.Printf("Error reading json: %s\n", err)
		return
	}

	mesh.Init()

	testRay := &Ray{
		Start:     vec3d.T{0.15, 0.11, 1},
		Direction: vec3d.T{0, 0, -1},
	}

	var (
		intersection *Intersection
	)
	intersection = mesh.Intersect(testRay)

	if intersection == nil {
		fmt.Printf("no intersection\n")
	} else {
		fmt.Printf("intersection point: %s\n", intersection.Point)
		fmt.Printf("caused by ray: %s\n", intersection.Incoming)
		fmt.Printf("at a distance: %.3g\n", intersection.Distance)
		fmt.Printf("with surface coordinates: (%.3g, %.3g)\n",
			intersection.U, intersection.V)
		fmt.Printf("surface normal: %s\n", intersection.Normal)
		fmt.Printf("surface coordinate system: Ox: %s, Oy: %s\n",
			intersection.SurfaceOx, intersection.SurfaceOy)
		fmt.Printf("surface material: %d\n", intersection.Material)
	}
}

func TestIntersectTwoTriangles(t *testing.T) {
	meshData := []byte(`{
		"vertices": [
			{
				"normal": [0, 0, 1],
				"coordinates": [0, 0, 0],
				"uv": [0, 0]
			},
			{
				"normal": [0, 0, 1],
				"coordinates": [1, 0, 0],
				"uv": [1, 0]
			},
			{
				"normal": [0, 0, 1],
				"coordinates": [0, 1, 0],
				"uv": [0, 1]
			},
			{
				"normal": [0, 0, -1],
				"coordinates": [0, 0, 4],
				"uv": [0, 0]
			},
			{
				"normal": [0, 0, -1],
				"coordinates": [1, 0, 4],
				"uv": [1, 0]
			},
			{
				"normal": [0, 0, -1],
				"coordinates": [0, 1, 4],
				"uv": [0, 1]
			}
		],
		"faces": [
			{
				"vertices": [3, 4, 5],
				"material": 5
			},
			{
				"vertices": [0, 1, 2],
				"material": 42
			}
		]
	}`)
	mesh := &Mesh{}
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
	if intersection.Material != 42 {
		t.Error("Intersected wrong triangle")
	}

	testRay.Direction[2] = 1

	intersection = mesh.Intersect(testRay)
	if intersection == nil {
		t.Fatal("Intersection shouldn't be nil")
	}
	if intersection.Material != 5 {
		t.Error("Intersected wrong triangle")
	}

}
