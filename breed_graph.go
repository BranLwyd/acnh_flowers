package breedgraph

import (
	"fmt"

	"github.com/BranLwyd/acnh_flowers/flower"
)

type Graph struct {
	tests map[string]Test

	verts        []*vertex
	vertMap      map[flower.GeneticDistribution]*vertex
	vertFrontier int
}

type vertex struct {
	gd   flower.GeneticDistribution
	pred *edge
}

type edge struct {
	pred [2]*vertex
	succ *vertex

	testName string
	cost     float64
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

func (g *Graph) VisitVertices(f func(Vertex)) {
	for _, v := range g.verts {
		f(Vertex{v})
	}
}

func (g *Graph) VisitEdges(f func(Edge)) {
	verts := make([]interface{}, len(g.verts))
	for i := range g.verts {
		verts[i] = g.verts[i]
	}
	visitSubgraphPathingToAllOf(verts, func(x interface{}) {
		if e, ok := x.(*edge); ok {
			f(Edge{e})
		}
	})
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
	visitSubgraphPathingToAllOf([]interface{}{e}, f)
}

func (v *vertex) pathCost() float64 {
	if v.pred == nil {
		return 0
	}
	return v.pred.pathCost()
}

func (v *vertex) visitPath(f func(interface{})) {
	visitSubgraphPathingToAllOf([]interface{}{v}, f)
}

// vertsAndEdges is MODIFIED & CONSUMED by this function.
func visitSubgraphPathingToAllOf(vertsAndEdges []interface{}, f func(interface{})) {
	stk := vertsAndEdges
	handled := map[interface{}]struct{}{}
	for len(stk) != 0 {
		var x interface{}
		stk, x = stk[:len(stk)-1], stk[len(stk)-1]
		if _, ok := handled[x]; ok {
			continue
		}
		handled[x] = struct{}{}

		f(x)
		switch x := x.(type) {
		case *vertex:
			if x.pred != nil {
				stk = append(stk, x.pred)
			}
		case *edge:
			stk = append(stk, x.pred[0], x.pred[1])
		default:
			panic(fmt.Sprintf("visitSubgraphsPathingTo: unexpected type %T", x))
		}
	}
}

type Test func(flower.GeneticDistribution) (_ flower.GeneticDistribution, cost float64)

var (
	NoTest Test = func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) { return gd, 1 }
)

func PhenotypeTest(s flower.Species, phenotype string) Test {
	return func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) {
		var succChances, totalChances uint64
		rslt := gd.Update(func(mgd *flower.MutableGeneticDistribution) {
			gd.Visit(func(g flower.Genotype, p uint64) {
				totalChances += p
				if s.Phenotype(g) == phenotype {
					succChances += p
				} else {
					mgd.SetOdds(g, 0)
				}
			})
		})
		if succChances == 0 {
			// This test can't be applied.
			return flower.GeneticDistribution{}, 0
		}
		return rslt, float64(totalChances) / float64(succChances)
	}
}

type Vertex struct{ v *vertex }

func (v Vertex) IsZero() bool                       { return v.v == nil }
func (v Vertex) Value() flower.GeneticDistribution  { return v.v.gd }
func (v Vertex) BestPredecessor() (_ Edge, ok bool) { return Edge{v.v.pred}, v.v.pred != nil }
func (v Vertex) PathCost() float64                  { return v.v.pathCost() }

func (v Vertex) VisitPathTo(vertexVisitor func(Vertex), edgeVisitor func(Edge)) {
	var verts []*vertex
	var edges []*edge

	v.v.visitPath(func(x interface{}) {
		switch x := x.(type) {
		case *vertex:
			verts = append(verts, x)
		case *edge:
			edges = append(edges, x)
		}
	})

	for _, v := range verts {
		vertexVisitor(Vertex{v})
	}
	for _, e := range edges {
		edgeVisitor(Edge{e})
	}
}

type Edge struct{ e *edge }

func (e Edge) IsZero() bool         { return e.e == nil }
func (e Edge) FirstParent() Vertex  { return Vertex{e.e.pred[0]} }
func (e Edge) SecondParent() Vertex { return Vertex{e.e.pred[1]} }
func (e Edge) Child() Vertex        { return Vertex{e.e.succ} }
func (e Edge) TestName() string     { return e.e.testName }
func (e Edge) EdgeCost() float64    { return e.e.cost }
func (e Edge) PathCost() float64    { return e.e.pathCost() }
