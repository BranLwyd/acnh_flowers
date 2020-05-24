package breedgraph

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/BranLwyd/acnh_flowers/flower"
)

type Graph struct {
	tests []*Test

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

	test *Test
	cost float64
}

func NewGraph(tests []*Test, initialFlowers []flower.GeneticDistribution) *Graph {
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

func (g *Graph) Search(pred func(flower.GeneticDistribution) bool) (_ Vertex, ok bool) {
	var rslt *vertex
	for _, v := range g.verts {
		if pred(v.gd) {
			if rslt == nil || v.pathCost() < rslt.pathCost() {
				rslt = v
			}
		}
	}
	return Vertex{rslt}, rslt != nil
}

func (g *Graph) Expand(keepPred func(flower.GeneticDistribution) bool) {
	initialVertCnt := len(g.verts)

	type result struct {
		e  *edge
		gd flower.GeneticDistribution
	}
	rsltsCh := make(chan []result)
	rsltsPool := &sync.Pool{New: func() interface{} { return []result(nil) }}

	// Spawn workers.
	totalWorkerCnt := runtime.GOMAXPROCS(0)
	workerCnt := int64(totalWorkerCnt)
	for i := 0; i < totalWorkerCnt; i++ {
		go func(base int) {
			defer func() {
				if atomic.AddInt64(&workerCnt, -1) == 0 {
					close(rsltsCh)
				}
			}()

			for i := base; i < initialVertCnt; i += totalWorkerCnt {
				rslts := rsltsPool.Get().([]result)
				va := g.verts[i]
				minJ := g.vertFrontier
				if i > minJ {
					minJ = i
				}
				for _, vb := range g.verts[minJ:initialVertCnt] {
					gd := va.gd.Breed(vb.gd)
					for _, test := range g.tests {
						gd, cost := test.Test(gd)
						if gd.IsZero() {
							// Test can't be applied to this distribution.
							continue
						}
						e := &edge{pred: [2]*vertex{va, vb}, test: test, cost: cost}
						rslts = append(rslts, result{e, gd})
					}
				}
				rsltsCh <- rslts
			}
		}(i)
	}

	// Handle results.
	for rslts := range rsltsCh {
		for _, rslt := range rslts {
			e, gd := rslt.e, rslt.gd
			if v, ok := g.vertMap[gd]; ok {
				// This vertex already exists. Update lowest-cost if necessary.
				oldPathCost, newPathCost := v.pathCost(), e.pathCost()
				if newPathCost < oldPathCost || (newPathCost == oldPathCost && e.test.Priority() < v.pred.test.Priority()) {
					e.succ, v.pred = v, e
				}
				continue
			}
			// This vertex does not yet exist in the graph. Create a new vertex, as long as the caller wants to keep it.
			if !keepPred(gd) {
				// Caller does not want us to keep this result.
				continue
			}
			v := &vertex{gd: gd, pred: e}
			e.succ, v.pred = v, e
			g.verts = append(g.verts, v)
			g.vertMap[gd] = v
		}
		rsltsPool.Put(rslts[:0])
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

type Test struct {
	name     string
	priority int
	test     func(flower.GeneticDistribution) (_ flower.GeneticDistribution, cost float64)
}

func (t *Test) Name() string  { return t.name }
func (t *Test) Priority() int { return t.priority }

func (t *Test) Test(gd flower.GeneticDistribution) (_ flower.GeneticDistribution, cost float64) {
	return t.test(gd)
}

var (
	NoTest *Test = &Test{"", 0, func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) { return gd, 1 }}
)

func PhenotypeTest(s flower.Species, phenotypes ...string) *Test {
	validPhenotype := func(phenotype string) bool {
		for _, ph := range phenotypes {
			if phenotype == ph {
				return true
			}
		}
		return false
	}

	name := fmt.Sprintf("P∈{%s}", strings.Join(phenotypes, ","))
	priority := len(phenotypes)
	return &Test{name, priority, func(gd flower.GeneticDistribution) (flower.GeneticDistribution, float64) {
		var succChances, totalChances uint64
		rslt := gd.Update(func(mgd *flower.MutableGeneticDistribution) {
			gd.Visit(func(g flower.Genotype, p uint64) bool {
				totalChances += p
				if validPhenotype(s.Phenotype(g)) {
					succChances += p
				} else {
					mgd.SetOdds(g, 0)
				}
				return true
			})
		})
		if succChances == 0 {
			// This test can't be applied.
			return flower.GeneticDistribution{}, 0
		}
		return rslt, float64(totalChances) / float64(succChances)
	}}
}

func PhenotypeTests(s flower.Species) []*Test {
	const maxInt = int(^uint(0) >> 1)
	return PhenotypeTestsUpToSize(s, maxInt)
}

func PhenotypeTestsUpToSize(s flower.Species, size int) []*Test {
	phenotypes := s.Phenotypes()
	if size >= len(phenotypes) {
		size = len(phenotypes) - 1
	}

	bits := make([]bool, len(phenotypes))
	var next func([]bool) bool
	next = func(bits []bool) bool {
		// Find lowest-order set bit.
		i := 0
		for i < len(bits) && !bits[i] {
			i++
		}
		if i == len(bits) {
			// No bits set at all? Can't increment.
			return false
		}

		bits[i] = false
		switch {
		case i == len(bits)-1:
			// The first bit set was the highest-order bit, so we can't increment any further.
			return false
		case !bits[i+1]:
			// The next-higher bit is not already set: just set it. We can try to increment again.
			bits[i+1] = true
			return true
		default:
			// The next-higher bit is already set: set the lowest-order bit and increment the remainder.
			bits[0] = true
			return next(bits[1:])
		}
	}

	rslt := []*Test{}
	for sz := 1; sz <= size; sz++ {
		// Initialize bits (first `sz` bits set).
		for i := range bits {
			bits[i] = (i < sz)
		}

		for {
			ps := make([]string, 0, sz)
			for i := range bits {
				if bits[i] {
					ps = append(ps, phenotypes[i])
				}
			}
			rslt = append(rslt, PhenotypeTest(s, ps...))
			if !next(bits) {
				break
			}
		}
	}
	return rslt
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
func (e Edge) Test() *Test          { return e.e.test }
func (e Edge) EdgeCost() float64    { return e.e.cost }
func (e Edge) PathCost() float64    { return e.e.pathCost() }
