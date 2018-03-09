package gp

import (
	"testing"
)

func TestNewIndividuoRand(t *testing.T) {
	ind := NewIndividuoRand(10)
	if len(ind.Genes) != 10 {
		t.Error("Expected lenght of 10")
	}
}

func TestIndividuoString(t *testing.T) {
	ind := Individuo{}
	ind.Genes = []int{1, 2, 3, 4, 5}
	str := ind.String()

	if str != "(0.00,5)->[1 2 3 4 5 ]" {
		t.Error("String not equals expected")
	}
}
