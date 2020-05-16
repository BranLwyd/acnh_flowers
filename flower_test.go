package flower

import (
	"fmt"
	"testing"
)

func TestGenotypeParsing(t *testing.T) {
	gs, err := NewGenotypeSerdeFromExample("XxYyZz")
	if err != nil {
		t.Fatalf("Could not create genotype serializer: %v", err)
	}

	for _, g0 := range []string{"xx", "Xx", "XX"} {
		for _, g1 := range []string{"yy", "Yy", "YY"} {
			for _, g2 := range []string{"zz", "Zz", "ZZ"} {
				genotype := fmt.Sprintf("%s%s%s", g0, g1, g2)
				t.Run(genotype, func(t *testing.T) {
					g, err := gs.ParseGenotype(genotype)
					if err != nil {
						t.Fatalf("ParseGenotype got unexpected error: %v", err)
					}
					got := gs.RenderGenotype(g)
					if got != genotype {
						t.Errorf("RenderGenotype(ParseGenotype(%q)) = %q, want %q", genotype, got, genotype)
					}
				})
			}
		}
	}
}
