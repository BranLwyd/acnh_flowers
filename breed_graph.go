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
				testedGD, _ := test(gd)
				if testedGD.IsZero() {
					// Test can't be applied to this distribution.
					continue
				}
				if _, ok := g.vertMap[testedGD]; ok {
					// Update lowest-cost once costing is added.
					continue
				}

				v := &vertex{gd: testedGD}
				e := &edge{pred: [2]*vertex{pa, pb}, succ: v, testName: testName}
				v.pred = e

				g.verts = append(g.verts, v)
				g.vertMap[testedGD] = v
				g.edges = append(g.edges, e)
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

func (g *Graph) VisitEdges(f func(parentA, parentB, child flower.GeneticDistribution, test string)) {
	for _, e := range g.edges {
		f(e.pred[0].gd, e.pred[1].gd, e.succ.gd, e.testName)
	}
}

func (g *Graph) VisitPathTo(gd flower.GeneticDistribution, vertexVisitor func(flower.GeneticDistribution), edgeVisitor func(parentA, parentB, child flower.GeneticDistribution, test string)) {
	v, ok := g.vertMap[gd]
	if !ok {
		return
	}

	var verts []*vertex
	var edges []*edge

	handled := map[flower.GeneticDistribution]struct{}{}
	stk := []*vertex{v}

	for len(stk) != 0 {
		var v *vertex
		stk, v = stk[:len(stk)-1], stk[len(stk)-1]

		if _, ok := handled[v.gd]; ok {
			continue
		}
		handled[v.gd] = struct{}{}

		verts = append(verts, v)
		if v.pred != nil {
			edges = append(edges, v.pred)
			stk = append(stk, v.pred.pred[0], v.pred.pred[1])
		}
	}

	for _, v := range verts {
		vertexVisitor(v.gd)
	}
	for _, e := range edges {
		edgeVisitor(
			e.pred[0].gd,
			e.pred[1].gd,
			e.succ.gd,
			e.testName)
	}
}

type Test func(flower.GeneticDistribution) (_ flower.GeneticDistribution, cost float64)

var (
	NoTest Test = func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) { return gd, 1 }
)

func PhenotypeTest(s flower.Species, phenotype string) Test {
	return func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) {
		var rslt flower.GeneticDistribution = gd
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
		return rslt, float64(totalChances) / float64(succChances)
	}
}
