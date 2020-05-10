// Package flower provides functionality for working with Animal Crossing: New
// Horizons flower genetics.
package flower

import (
	"fmt"
)

const (
	threeGeneGenotypeCount = 64
)

// Species3 represents a specific species of flower with three genes, such as
// Windflower or Mum.
type Species3 struct {
	name string // a human-readable name for this species, e.g. "Windflower".

	// Mappings from numeric gene values to human-readable gene values.
	gene0 [3]string // possible human-readable values of the first gene, e.g. {rr, Rr, RR}
	gene1 [3]string // possible human-readable values of the second gene, e.g. {ww, Ww, WW}
	gene2 [3]string // possible human-readable values of the third gene, e.g. {yy, Yy, YY}
}

func NewSpecies3(name, genes string) Species3 {
	// TODO: validate genes
	return Species3{
		name:  name,
		gene0: [3]string{fmt.Sprintf("%[1]c%[1]c", genes[1]), genes[0:2], fmt.Sprintf("%[1]c%[1]c", genes[0])},
		gene1: [3]string{fmt.Sprintf("%[1]c%[1]c", genes[3]), genes[2:4], fmt.Sprintf("%[1]c%[1]c", genes[2])},
		gene2: [3]string{fmt.Sprintf("%[1]c%[1]c", genes[5]), genes[4:6], fmt.Sprintf("%[1]c%[1]c", genes[4])},
	}
}

func (s Species3) ParseGenotype(genotype string) (Genotype, error) {
	var rslt Genotype

	if len(genotype) != 6 {
		return 0, fmt.Errorf("genotype %q has wrong length (expected 6)", genotype)
	}

	for _, x := range []struct {
		gene   [3]string
		offset uint
	}{
		{s.gene0, 0},
		{s.gene1, 2},
		{s.gene2, 4},
	} {
		found := false
		for i, v := range x.gene {
			if v == genotype[x.offset:x.offset+2] {
				rslt |= Genotype(i << x.offset)
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("unparsable gene %q", genotype[x.offset:x.offset+2])
		}
	}
	return rslt, nil
}

func (s Species3) RenderGenotype(g Genotype) string {
	return fmt.Sprintf("%s%s%s", s.gene0[g.gene0()], s.gene1[g.gene1()], s.gene2[g.gene2()])
}

// Genotype represents a specific set of genes for a species, e.g. RrwwYY.
type Genotype uint8

// Internally, each two consecutive bits of a Genotype value represents a gene.
//  0 == 0b00 is dual-recessive (rr).
//  1 == 0b01 is dominant/recessive (Rr).
//  2 == 0b10 is dual-domninant (RR).
//  3 == 0b11 is unused.

func (g Genotype) gene0() uint8 { return uint8((g >> 0) & 0b11) }
func (g Genotype) gene1() uint8 { return uint8((g >> 2) & 0b11) }
func (g Genotype) gene2() uint8 { return uint8((g >> 4) & 0b11) }

func (g Genotype) ToGeneticDistribution3() GeneticDistribution3 {
	var rslt GeneticDistribution3
	rslt.dist[g] = 1
	return rslt
}

// GeneticDistribution3 represents a probability distribution over all possible
// genotypes for a three-gene species.
type GeneticDistribution3 struct {
	dist [threeGeneGenotypeCount]uint64 // TODO: is uint64 big enough?
}

func (gda GeneticDistribution3) Breed(gdb GeneticDistribution3) GeneticDistribution3 {
	var rslt GeneticDistribution3

	// Breed each pair of possible genotypes into the result.
	for ga, pa := range gda.dist {
		if pa == 0 {
			continue
		}
		ga := Genotype(ga)
		for gb, pb := range gdb.dist {
			if pb == 0 {
				continue
			}
			gb := Genotype(gb)
			rslt.breedInto(pa*pb, ga, gb)
		}
	}

	// Reduce the result.
	g := rslt.dist[0]
	for _, x := range rslt.dist[1:] {
		if g == 1 {
			break
		}
		g = gcd(g, x)
	}
	for i := range rslt.dist {
		rslt.dist[i] /= g
	}
	return rslt
}

func (gda *GeneticDistribution3) breedInto(weight uint64, ga, gb Genotype) {
	wt0 := punnetSquareLookupTable[ga.gene0()][gb.gene0()]
	wt1 := punnetSquareLookupTable[ga.gene1()][gb.gene1()]
	wt2 := punnetSquareLookupTable[ga.gene2()][gb.gene2()]

	for g0, w0 := range wt0 {
		for g1, w1 := range wt1 {
			for g2, w2 := range wt2 {
				gda.dist[g0|(g1<<2)|(g2<<4)] += weight * w0 * w1 * w2
			}
		}
	}
}

var (
	// TODO: generate this lookup table from code, to decrease odds of error
	punnetSquareLookupTable = [3][3][3]uint64{
		// ga == 0 (rr)
		[3][3]uint64{
			[3]uint64{4, 0, 0},
			[3]uint64{2, 2, 0},
			[3]uint64{0, 4, 0},
		},

		// ga = 1 (Rr)
		[3][3]uint64{
			[3]uint64{2, 2, 0},
			[3]uint64{1, 2, 1},
			[3]uint64{0, 2, 2},
		},

		// ga = 2 (RR)
		[3][3]uint64{
			[3]uint64{0, 4, 0},
			[3]uint64{0, 2, 2},
			[3]uint64{0, 0, 4},
		},
	}
)

// Based on https://en.wikipedia.org/wiki/Binary_GCD_algorithm#Iterative_version_in_C.
func gcd(u, v uint64) uint64 {
	// Base cases.
	if u == 0 {
		return v
	}
	if v == 0 {
		return u
	}

	// Remove largest factor of 2.
	shift := 0
	for (u|v)&1 == 0 {
		shift++
		u >>= 1
		v >>= 1
	}

	// Remove additional, non-common factors of 2 from u.
	for u&1 == 0 {
		u >>= 1
	}

	// Loop invariant: u is odd.
	for v != 0 {
		for v&1 == 0 {
			v >>= 1
		}
		if u > v {
			u, v = v, u
		}
		v -= u
	}
	return u << shift
}
