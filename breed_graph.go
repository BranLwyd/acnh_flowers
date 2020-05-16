package breedgraph

import (
	"github.com/BranLwyd/acnh_flowers/flower"
)

type Graph struct {
	tests map[string]Test

	verts        []*vertex
	vertMap      map[flower.GeneticDistribution]*vertex
	vertFrontier int

	edges []*edge
}

type edge struct {
	pred [2]*vertex
	succ *vertex

	testName string
	cost     float64
}

type vertex struct {
	gd flower.GeneticDistribution

	pred *edge
}

func NewGraph(tests map[string]Test, initialFlowers []flower.GeneticDistribution) *Graph {
	verts := make([]*vertex, len(initialFlowers))
	vertMap := map[flower.GeneticDistribution]*vertex{}
	for i, gd := range initialFlowers {
		v := &vertex{gd, nil}
		verts[i] = v
		vertMap[gd] = v
	}
	return &Graph{
		tests:   tests,
		verts:   verts,
		vertMap: vertMap,
	}
}

func (g *Graph) Expand() {
	initialVertCnt := len(g.verts)
	for i, pa := range g.verts[:initialVertCnt] {
		minJ := g.vertFrontier
		if i > minJ {
			minJ = i
		}
		for _, pb := range g.verts[minJ:initialVertCnt] {
			gd := pa.gd.Breed(pb.gd)

			for testName, test := range g.tests {
				gd, cost := test(gd)
				if gd.IsZero() {
					// Test can't be applied to this distribution.
					continue
				}
				e := &edge{pred: [2]*vertex{pa, pb}, testName: testName, cost: cost}
				g.edges = append(g.edges, e)

				if v, ok := g.vertMap[gd]; ok {
					// This vertex already exists. Update lowest-cost if necessary.
					e.succ = v
					oldPathCost, newPathCost := v.pathCost(), e.pathCost()
					if newPathCost < oldPathCost {
						v.pred = e
					}
					continue
				}

				// This vertex does not yet exist in the graph.
				v := &vertex{gd: gd, pred: e}
				e.succ = v
				g.verts = append(g.verts, v)
				g.vertMap[gd] = v
			}
		}
	}
	g.vertFrontier = initialVertCnt
}

func (g *Graph) VisitVertices(f func(flower.GeneticDistribution)) {
	for _, v := range g.verts {
		f(v.gd)
	}
}

func (g *Graph) VisitEdges(f func(parentA, parentB, child flower.GeneticDistribution, test string, cost float64)) {
	for _, e := range g.edges {
		f(e.pred[0].gd, e.pred[1].gd, e.succ.gd, e.testName, e.cost)
	}
}

func (g *Graph) VisitPathTo(gd flower.GeneticDistribution, vertexVisitor func(flower.GeneticDistribution), edgeVisitor func(parentA, parentB, child flower.GeneticDistribution, test string, cost float64)) {
	v, ok := g.vertMap[gd]
	if !ok {
		return
	}

	var verts []*vertex
	var edges []*edge

	v.visitPath(func(x interface{}) {
		switch x := x.(type) {
		case *vertex:
			verts = append(verts, x)
		case *edge:
			edges = append(edges, x)
		}
	})

	for _, v := range verts {
		vertexVisitor(v.gd)
	}
	for _, e := range edges {
		edgeVisitor(e.pred[0].gd, e.pred[1].gd, e.succ.gd, e.testName, e.cost)
	}
}

func (e *edge) pathCost() float64 {
	var cost float64
	e.visitPath(func(x interface{}) {
		if e, ok := x.(*edge); ok {
			cost += e.cost
		}
	})
	return cost
}

func (e *edge) visitPath(f func(interface{})) {
	stk := []*edge{e}
	handled := map[*edge]struct{}{}
	for len(stk) != 0 {
		var e *edge
		stk, e = stk[:len(stk)-1], stk[len(stk)-1]
		if _, ok := handled[e]; ok {
			continue
		}
		handled[e] = struct{}{}

		f(e)

		f(e.pred[0])
		if e.pred[0].pred != nil {
			stk = append(stk, e.pred[0].pred)
		}

		f(e.pred[1])
		if e.pred[1].pred != nil {
			stk = append(stk, e.pred[1].pred)
		}
	}
}

func (v *vertex) pathCost() float64 {
	if v.pred == nil {
		return 0
	}
	return v.pred.pathCost()
}

func (v *vertex) visitPath(f func(interface{})) {
	f(v)
	if v.pred != nil {
		v.pred.visitPath(f)
	}
}

type Test func(flower.GeneticDistribution) (_ flower.GeneticDistribution, cost float64)

var (
	NoTest Test = func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) { return gd, 1 }
)

func PhenotypeTest(s flower.Species, phenotype string) Test {
	return func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) {
		rslt := gd
		var succChances, totalChances uint64
		for g, p := range rslt {
			g := flower.Genotype(g)
			totalChances += p
			if s.Phenotype(g) == phenotype {
				succChances += p
			} else {
				rslt[g] = 0
			}
		}
		if succChances == 0 {
			// This test can't be applied.
			return flower.GeneticDistribution{}, 0
		}
		return rslt.Reduce(), float64(totalChances) / float64(succChances)
	}
}
