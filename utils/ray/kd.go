package ray

import (
	"fmt"
)

type KDtree struct {
	Axis      int
	Median    float64
	Triangles []int
	Children  [2]*KDtree
}

func NewLeaf(triangles []int) *KDtree {
	return &KDtree{
		Axis:      Leaf,
		Triangles: triangles,
	}
}

func NewNode(median float64, axis int) *KDtree {
	var children [2]*KDtree
	return &KDtree{
		Axis:     axis,
		Median:   median,
		Children: children,
	}
}

func (t *KDtree) String() string {
	if t.Axis == Leaf {
		return fmt.Sprintf("%v", t.Triangles)
	}
	return fmt.Sprintf("%d{%.3g}(%s, %s)", t.Axis, t.Median, t.Children[0], t.Children[1])
}
