package main

import (
	"fmt"

	"github.com/BranLwyd/acnh_flowers/breedgraph"
	"github.com/BranLwyd/acnh_flowers/flower"
)

func main() {
	s := flower.NewSpecies3("Tulips", "RrYySs")
	seedWhite := must(s.ParseGenotype("rryySs")).ToGeneticDistribution3()
	seedYellow := must(s.ParseGenotype("rrYYss")).ToGeneticDistribution3()
	seedRed := must(s.ParseGenotype("RRyySs")).ToGeneticDistribution3()

	names := map[flower.GeneticDistribution3]string{}
	names[seedWhite] = "Seed White (rryySs)"
	names[seedYellow] = "Seed Yellow (rrYYss)"
	names[seedRed] = "Seed Red (RRyySs)"

	g := breedgraph.NewGraph3(seedWhite, seedYellow, seedRed)
	g.Expand()

	printDotGraph(s, g, names)
}

func printGraph(s flower.Species3, g *breedgraph.Graph3, names map[flower.GeneticDistribution3]string) {
	name := func(gd flower.GeneticDistribution3) string {
		if name, ok := names[gd]; ok {
			return name
		}
		name := s.RenderGeneticDistribution3(gd)
		names[gd] = name
		return name
	}

	fmt.Println("All flowers:")
	g.VisitVertices(func(gd flower.GeneticDistribution3) {
		fmt.Printf("  %s\n", name(gd))
	})

	fmt.Println("Lineage:")
	g.VisitEdges(func(pa, pb, c flower.GeneticDistribution3) {
		fmt.Printf("  %s and %s make %s\n", name(pa), name(pb), name(c))
	})

}

func printDotGraph(s flower.Species3, g *breedgraph.Graph3, names map[flower.GeneticDistribution3]string) {
	name := func(gd flower.GeneticDistribution3) string {
		if name, ok := names[gd]; ok {
			return name
		}
		name := s.RenderGeneticDistribution3(gd)
		names[gd] = name
		return name
	}

	// Print vertices.
	fmt.Println("digraph {")
	g.VisitVertices(func(gd flower.GeneticDistribution3) {
		fmt.Printf(`  "%s"`, name(gd))
		fmt.Println()
	})
	fmt.Println()

	// Print edges.
	g.VisitEdges(func(pa, pb, c flower.GeneticDistribution3) {
		fmt.Printf(`  {"%s" "%s"} -> "%s"`, name(pa), name(pb), name(c))
		fmt.Println()
	})
	fmt.Println("}")
}

func must(g flower.Genotype, err error) flower.Genotype {
	if err != nil {
		panic(err)
	}
	return g
}
