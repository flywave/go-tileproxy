package ray

import (
	"math"
	"unsafe"

	vec3d "github.com/flywave/go3d/float64/vec3"
)

func solveEquation(a, b, c vec3d.T) (float64, float64) {
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

func negative(v *vec3d.T) vec3d.T {
	return v.Scaled(-1)
}

func mixedProduct(a, b, c *vec3d.T) float64 {
	ab := vec3d.Cross(a, b)
	return vec3d.Dot(&ab, c)
}

const (
	MaxTreeDepth     = 10
	TrianglesPerLeaf = 20
)

type Triangle struct {
	AB, AC, ABxAC *vec3d.T `json:"-"`
}

type RayMesh struct {
	Vertices  []vec3d.T  `json:"vertices"`
	Faces     [][3]int   `json:"faces"`
	Triangles []Triangle `json:"-"`
	tree      *KDtree    `json:"-"`
	BBox      *BBox      `json:"-"`
}

func NewRayMesh(vertices [][3]float64, faces [][3]int, bbox [2][3]float64) *RayMesh {
	vvecs := *(*[]vec3d.T)(unsafe.Pointer(&vertices))
	vbbox := *(*[2]vec3d.T)(unsafe.Pointer(&bbox))
	ret := &RayMesh{Vertices: vvecs, Faces: faces, BBox: NewBBoxWithData(vbbox)}
	ret.Init()
	return ret
}

func (m *RayMesh) Init() {
	allIndices := make([]int, len(m.Faces))

	for i := range allIndices {
		allIndices[i] = i
	}

	m.Triangles = make([]Triangle, len(m.Faces))
	if m.BBox == nil {
		m.BBox = m.GetBBox()
	}
	m.tree = m.newKDtree(m.BBox, allIndices, 0)
	for i := range m.Faces {
		triangle := &m.Faces[i]

		A := &m.Vertices[triangle[0]]
		B := &m.Vertices[triangle[1]]
		C := &m.Vertices[triangle[2]]

		AB := vec3d.Sub(B, A)
		AC := vec3d.Sub(C, A)

		m.Triangles[i].AB = &AB
		m.Triangles[i].AC = &AC
		c := vec3d.Cross(&AB, &AC)
		m.Triangles[i].ABxAC = &c
	}
}

func (m *RayMesh) SlowIntersect(incoming *Ray) *Intersection {
	intersection := &Intersection{}
	intersection.Distance = Inf
	found := false
	for i, triangle := range m.Faces {
		if m.intersectTriangle(incoming, triangle, &m.Triangles[i], intersection, nil) {
			found = true
		}
	}
	if !found {
		return nil
	}
	return intersection
}

func (m *RayMesh) Intersect(incoming *Ray) *Intersection {
	incoming.Init()
	if !m.BBox.Intersect(incoming) {
		return nil
	}
	intersectionInfo := &Intersection{Distance: Inf}
	if m.IntersectKD(incoming, m.BBox, m.tree, intersectionInfo) {
		return intersectionInfo
	}
	return nil
}

func IntersectTriangle(ray *Ray, A, B, C *vec3d.T) (bool, float64) {
	AB := vec3d.Sub(B, A)
	AC := vec3d.Sub(C, A)
	reverseDirection := negative(&ray.Direction)
	distToA := vec3d.Sub(&ray.Start, A)
	ABxAC := vec3d.Cross(&AB, &AC)
	det := vec3d.Dot(&ABxAC, &reverseDirection)
	reverseDet := 1 / det
	if math.Abs(det) < Epsilon {
		return false, Inf
	}
	lambda2 := mixedProduct(&distToA, &AC, &reverseDirection) * reverseDet
	lambda3 := mixedProduct(&AB, &distToA, &reverseDirection) * reverseDet
	gamma := vec3d.Dot(&ABxAC, &distToA) * reverseDet
	if gamma < 0 {
		return false, Inf
	}
	if lambda2 < 0 || lambda2 > 1 || lambda3 < 0 || lambda3 > 1 || lambda2+lambda3 > 1 {
		return false, Inf
	}
	return true, gamma
}

func (m *RayMesh) intersectTriangle(ray *Ray, vertices [3]int, triangle *Triangle, intersection *Intersection, bbox *BBox) bool {
	A := &m.Vertices[vertices[0]]
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
	lambda2 := mixedProduct(&distToA, &rayDir, triangle.AC) * reverseDet
	lambda3 := -mixedProduct(&distToA, &rayDir, triangle.AB) * reverseDet
	if lambda2 < 0 || lambda2 > 1 || lambda3 < 0 || lambda3 > 1 || lambda2+lambda3 > 1 {
		return false
	}

	sc := rayDir.Scaled(intersectDist)
	ip := vec3d.Add(&ray.Start, &sc)

	if bbox != nil && !bbox.Inside(ip) {
		return false
	}
	intersection.Point = &ip
	intersection.Distance = intersectDist
	intersection.Incoming = ray
	return true
}

func (m *RayMesh) GetBBox() *BBox {
	bbox := NewBBox()
	for _, vertex := range m.Vertices {
		bbox.AddPoint(vertex)
	}
	return bbox
}

func (m *RayMesh) newKDtree(bbox *BBox, trianglesIndices []int, depth int) *KDtree {
	if depth > MaxTreeDepth || len(trianglesIndices) < TrianglesPerLeaf {
		node := NewLeaf(trianglesIndices)
		return node
	}
	axis := (depth + 2) % 3
	leftLimit := bbox.MaxVolume[axis]
	righLimit := bbox.MinVolume[axis]

	median := (leftLimit + righLimit) / 2

	var leftTriangles, rightTriangles []int
	var A, B, C *vec3d.T
	leftBBox, rightBBox := bbox.Split(axis, median)
	for _, index := range trianglesIndices {
		A = &m.Vertices[m.Faces[index][0]]
		B = &m.Vertices[m.Faces[index][1]]
		C = &m.Vertices[m.Faces[index][2]]

		if leftBBox.IntersectTriangle(*A, *B, *C) {
			leftTriangles = append(leftTriangles, index)
		}

		if rightBBox.IntersectTriangle(*A, *B, *C) {
			rightTriangles = append(rightTriangles, index)
		}
	}
	node := NewNode(median, axis)
	leftChild := m.newKDtree(leftBBox, leftTriangles, depth+1)
	rightChild := m.newKDtree(rightBBox, rightTriangles, depth+1)
	node.Children[0] = leftChild
	node.Children[1] = rightChild
	return node
}

func (m *RayMesh) IntersectKD(ray *Ray, bbox *BBox, node *KDtree, intersectionInfo *Intersection) bool {
	foundIntersection := false
	if node.Axis == Leaf {
		for _, triangle := range node.Triangles {
			if m.intersectTriangle(ray, m.Faces[triangle], &m.Triangles[triangle], intersectionInfo, bbox) {
				foundIntersection = true
			}
		}
		return foundIntersection
	}

	leftBBoxChild, rightBBoxChild := bbox.Split(node.Axis, node.Median)

	var firstBBox, secondBBox *BBox
	var firstNodeChild, secondNodeChild *KDtree

	if getDimension(&ray.Start, node.Axis) <= node.Median {
		firstBBox = leftBBoxChild
		secondBBox = rightBBoxChild
		firstNodeChild = node.Children[0]
		secondNodeChild = node.Children[1]
	} else {
		firstBBox = rightBBoxChild
		secondBBox = leftBBoxChild
		firstNodeChild = node.Children[1]
		secondNodeChild = node.Children[0]
	}

	if bbox.IntersectWall(node.Axis, node.Median, ray) {
		if m.IntersectKD(ray, firstBBox, firstNodeChild, intersectionInfo) {
			return true
		}
		return m.IntersectKD(ray, secondBBox, secondNodeChild, intersectionInfo)
	}
	if firstBBox.Intersect(ray) {
		return m.IntersectKD(ray, firstBBox, firstNodeChild, intersectionInfo)
	}
	return m.IntersectKD(ray, secondBBox, secondNodeChild, intersectionInfo)
}
