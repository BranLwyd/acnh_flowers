package main

import (
	"fmt"
	"os"

	"github.com/BranLwyd/acnh_flowers/breedgraph"
	"github.com/BranLwyd/acnh_flowers/flower"
)

func main() {
	// Initial flowers.
	hyacinths := flower.Hyacinths()
	seedWhite := must(hyacinths.ParseGenotype("rryyWw")).ToGeneticDistribution()
	seedYellow := must(hyacinths.ParseGenotype("rrYYWW")).ToGeneticDistribution()
	seedRed := must(hyacinths.ParseGenotype("RRyyWw")).ToGeneticDistribution()
	blueHyacinthsA := must(hyacinths.ParseGenotype("rryyww")).ToGeneticDistribution()
	blueHyacinthsB := must(hyacinths.ParseGenotype("RRYyWW")).ToGeneticDistribution()

	// Breeding tests.
	tests := map[string]breedgraph.Test{
		"":       breedgraph.NoTest,
		"Blue":   breedgraph.PhenotypeTest(hyacinths, "Blue"),
		"Orange": breedgraph.PhenotypeTest(hyacinths, "Orange"),
		"Pink":   breedgraph.PhenotypeTest(hyacinths, "Pink"),
		"Purple": breedgraph.PhenotypeTest(hyacinths, "Purple"),
		"Red":    breedgraph.PhenotypeTest(hyacinths, "Red"),
		"White":  breedgraph.PhenotypeTest(hyacinths, "White"),
		"Yellow": breedgraph.PhenotypeTest(hyacinths, "Yellow"),
	}

	g := breedgraph.NewGraph(tests, []flower.GeneticDistribution{seedWhite, seedYellow, seedRed})
	for i := 0; i < 3; i++ {
		fmt.Fprintf(os.Stderr, "Beginning graph expansion step %d\n", i+1)
		g.Expand()
	}

	// Find candidate distribution, or fail out if this is impossible.
	var blueHyacinths breedgraph.Vertex
	g.VisitVertices(func(v breedgraph.Vertex) {
		// Is this a suitable candidate?
		for g, p := range v.Value() {
			if p == 0 {
				continue
			}
			if hyacinths.Phenotype(flower.Genotype(g)) != "Blue" {
				return
			}
		}

		// It is a suitable candidate. Is it the cheapeast candidate we've found so far?
		if blueHyacinths.IsZero() || v.PathCost() < blueHyacinths.PathCost() {
			blueHyacinths = v
		}
	})
	if blueHyacinths.IsZero() {
		fmt.Fprintf(os.Stderr, "No blue hyacinths possible.\n")
		os.Exit(1)
	}

	// Print result.
	names := map[flower.GeneticDistribution]string{}
	names[seedWhite] = "Seed White (rryyWw)"
	names[seedYellow] = "Seed Yellow (rrYYWW)"
	names[seedRed] = "Seed Red (RRyyWw)"
	names[blueHyacinthsA] = "Blue Hyacinths (rryyww)"
	names[blueHyacinthsB] = "Blue Hyacinths (RRYyWW)"
	printDotGraphPathTo(hyacinths, blueHyacinths, names)
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
		return fmt.Sprintf("%s(%.2f)", test, cost)
	}
	return fmt.Sprintf("%.02f", cost)
}

func must(g flower.Genotype, err error) flower.Genotype {
	if err != nil {
		panic(err)
	}
	return g
}
