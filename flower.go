// Package flower provides functionality for working with Animal Crossing: New
// Horizons flower genetics.
package flower

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func Cosmos() Species      { return cosmos }
func Hyacinths() Species   { return hyacinths }
func Lilies() Species      { return lilies }
func Mums() Species        { return mums }
func Pansies() Species     { return pansies }
func Roses() Species       { return roses }
func Tulips() Species      { return tulips }
func Windflowers() Species { return windflowers }

// Species represents a specific species of flower, such as Windflower or Mum.
type Species struct {
	name       string              // a human-readable name for this species, e.g. "Windflowers".
	phenotypes map[Genotype]string // phenotypes by genotype
	serde      GenotypeSerde       // the (default) serializer/deserializer for genotypes; also determines gene count
}

func newSpecies(name string, phenotypes map[string]string) (Species, error) {
	s := Species{name: name}
	gsInit := false
	var gs GenotypeSerde
	pts := map[Genotype]string{}
	for gStr, p := range phenotypes {
		if !gsInit {
			serde, err := NewGenotypeSerdeFromExample(gStr)
			if err != nil {
				return Species{}, fmt.Errorf("couldn't parse genotype %q: %v", gStr, err)
			}
			gs, gsInit = serde, true
		}

		g, err := gs.ParseGenotype(gStr)
		if err != nil {
			return Species{}, err
		}
		pts[g] = p
	}
	s.phenotypes = pts
	s.serde = gs

	if gs.GeneCount() == 3 && len(s.phenotypes) != 27 {
		return Species{}, fmt.Errorf("got %d phenotypes, expected 27", len(phenotypes))
	}
	if gs.GeneCount() == 4 && len(s.phenotypes) != 81 {
		return Species{}, fmt.Errorf("got %d phenotypes, expected 81", len(phenotypes))
	}

	return s, nil
}

func mustSpecies(name string, phenotypes map[string]string) Species {
	s, err := newSpecies(name, phenotypes)
	if err != nil {
		panic(fmt.Sprintf("Could not create species %q: %v", name, err))
	}
	return s
}

