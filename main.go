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
	g.Expand()
	g.Expand()

	// Find candidate distribution, or fail out if this is impossible.
	var blueHyacinths flower.GeneticDistribution
	cnt := 0
	g.VisitVertices(func(gd flower.GeneticDistribution) {
		for g, p := range gd {
			if p == 0 {
				continue
			}
			if hyacinths.Phenotype(flower.Genotype(g)) != "Blue" {
				return
			}
		}
		blueHyacinths = gd
		cnt++
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
	printDotGraphPathTo(hyacinths, g, blueHyacinths, names)
}

/*
func main() {
	// Initial flowers.
	mums := flower.Mums()
	seedWhite := must(mums.ParseGenotype("rryyWw")).ToGeneticDistribution()
	seedYellow := must(mums.ParseGenotype("rrYYWW")).ToGeneticDistribution()
	seedRed := must(mums.ParseGenotype("RRyyWW")).ToGeneticDistribution()
	greenMumsA := must(mums.ParseGenotype("RRYYWw")).ToGeneticDistribution()
	greenMumsB := must(mums.ParseGenotype("RRYYWW")).ToGeneticDistribution()

	// Breeding tests.
	tests := map[string]breedgraph.Test{
		"":       breedgraph.NoTest,
		"Green":  breedgraph.PhenotypeTest(mums, "Green"),
		"Pink":   breedgraph.PhenotypeTest(mums, "Pink"),
		"Purple": breedgraph.PhenotypeTest(mums, "Purple"),
		"Red":    breedgraph.PhenotypeTest(mums, "Red"),
		"White":  breedgraph.PhenotypeTest(mums, "White"),
		"Yellow": breedgraph.PhenotypeTest(mums, "Yellow"),
	}

	g := breedgraph.NewGraph(tests, []flower.GeneticDistribution{seedWhite, seedYellow, seedRed})
	g.Expand()
	g.Expand()

	// Find candidate distribution, or fail out if this is impossible.
	var greenMums flower.GeneticDistribution
	g.VisitVertices(func(gd flower.GeneticDistribution) {
		for g, p := range gd {
			if p == 0 {
				continue
			}
			if mums.Phenotype(flower.Genotype(g)) != "Green" {
				return
			}
		}
		greenMums = gd
	})
	if greenMums.IsZero() {
		fmt.Fprintf(os.Stderr, "No green mums possible.\n")
		os.Exit(1)
	}

	// Print result.
	names := map[flower.GeneticDistribution]string{}
	names[seedWhite] = "Seed White (rryyWw)"
	names[seedYellow] = "Seed Yellow (rrYYWW)"
	names[seedRed] = "Seed Red (RRyyWW)"
	names[greenMumsA] = "Green Mums (RRYYWw)"
	names[greenMumsB] = "Green Mums (RRYYWW)"
	printDotGraphPathTo(mums, g, greenMums, names)
}
*/

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
	g.VisitVertices(func(gd flower.GeneticDistribution) {
		fmt.Printf("  %s\n", name(gd))
	})

	fmt.Println("Lineage:")
	g.VisitEdges(func(pa, pb, c flower.GeneticDistribution, test string) {
		fmt.Printf("  %s and %s make %s [test = %s]\n", name(pa), name(pb), name(c), test)
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
	g.VisitVertices(func(gd flower.GeneticDistribution) {
		fmt.Printf(`  "%s"`, name(gd))
		fmt.Println()
	})
	fmt.Println()

	// Print edges.
	g.VisitEdges(func(pa, pb, c flower.GeneticDistribution, test string) {
		fmt.Printf(`  {"%s" "%s"} -> "%s" [headlabel = "%s"]`, name(pa), name(pb), name(c), test)
		fmt.Println()
	})
	fmt.Println("}")
}

func printDotGraphPathTo(s flower.Species, g *breedgraph.Graph, target flower.GeneticDistribution, names map[flower.GeneticDistribution]string) {
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
	g.VisitPathTo(target, func(gd flower.GeneticDistribution) {
		fmt.Printf(`  "%s"`, name(gd))
		fmt.Println()
	}, func(pa, pb, c flower.GeneticDistribution, test string) {
		fmt.Printf(`  {"%s" "%s"} -> "%s" [headlabel = "%s"]`, name(pa), name(pb), name(c), test)
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
