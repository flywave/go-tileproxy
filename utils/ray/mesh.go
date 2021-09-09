package ray

import (
	"math"

	vec3d "github.com/flywave/go3d/float64/vec3"
)

func SolveEquation(a, b, c vec3d.T) (float64, float64) {
	coefficientMatrix := [2][2]float64{{a[0], b[0]}, {a[1], b[1]}}
	constantCoefficientMatrix := [2]float64{c[0], c[1]}

	det := coefficientMatrix[0][0]*coefficientMatrix[1][1] -
		coefficientMatrix[1][0]*coefficientMatrix[0][1]
	x := (constantCoefficientMatrix[0]*coefficientMatrix[1][1] -
		constantCoefficientMatrix[1]*coefficientMatrix[0][1]) / det
	y := (constantCoefficientMatrix[1]*coefficientMatrix[0][0] -
		constantCoefficientMatrix[0]*coefficientMatrix[1][0]) / det
	return x, y
}

func Negative(v *vec3d.T) vec3d.T {
	return v.Scaled(-1)
}

func MixedProduct(a, b, c *vec3d.T) float64 {
	ab := vec3d.Cross(a, b)
	return vec3d.Dot(&ab, c)
}

const (
	MaxTreeDepth     = 10
	TrianglesPerLeaf = 20
)

type Vertex struct {
	Normal      vec3d.T `json:"normal"`
	Coordinates vec3d.T `json:"coordinates"`
	UV          vec3d.T `json:"uv"`
}

type Triangle struct {
	Vertices      [3]int   `json:"vertices"`
	Material      int      `json:"material"`
	AB, AC, ABxAC *vec3d.T `json:"-"`
	Normal        *vec3d.T `json:"normal"`
	surfaceOx     *vec3d.T `json:"-"`
	surfaceOy     *vec3d.T `json:"-"`
}

type Mesh struct {
	Vertices    []Vertex     `json:"vertices"`
	Faces       []Triangle   `json:"faces"`
	tree        *KDtree      `json:"-"`
	BoundingBox *BoundingBox `json:"-"`
}

func (m *Mesh) Init() {
	allIndices := make([]int, len(m.Faces))
	for i := range allIndices {
		allIndices[i] = i
	}

	m.BoundingBox = m.GetBoundingBox()
	m.tree = m.newKDtree(m.BoundingBox, allIndices, 0)
	for i := range m.Faces {
		triangle := &m.Faces[i]

		A := &m.Vertices[triangle.Vertices[0]].Coordinates
		B := &m.Vertices[triangle.Vertices[1]].Coordinates
		C := &m.Vertices[triangle.Vertices[2]].Coordinates

		AB := vec3d.Sub(B, A)
		AC := vec3d.Sub(C, A)

		triangle.AB = &AB
		triangle.AC = &AC
		c := vec3d.Cross(&AB, &AC)
		triangle.ABxAC = &c

		surfaceA := &m.Vertices[triangle.Vertices[0]].UV
		surfaceB := &m.Vertices[triangle.Vertices[1]].UV
		surfaceC := &m.Vertices[triangle.Vertices[2]].UV

		surfaceAB := vec3d.Sub(surfaceB, surfaceA)
		surfaceAC := vec3d.Sub(surfaceC, surfaceA)

		px, qx := SolveEquation(surfaceAB, surfaceAC, vec3d.T{1, 0, 0})
		py, qy := SolveEquation(surfaceAB, surfaceAC, vec3d.T{0, 1, 0})
		aas, acs := AB.Scaled(px), AC.Scaled(qx)
		ox := vec3d.Add(&aas, &acs)
		triangle.surfaceOx = &ox
		aas, acs = AB.Scaled(py), AC.Scaled(qy)
		oy := vec3d.Add(&aas, &acs)
		triangle.surfaceOy = &oy
	}
}

func (m *Mesh) SlowIntersect(incoming *Ray) *Intersection {
	intersection := &Intersection{}
	intersection.Distance = Inf
	found := false
	for _, triangle := range m.Faces {
		if m.intersectTriangle(incoming, &triangle, intersection, nil) {
			found = true
		}
	}
	if !found {
		return nil
	}
	return intersection
}