func (s Species) Name() string                { return s.name }
func (s Species) GeneCount() int              { return s.serde.GeneCount() }
func (s Species) Phenotype(g Genotype) string { return s.phenotypes[g] }
func (s Species) ParseGenotype(genotype string) (Genotype, error) {
	return s.serde.ParseGenotype(genotype)
}
func (s Species) RenderGenotype(g Genotype) string { return s.serde.RenderGenotype(g) }
func (s Species) RenderGeneticDistribution(gd GeneticDistribution) string {
	return s.serde.RenderGeneticDistribution(gd)
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
func (g Genotype) gene3() uint8 { return uint8((g >> 6) & 0b11) }

func (g Genotype) ToGeneticDistribution() GeneticDistribution {
	var gd GeneticDistribution
	gd[g] = 1
	return gd
}

type GenotypeSerde struct {
	gene0 [3]string // contents of these will be something like {"rr", "Rr", "RR"}
	gene1 [3]string
	gene2 [3]string
	gene3 [3]string // {"", "", ""} for 3-gene species
}

func NewGenotypeSerdeFromExample(genotype string) (GenotypeSerde, error) {
	if len(genotype) != 6 && len(genotype) != 8 {
		return GenotypeSerde{}, fmt.Errorf("genotype %q has wrong length (expected 6 or 8)", genotype)
	}

	genesFrom := func(gene string) ([3]string, error) {
		lo, hi := strings.ToLower(gene[0:1]), strings.ToUpper(gene[0:1])
		genes := [3]string{lo + lo, hi + lo, hi + hi}
		if gene != genes[0] && gene != genes[1] && gene != genes[2] {
			return [3]string{}, fmt.Errorf("could not parse gene %q", gene)
		}
		return genes, nil
	}

	gene0, err := genesFrom(genotype[0:2])
	if err != nil {
		return GenotypeSerde{}, err
	}
	gene1, err := genesFrom(genotype[2:4])
	if err != nil {
		return GenotypeSerde{}, err
	}
	gene2, err := genesFrom(genotype[4:6])
	if err != nil {
		return GenotypeSerde{}, err
	}
	var gene3 [3]string
	if len(genotype) == 8 {
		gene3, err = genesFrom(genotype[6:8])
		if err != nil {
			return GenotypeSerde{}, err
		}
	}

	if gene0 == gene1 || gene0 == gene2 || gene0 == gene3 || gene1 == gene2 || gene1 == gene3 || gene2 == gene3 {
		return GenotypeSerde{}, fmt.Errorf("duplicate gene letters (%q, %q, %q, %q)", gene0[0], gene1[0], gene2[0], gene3[0])
	}

	return GenotypeSerde{gene0, gene1, gene2, gene3}, nil
}

func NewGenotypeSerdeFromExampleDistribution(geneticDistribution string) (GenotypeSerde, error) {
	_, gs, err := parseGeneticDistribution(GenotypeSerde{}, geneticDistribution)
	return gs, err
}

func (gs GenotypeSerde) IsZero() bool {
	var zero GenotypeSerde
	return gs == zero
}

func (gs GenotypeSerde) GeneCount() int {
	if gs.gene3[0] == "" {
		return 3
	}
	return 4
}

func (gs GenotypeSerde) ParseGenotype(genotype string) (Genotype, error) {
	var rslt Genotype

	if gs.gene3[0] == "" && len(genotype) != 6 {
		return 0, fmt.Errorf("genotype %q has wrong length (expected 6)", genotype)
	}
	if gs.gene3[0] != "" && len(genotype) != 8 {
		return 0, fmt.Errorf("genotype %q has wrong length (expected 8)", genotype)
	}

	for _, x := range []struct {
		gene   [3]string
		offset uint
	}{
		{gs.gene0, 0},
		{gs.gene1, 2},
		{gs.gene2, 4},
		{gs.gene3, 6},
	} {
		if x.gene[0] == "" {
			break
		}

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

func (gs GenotypeSerde) RenderGenotype(g Genotype) string {
	if gs.gene3[0] == "" {
		return fmt.Sprintf("%s%s%s", gs.gene0[g.gene0()], gs.gene1[g.gene1()], gs.gene2[g.gene2()])
	}
	return fmt.Sprintf("%s%s%s%s", gs.gene0[g.gene0()], gs.gene1[g.gene1()], gs.gene2[g.gene2()], gs.gene3[g.gene3()])
}

func (gs GenotypeSerde) ParseGeneticDistribution(geneticDistribution string) (GeneticDistribution, error) {
	gd, _, err := parseGeneticDistribution(gs, geneticDistribution)
	return gd, err
}

var genotypeRe = regexp.MustCompile(`^\w{6}(\w{2})?$`)

func parseGeneticDistribution(gs GenotypeSerde, geneticDistribution string) (GeneticDistribution, GenotypeSerde, error) {
	maybeCreateGS := func(geneticDistribution string) error {
		if !gs.IsZero() {
			return nil
		}
		newGS, err := NewGenotypeSerdeFromExample(geneticDistribution)
		if err != nil {
			return err
		}
		gs = newGS
		return nil
	}

	if genotypeRe.MatchString(geneticDistribution) {
		if err := maybeCreateGS(geneticDistribution); err != nil {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genotype as genetic distribution: %v", err)
		}
		gd, err := gs.ParseGenotype(geneticDistribution)
		if err != nil {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genotype as genetic distribution: %v", err)
		}
		return gd.ToGeneticDistribution(), gs, nil
	}

	var rslt GeneticDistribution
	if len(geneticDistribution) == 0 || geneticDistribution[0] != '{' || geneticDistribution[len(geneticDistribution)-1] != '}' {
		return GeneticDistribution{}, GenotypeSerde{}, errors.New("couldn't parse genetic distribution: not wrapped in curly quotes")
	}
	geneticDistribution = geneticDistribution[1 : len(geneticDistribution)-1]
	for _, term := range strings.Split(geneticDistribution, ",") {
		term = strings.TrimSpace(term)
		termSpl := strings.SplitN(term, ":", 2)
		if len(termSpl) != 2 {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genetic distribution: unparseable term %q", term)
		}

		odds, err := strconv.ParseUint(strings.TrimSpace(termSpl[0]), 10, 64)
		if err != nil {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genetic distribution: couldn't parse odds for term %q: %v", term, err)
		}
		if odds == 0 {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genetic distribution: couldn't parse odds for term %q: odds are zero", term)
		}

		if err := maybeCreateGS(termSpl[1]); err != nil {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genetic distribution: %v", err)
		}
		g, err := gs.ParseGenotype(strings.TrimSpace(termSpl[1]))
		if err != nil {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genetic distribution: couldn't parse genotype for term %q: %v", term, err)
		}
		if rslt[g] != 0 {
			return GeneticDistribution{}, GenotypeSerde{}, fmt.Errorf("couldn't parse genetic distribution: duplicate genotype %q", gs.RenderGenotype(g))
		}

		rslt[g] = odds
	}
	return rslt.Reduce(), gs, nil
}

func (gs GenotypeSerde) RenderGeneticDistribution(gd GeneticDistribution) string {
	var sb strings.Builder
	written := false
	sb.WriteString("{")
	for g, p := range gd {
		if p == 0 {
			continue
		}
		if written {
			sb.WriteString(", ")
		}
		sb.WriteString(strconv.FormatUint(p, 10))
		sb.WriteString(":")
		sb.WriteString(gs.RenderGenotype(Genotype(g)))
		written = true
	}
	sb.WriteString("}")
	return sb.String()
}

// GeneticDistribution represents a probability distribution over all possible genotypes.
type GeneticDistribution [256]uint64

func (gda GeneticDistribution) Breed(gdb GeneticDistribution) GeneticDistribution {
	var rslt GeneticDistribution

	// Breed each pair of possible genotypes into the result.
	for ga, pa := range gda {
		if pa == 0 {
			continue
		}
		ga := Genotype(ga)
		for gb, pb := range gdb {
			if pb == 0 {
				continue
			}
			gb := Genotype(gb)
			breedInto(&rslt, pa*pb, ga, gb)
		}
	}
	return rslt.Reduce()
}

func (gd GeneticDistribution) IsZero() bool {
	var zero GeneticDistribution
	return gd == zero
}

func (gd GeneticDistribution) Reduce() GeneticDistribution {
	rslt := gd
	if rslt.IsZero() {
		return rslt
	}

	g := rslt[0]
	for _, p := range rslt[1:] {
		if g == 1 {
			return rslt
		}
		g = gcd(g, p)
	}
	if g == 1 {
		return rslt
	}
	for i := range rslt {
		rslt[i] /= g
	}
	return rslt
}

func breedInto(gd *GeneticDistribution, weight uint64, ga, gb Genotype) {
	wt0 := punnetSquareLookupTable[ga.gene0()][gb.gene0()]
	wt1 := punnetSquareLookupTable[ga.gene1()][gb.gene1()]
	wt2 := punnetSquareLookupTable[ga.gene2()][gb.gene2()]
	wt3 := punnetSquareLookupTable[ga.gene3()][gb.gene3()]

	for g0, w0 := range wt0 {
		for g1, w1 := range wt1 {
			for g2, w2 := range wt2 {
				for g3, w3 := range wt3 {
					gd[g0|(g1<<2)|(g2<<4)|(g3<<6)] += weight * w0 * w1 * w2 * w3
				}
			}
		}
	}
}

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

//
// Lookup tables & other data only after this point.
//
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

	cosmos = mustSpecies("Cosmos", map[string]string{
		"rryyss": "White",
		"rryySs": "White",
		"rryySS": "White",
		"rrYyss": "Yellow",
		"rrYySs": "Yellow",
		"rrYySS": "White",
		"rrYYss": "Yellow",
		"rrYYSs": "Yellow",
		"rrYYSS": "Yellow",
		"Rryyss": "Pink",
		"RryySs": "Pink",
		"RryySS": "Pink",
		"RrYyss": "Orange",
		"RrYySs": "Orange",
		"RrYySS": "Pink",
		"RrYYss": "Orange",
		"RrYYSs": "Orange",
		"RrYYSS": "Orange",
		"RRyyss": "Red",
		"RRyySs": "Red",
		"RRyySS": "Red",
		"RRYyss": "Orange",
		"RRYySs": "Orange",
		"RRYySS": "Red",
		"RRYYss": "Black",
		"RRYYSs": "Black",
		"RRYYSS": "Red",
	})

	hyacinths = mustSpecies("Hyacinths", map[string]string{
		"rryyWW": "White",
		"rryyWw": "White",
		"rryyww": "Blue",
		"rrYyWW": "Yellow",
		"rrYyWw": "Yellow",
		"rrYyww": "White",
		"rrYYWW": "Yellow",
		"rrYYWw": "Yellow",
		"rrYYww": "Yellow",
		"RryyWW": "Red",
		"RryyWw": "Pink",
		"Rryyww": "White",
		"RrYyWW": "Orange",
		"RrYyWw": "Yellow",
		"RrYyww": "Yellow",
		"RrYYWW": "Orange",
		"RrYYWw": "Yellow",
		"RrYYww": "Yellow",
		"RRyyWW": "Red",
		"RRyyWw": "Red",
		"RRyyww": "Red",
		"RRYyWW": "Blue",
		"RRYyWw": "Red",
		"RRYyww": "Red",
		"RRYYWW": "Purple",
		"RRYYWw": "Purple",
		"RRYYww": "Purple",
	})

	lilies = mustSpecies("Lilies", map[string]string{
		"rryyss": "White",
		"rryySs": "White",
		"rryySS": "White",
		"rrYyss": "Yellow",
		"rrYySs": "White",
		"rrYySS": "White",
		"rrYYss": "Yellow",
		"rrYYSs": "Yellow",
		"rrYYSS": "White",
		"Rryyss": "Red",
		"RryySs": "Pink",
		"RryySS": "White",
		"RrYyss": "Orange",
		"RrYySs": "Yellow",
		"RrYySS": "Yellow",
		"RrYYss": "Orange",
		"RrYYSs": "Yellow",
		"RrYYSS": "Yellow",
		"RRyyss": "Black",
		"RRyySs": "Red",
		"RRyySS": "Pink",
		"RRYyss": "Black",
		"RRYySs": "Red",
		"RRYySS": "Pink",
		"RRYYss": "Orange",
		"RRYYSs": "Orange",
		"RRYYSS": "White",
	})

	mums = mustSpecies("Mums", map[string]string{
		"rryyWW": "White",
		"rryyWw": "White",
		"rryyww": "Purple",
		"rrYyWW": "Yellow",
		"rrYyWw": "Yellow",
		"rrYyww": "White",
		"rrYYWW": "Yellow",
		"rrYYWw": "Yellow",
		"rrYYww": "Yellow",
		"RryyWW": "Pink",
		"RryyWw": "Pink",
		"Rryyww": "Pink",
		"RrYyWW": "Yellow",
		"RrYyWw": "Red",
		"RrYyww": "Pink",
		"RrYYWW": "Purple",
		"RrYYWw": "Purple",
		"RrYYww": "Purple",
		"RRyyWW": "Red",
		"RRyyWw": "Red",
		"RRyyww": "Red",
		"RRYyWW": "Purple",
		"RRYyWw": "Purple",
		"RRYyww": "Red",
		"RRYYWW": "Green",
		"RRYYWw": "Green",
		"RRYYww": "Red",
	})

	pansies = mustSpecies("Pansies", map[string]string{
		"rryyWW": "White",
		"rryyWw": "White",
		"rryyww": "Blue",
		"rrYyWW": "Yellow",
		"rrYyWw": "Yellow",
		"rrYyww": "Blue",
		"rrYYWW": "Yellow",
		"rrYYWw": "Yellow",
		"rrYYww": "Yellow",
		"RryyWW": "Red",
		"RryyWw": "Red",
		"Rryyww": "Blue",
		"RrYyWW": "Orange",
		"RrYyWw": "Orange",
		"RrYyww": "Orange",
		"RrYYWW": "Yellow",
		"RrYYWw": "Yellow",
		"RrYYww": "Yellow",
		"RRyyWW": "Red",
		"RRyyWw": "Red",
		"RRyyww": "Purple",
		"RRYyWW": "Red",
		"RRYyWw": "Red",
		"RRYyww": "Purple",
		"RRYYWW": "Orange",
		"RRYYWw": "Orange",
		"RRYYww": "Purple",
	})

	roses = mustSpecies("Roses", map[string]string{
		"rryyWWss": "White",
		"rryyWWSs": "White",
		"rryyWWSS": "White",
		"rryyWwss": "White",
		"rryyWwSs": "White",
		"rryyWwSS": "White",
		"rryywwss": "Purple",
		"rryywwSs": "Purple",
		"rryywwSS": "Purple",
		"rrYyWWss": "Yellow",
		"rrYyWWSs": "Yellow",
		"rrYyWWSS": "Yellow",
		"rrYyWwss": "White",
		"rrYyWwSs": "White",
		"rrYyWwSS": "White",
		"rrYywwss": "Purple",
		"rrYywwSs": "Purple",
		"rrYywwSS": "Purple",
		"rrYYWWss": "Yellow",
		"rrYYWWSs": "Yellow",
		"rrYYWWSS": "Yellow",
		"rrYYWwss": "Yellow",
		"rrYYWwSs": "Yellow",
		"rrYYWwSS": "Yellow",
		"rrYYwwss": "White",
		"rrYYwwSs": "White",
		"rrYYwwSS": "White",
		"RryyWWss": "Red",
		"RryyWWSs": "Pink",
		"RryyWWSS": "White",
		"RryyWwss": "Red",
		"RryyWwSs": "Pink",
		"RryyWwSS": "White",
		"Rryywwss": "Red",
		"RryywwSs": "Pink",
		"RryywwSS": "Purple",
		"RrYyWWss": "Orange",
		"RrYyWWSs": "Yellow",
		"RrYyWWSS": "Yellow",
		"RrYyWwss": "Red",
		"RrYyWwSs": "Pink",
		"RrYyWwSS": "White",
		"RrYywwss": "Red",
		"RrYywwSs": "Pink",
		"RrYywwSS": "Purple",
		"RrYYWWss": "Orange",
		"RrYYWWSs": "Yellow",
		"RrYYWWSS": "Yellow",
		"RrYYWwss": "Orange",
		"RrYYWwSs": "Yellow",
		"RrYYWwSS": "Yellow",
		"RrYYwwss": "Red",
		"RrYYwwSs": "Pink",
		"RrYYwwSS": "White",
		"RRyyWWss": "Black",
		"RRyyWWSs": "Red",
		"RRyyWWSS": "Pink",
		"RRyyWwss": "Black",
		"RRyyWwSs": "Red",
		"RRyyWwSS": "Pink",
		"RRyywwss": "Black",
		"RRyywwSs": "Red",
		"RRyywwSS": "Pink",
		"RRYyWWss": "Orange",
		"RRYyWWSs": "Orange",
		"RRYyWWSS": "Yellow",
		"RRYyWwss": "Red",
		"RRYyWwSs": "Red",
		"RRYyWwSS": "White",
		"RRYywwss": "Black",
		"RRYywwSs": "Red",
		"RRYywwSS": "Purple",
		"RRYYWWss": "Orange",
		"RRYYWWSs": "Orange",
		"RRYYWWSS": "Yellow",
		"RRYYWwss": "Orange",
		"RRYYWwSs": "Orange",
		"RRYYWwSS": "Yellow",
		"RRYYwwss": "Blue",
		"RRYYwwSs": "Red",
		"RRYYwwSS": "White",
	})

	tulips = mustSpecies("Tulips", map[string]string{
		"rryyss": "White",
		"rryySs": "White",
		"rryySS": "White",
		"rrYyss": "Yellow",
		"rrYySs": "Yellow",
		"rrYySS": "White",
		"rrYYss": "Yellow",
		"rrYYSs": "Yellow",
		"rrYYSS": "Yellow",
		"Rryyss": "Red",
		"RryySs": "Pink",
		"RryySS": "White",
		"RrYyss": "Orange",
		"RrYySs": "Yellow",
		"RrYySS": "Yellow",
		"RrYYss": "Orange",
		"RrYYSs": "Yellow",
		"RrYYSS": "Yellow",
		"RRyyss": "Black",
		"RRyySs": "Red",
		"RRyySS": "Red",
		"RRYyss": "Black",
		"RRYySs": "Red",
		"RRYySS": "Red",
		"RRYYss": "Purple",
		"RRYYSs": "Purple",
		"RRYYSS": "Purple",
	})

	windflowers = mustSpecies("Windflowers", map[string]string{
		"rrooWW": "White",
		"rrooWw": "White",
		"rrooww": "Blue",
		"rrOoWW": "Orange",
		"rrOoWw": "Orange",
		"rrOoww": "Blue",
		"rrOOWW": "Orange",
		"rrOOWw": "Orange",
		"rrOOww": "Orange",
		"RrooWW": "Red",
		"RrooWw": "Red",
		"Rrooww": "Blue",
		"RrOoWW": "Pink",
		"RrOoWw": "Pink",
		"RrOoww": "Pink",
		"RrOOWW": "Orange",
		"RrOOWw": "Orange",
		"RrOOww": "Orange",
		"RRooWW": "Red",
		"RRooWw": "Red",
		"RRooww": "Purple",
		"RROoWW": "Red",
		"RROoWw": "Red",
		"RROoww": "Purple",
		"RROOWW": "Pink",
		"RROOWw": "Pink",
		"RROOww": "Purple",
	})
)
