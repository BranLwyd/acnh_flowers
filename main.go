package main

import (
	"fmt"
	"os"

	"github.com/BranLwyd/acnh_flowers/breedgraph"
	"github.com/BranLwyd/acnh_flowers/flower"
)

func main() {
	// Initial flowers.
	roses := flower.Roses()
	seedWhite := must(roses.ParseGenotype("rryyWwss")).ToGeneticDistribution()
	seedYellow := must(roses.ParseGenotype("rrYYWWss")).ToGeneticDistribution()
	seedRed := must(roses.ParseGenotype("RRyyWWSs")).ToGeneticDistribution()
	blueRoses := must(roses.ParseGenotype("RRYYwwss")).ToGeneticDistribution()

	// Breeding tests.
	tests := map[string]breedgraph.Test{"": breedgraph.NoTest}
	for name, test := range breedgraph.PhenotypeTests(roses) {
		tests[name] = test
	}

	g := breedgraph.NewGraph(tests, []flower.GeneticDistribution{seedWhite, seedYellow, seedRed})
	for i := 0; i < 3; i++ {
		fmt.Fprintf(os.Stderr, "Beginning graph expansion step %d...\n", i+1)
		g.Expand()
	}

	// Find candidate distribution, or fail out if this is impossible.
	candidate, ok := g.Search(func(gd flower.GeneticDistribution) bool {
		isSuitable := true
		gd.Visit(func(g flower.Genotype, _ uint64) {
			if roses.Phenotype(g) != "Blue" {
				isSuitable = false
			}
		})
		return isSuitable
	})
	if !ok {
		fmt.Fprintf(os.Stderr, "No blue roses possible.\n")
		os.Exit(1)
	}

	// Print result.
	names := map[flower.GeneticDistribution]string{}
	names[seedWhite] = "Seed White (rryyWwss)"
	names[seedYellow] = "Seed Yellow (rrYYWWss)"
	names[seedRed] = "Seed Red (RRyyWWSs)"
	names[blueRoses] = "Blue Roses (RRYYwwss)"
	printDotGraphPathTo(roses, candidate, names)
}

func printGraph(s flower.Species, g *breedgraph.Graph, names map[flower.GeneticDistribution]string) {
	name := func(gd flower.GeneticDistribution) string {
		if name, ok := names[gd]; ok {
			return name
		}
		name := s.RenderGeneticDistribution(gd)
		names[gd] = name
		return name
	}

	fmt.Println("All flowers:")
	g.VisitVertices(func(v breedgraph.Vertex) {
		fmt.Printf("  %s\n", name(v.Value()))
	})

	fmt.Println("Lineage:")
	g.VisitEdges(func(e breedgraph.Edge) {
		fmt.Printf("  %s and %s make %s [test = %q, cost = %.02f]\n", name(e.FirstParent().Value()), name(e.SecondParent().Value()), name(e.Child().Value()), e.TestName(), e.EdgeCost())
	})

}

func printDotGraph(s flower.Species, g *breedgraph.Graph, names map[flower.GeneticDistribution]string) {
	name := func(gd flower.GeneticDistribution) string {
		if name, ok := names[gd]; ok {
			return name
		}
		name := s.RenderGeneticDistribution(gd)
		names[gd] = name
		return name
	}

	// Print vertices.
	fmt.Println("digraph {")
	g.VisitVertices(func(v breedgraph.Vertex) {
		fmt.Printf(`  "%s"`, name(v.Value()))
		fmt.Println()
	})
	fmt.Println()

	// Print edges.
	g.VisitEdges(func(e breedgraph.Edge) {
		fmt.Printf(`  {"%s" "%s"} -> "%s" [headlabel="%s"]`, name(e.FirstParent().Value()), name(e.SecondParent().Value()), name(e.Child().Value()), edgeLabel(e.TestName(), e.EdgeCost()))
		fmt.Println()
	})
	fmt.Println("}")
}

func printDotGraphPathTo(s flower.Species, v breedgraph.Vertex, names map[flower.GeneticDistribution]string) {
	name := func(gd flower.GeneticDistribution) string {
		if name, ok := names[gd]; ok {
			return name
		}
		name := s.RenderGeneticDistribution(gd)
		names[gd] = name
		return name
	}

	// Print vertices.
	fmt.Println("digraph {")
	v.VisitPathTo(func(v breedgraph.Vertex) {
		fmt.Printf(`  "%s"`, name(v.Value()))
		fmt.Println()
	}, func(e breedgraph.Edge) {
		fmt.Printf(`  {"%s" "%s"} -> "%s" [headlabel="%s"]`, name(e.FirstParent().Value()), name(e.SecondParent().Value()), name(e.Child().Value()), edgeLabel(e.TestName(), e.EdgeCost()))
		fmt.Println()
	})
	fmt.Println("}")
}

func edgeLabel(test string, cost float64) string {
	if test != "" {
		return fmt.Sprintf("%s (%.2f)", test, cost)
	}
	return fmt.Sprintf("%.02f", cost)
}

func must(g flower.Genotype, err error) flower.Genotype {
	if err != nil {
		panic(err)
	}
	return g
}