func (m *Mesh) Intersect(incoming *Ray) *Intersection {
	incoming.Init()
	if !m.BoundingBox.Intersect(incoming) {
		return nil
	}
	intersectionInfo := &Intersection{Distance: Inf}
	if m.IntersectKD(incoming, m.BoundingBox, m.tree, intersectionInfo) {
		return intersectionInfo
	}
	return nil
}

func IntersectTriangle(ray *Ray, A, B, C *vec3d.T) (bool, float64) {
	AB := vec3d.Sub(B, A)
	AC := vec3d.Sub(C, A)
	reverseDirection := Negative(&ray.Direction)
	distToA := vec3d.Sub(&ray.Start, A)
	ABxAC := vec3d.Cross(&AB, &AC)
	det := vec3d.Dot(&ABxAC, &reverseDirection)
	reverseDet := 1 / det
	if math.Abs(det) < Epsilon {
		return false, Inf
	}
	lambda2 := MixedProduct(&distToA, &AC, &reverseDirection) * reverseDet
	lambda3 := MixedProduct(&AB, &distToA, &reverseDirection) * reverseDet
	gamma := vec3d.Dot(&ABxAC, &distToA) * reverseDet
	if gamma < 0 {
		return false, Inf
	}
	if lambda2 < 0 || lambda2 > 1 || lambda3 < 0 || lambda3 > 1 || lambda2+lambda3 > 1 {
		return false, Inf
	}
	return true, gamma
}

func (m *Mesh) intersectTriangle(ray *Ray, triangle *Triangle, intersection *Intersection, boundingBox *BoundingBox) bool {
	A := &m.Vertices[triangle.Vertices[0]].Coordinates
	distToA := vec3d.Sub(&ray.Start, A)
	rayDir := ray.Direction
	ABxAC := triangle.ABxAC
	det := -vec3d.Dot(ABxAC, &rayDir)
	if math.Abs(det) < Epsilon {
		return false
	}
	reverseDet := 1 / det
	intersectDist := vec3d.Dot(ABxAC, &distToA) * reverseDet
	if intersectDist < 0 || intersectDist > intersection.Distance {
		return false
	}
	lambda2 := MixedProduct(&distToA, &rayDir, triangle.AC) * reverseDet
	lambda3 := -MixedProduct(&distToA, &rayDir, triangle.AB) * reverseDet
	if lambda2 < 0 || lambda2 > 1 || lambda3 < 0 || lambda3 > 1 || lambda2+lambda3 > 1 {
		return false
	}

	sc := rayDir.Scaled(intersectDist)
	ip := vec3d.Add(&ray.Start, &sc)

	if boundingBox != nil && !boundingBox.Inside(ip) {
		return false
	}
	intersection.Point = &ip
	intersection.Distance = intersectDist
	if triangle.Normal != nil {
		intersection.Normal = triangle.Normal
	} else {
		Anormal := &m.Vertices[triangle.Vertices[0]].Normal
		Bnormal := &m.Vertices[triangle.Vertices[1]].Normal
		Cnormal := &m.Vertices[triangle.Vertices[2]].Normal
		bas := vec3d.Sub(Bnormal, Anormal)
		ABxlambda2 := bas.Scaled(lambda2)
		cas := vec3d.Sub(Cnormal, Anormal)
		ACxlambda3 := cas.Scaled(lambda3)
		abc := vec3d.Add(&ABxlambda2, &ACxlambda3)
		normal := vec3d.Add(Anormal, &abc)
		intersection.Normal = &normal
	}
	uvA := &m.Vertices[triangle.Vertices[0]].UV
	uvB := &m.Vertices[triangle.Vertices[1]].UV
	uvC := &m.Vertices[triangle.Vertices[2]].UV

	uvba := vec3d.Sub(uvB, uvA)
	uvABxlambda2 := uvba.Scaled(lambda2)
	uvca := vec3d.Sub(uvC, uvA)
	uvACxlambda3 := uvca.Scaled(lambda3)

	uvabbc := vec3d.Add(&uvABxlambda2, &uvACxlambda3)
	uv := vec3d.Add(uvA, &uvabbc)

	intersection.U = uv[0]
	intersection.V = uv[1]

	intersection.SurfaceOx = triangle.surfaceOx
	intersection.SurfaceOy = triangle.surfaceOy

	intersection.Incoming = ray
	intersection.Material = triangle.Material
	return true
}

