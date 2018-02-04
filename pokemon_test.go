package pogo

import (
	"fmt"
	"testing"
)

type testMon struct {
	input string
	name string
	maxcp int
}

var testMons = []testMon{
	{"rayquaza", "Rayquaza", 3645},
	{"Regirock", "Regirock", 3087},
	{"mewtwo", "Mewtwo", 3982},
	{"weedle", "Weedle", 397},
}

func TestGetPokemon(t *testing.T) {
	for _, pokemon := range testMons {
		p, err := GetPokemon(pokemon.input)
		if err != nil {
			t.Error(
			"For", pokemon.input,
			"expected", pokemon.name,
			"got", err.Error())
			return
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
	// Rayquaza
	// Regirock
	// Mewtwo
	// Weedle
	// Pokemon not found.
}

func TestPokemon_GetMaxCP(t *testing.T) {
	for _, pokemon := range testMons {
		p, err := GetPokemon(pokemon.input)
		if err != nil {
			t.Error("Unable to get pokemon")
			return
		}
		maxcp := p.GetMaxCP()
		if maxcp != pokemon.maxcp {
			t.Error(
			"For", pokemon.input,
			"expected", pokemon.maxcp,
			"got", maxcp)
		}
	}
}

func ExamplePokemon_GetMaxCP() {
	pokemon, err := GetPokemon("weedle")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Max CP for Weedle is", pokemon.GetMaxCP())
	}
	// Output:
	// Max CP for Weedle is 397
}

func ExamplePokemon_GetCP() {
	pokemon, err := GetPokemon("groudon")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		cp := pokemon.GetCP(20, 15, 14, 15)
		fmt.Println("CP for Groudon at level 20 with 15/14/15 IVs is", cp)
	}
	// Output:
	// CP for Groudon at level 20 with 15/14/15 IVs is 2323
}