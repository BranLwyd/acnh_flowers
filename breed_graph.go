package breedgraph

import (
	"github.com/BranLwyd/acnh_flowers/flower"
)

type edge3 struct {
	pred [2]*vertex3
	succ *vertex3
}

type vertex3 struct {
	gd flower.GeneticDistribution3
}

type Graph3 struct {
	verts        []*vertex3
	vertMap      map[flower.GeneticDistribution3]*vertex3
	vertFrontier int

	edges []*edge3
}

func NewGraph3(initialFlowers ...flower.GeneticDistribution3) *Graph3 {
	verts := make([]*vertex3, len(initialFlowers))
	vertMap := map[flower.GeneticDistribution3]*vertex3{}
	for i, gd := range initialFlowers {
		v := &vertex3{gd}
		verts[i] = v
		vertMap[gd] = v
	}
	return &Graph3{verts: verts, vertMap: vertMap}
}

func (g *Graph3) Expand() {
	initialVertCnt := len(g.verts)
	for i, pa := range g.verts[:initialVertCnt] {
		minJ := g.vertFrontier
		if i > minJ {
			minJ = i
		}
		for _, pb := range g.verts[minJ:initialVertCnt] {
			gd := pa.gd.Breed(pb.gd)
			if _, ok := g.vertMap[gd]; ok {
				// Update lowest-cost once costing is added.
				continue
			}
			v := &vertex3{gd}
			g.verts = append(g.verts, v)
			g.vertMap[gd] = v
			g.edges = append(g.edges, &edge3{pred: [2]*vertex3{pa, pb}, succ: v})
		}
	}
	g.vertFrontier = initialVertCnt
}

func (g *Graph3) VisitVertices(f func(flower.GeneticDistribution3)) {
	for _, v := range g.verts {
		f(v.gd)
	}
}

func (g *Graph3) VisitEdges(f func(parentA, parentB, child flower.GeneticDistribution3)) {
	for _, e := range g.edges {
		f(e.pred[0].gd, e.pred[1].gd, e.succ.gd)
	}
}