func (m *Mesh) GetBoundingBox() *BoundingBox {
	boundingBox := NewBoundingBox()
	for _, vertex := range m.Vertices {
		boundingBox.AddPoint(vertex.Coordinates)
	}
	return boundingBox
}

func (m *Mesh) newKDtree(boundingBox *BoundingBox, trianglesIndices []int, depth int) *KDtree {
	if depth > MaxTreeDepth || len(trianglesIndices) < TrianglesPerLeaf {
		node := NewLeaf(trianglesIndices)
		return node
	}
	axis := (depth + 2) % 3
	leftLimit := boundingBox.MaxVolume[axis]
	righLimit := boundingBox.MinVolume[axis]

	median := (leftLimit + righLimit) / 2

	var leftTriangles, rightTriangles []int
	var A, B, C *vec3d.T
	leftBoundingBox, rightBoundingBox := boundingBox.Split(axis, median)
	for _, index := range trianglesIndices {
		A = &m.Vertices[m.Faces[index].Vertices[0]].Coordinates
		B = &m.Vertices[m.Faces[index].Vertices[1]].Coordinates
		C = &m.Vertices[m.Faces[index].Vertices[2]].Coordinates

		if leftBoundingBox.IntersectTriangle(*A, *B, *C) {
			leftTriangles = append(leftTriangles, index)
		}

		if rightBoundingBox.IntersectTriangle(*A, *B, *C) {
			rightTriangles = append(rightTriangles, index)
		}
	}
	node := NewNode(median, axis)
	leftChild := m.newKDtree(leftBoundingBox, leftTriangles, depth+1)
	rightChild := m.newKDtree(rightBoundingBox, rightTriangles, depth+1)
	node.Children[0] = leftChild
	node.Children[1] = rightChild
	return node
}

func (m *Mesh) IntersectKD(ray *Ray, boundingBox *BoundingBox, node *KDtree, intersectionInfo *Intersection) bool {
	foundIntersection := false
	if node.Axis == Leaf {
		for _, triangle := range node.Triangles {
			if m.intersectTriangle(ray, &m.Faces[triangle], intersectionInfo, boundingBox) {
				foundIntersection = true
			}
		}
		return foundIntersection
	}

	leftBoundingBoxChild, rightBoundingBoxChild := boundingBox.Split(node.Axis, node.Median)

	var firstBoundingBox, secondBoundingBox *BoundingBox
	var firstNodeChild, secondNodeChild *KDtree
	if GetDimension(&ray.Start, node.Axis) <= node.Median {
		firstBoundingBox = leftBoundingBoxChild
		secondBoundingBox = rightBoundingBoxChild
		firstNodeChild = node.Children[0]
		secondNodeChild = node.Children[1]
	} else {
		firstBoundingBox = rightBoundingBoxChild
		secondBoundingBox = leftBoundingBoxChild
		firstNodeChild = node.Children[1]
		secondNodeChild = node.Children[0]
	}

	if boundingBox.IntersectWall(node.Axis, node.Median, ray) {
		if m.IntersectKD(ray, firstBoundingBox, firstNodeChild, intersectionInfo) {
			return true
		}
		return m.IntersectKD(ray, secondBoundingBox, secondNodeChild, intersectionInfo)
	}
	if firstBoundingBox.Intersect(ray) {
		return m.IntersectKD(ray, firstBoundingBox, firstNodeChild, intersectionInfo)
	}
	return m.IntersectKD(ray, secondBoundingBox, secondNodeChild, intersectionInfo)
}
