package pogo

import (
	"fmt"
	"testing"
)

type testMon struct {
	input string
	name string
}

var testMons = []testMon{
	{"rayquaza", "Rayquaza"},
	{"Regirock", "Regirock"},
	{"mewtwo", "Mewtwo"},
	{"weedle", "Weedle"},
}

func TestGetPokemon(t *testing.T) {
	for _, pokemon := range testMons {
		p, err := GetPokemon(pokemon.name)
		if err != nil {
			t.Error(
			"For", pokemon.input,
			"expected", pokemon.name,
			"got", err.Error())
		}
		if pokemon.name != p.Name {
			t.Error(
			"For", pokemon.input,
			"expected", pokemon.name,
			"got", p.Name)
		}
	}
}

func ExampleGetPokemon() {
	exampleMons := []string{"rayquaza", "Regirock", "mewtwo", "weedle", "nokemon"}
	for _, pokemonName := range exampleMons {
		p, err := GetPokemon(pokemonName)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(p.Name)
		}
	}
	// Output:
	// rayquaza
	// Regirock
	// mewtwo
	// weedle
	// Pokemon not found.
}