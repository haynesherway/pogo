// Package pogo provides functions to get different types of stats for Pokemon Go
package pogo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

//const POKE_API = "http://pokeapi.co/api/v2/"
const HAYNESBOT_IMG = "https://github.com/haynesherway/pogo/blob/master/pics/"

// Locations of the json files
var (
	JSON_LOCATION = os.Getenv("GOPATH") + "/src/github.com/haynesherway/pogo/json/"
	POKEMON_FILE  = "/pokemon.json"
	MOVES_FILE    = "/move.json"
	TYPES_FILE    = "/type.json"
	ICONS_FILE    = "/home/pi/Public/Images/pokemon_icons/"
	ASSETS_FILE   = "/home/pi/Public/Images/PogoAssets/pokemon_icons/"
)

// Errors
var (
	ERR_NOT_FOUND = errors.New("Pokemon not found.")
)

var pokemonMap map[string]Pokemon
var dexMap map[int]string

// Pokemon is a resource representing a single pokemon
type Pokemon struct {
	Name  string       `json:"name"`
	ID    string       `json:"id"`
	Dex   int          `json:"dex"`
	Types TypeList     `json:"types"`
	Forms FormList     `json:"forms"`
	Stats PokemonStats `json:"stats"`
	Moves
	MaxCP int `json:"maxCP"`
	Icons
	TypeRelations
	API
}

var pokemonAliases = map[string]string{
	"mewtwo-a":         "mewtwo-armored%s",
	"-alola":           "alolan-%s",
	"castform-":        "%s-castform",
	"shellos-east-sea": "shellos-east%s",
	"shellos-west-sea": "shellos-west%s",
	"-normal":          "%s",
	"deoxys-":          "%s-deoxys",
	"galarian-":        "%s-galarian",
}

type Icons struct {
	Normal string
	Shiny  string
}

// Type API holds the pokemon data from the poke api
type API struct {
	url     string
	Sprites PokemonSprites `json:"sprites"`
}

// Type PokemonSprites is a representation of the sprites from poke api
type PokemonSprites struct {
	Front string `json:"front_default"`
}

type FormList []*PokemonForm

func (f FormList) Len() int {
	ct := 0
	for _, _ = range f {
		ct++
	}
	return ct
}

// PokemonForm is a resourse representing a pokemon form
type PokemonForm struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	AssetBundleValue  int    `json:"assetBundleValue"`
	AssetBundleSuffix string `json:"assetBundleSuffix"`
	Pokemon
}

// PokemonStats is a resource representing base stats for a pokemon
type PokemonStats struct {
	BaseStamina int `json:"baseStamina"`
	BaseAttack  int `json:"baseAttack"`
	BaseDefense int `json:"baseDefense"`
}

func (t TypeRelation) Print() string {
	str := []string{}
	for _, ty := range t {
		str = append(str, ty)
	}
	return strings.Join(str, ", ")
}

func (t TypeRelation) Len() int {
	ct := 0
	for _, _ = range t {
		ct++
	}
	return ct
}

// GetPokemon returns a Pokemon resource
func GetPokemon(pokemonName string) (*Pokemon, error) {
	// Check if a dex number was sent
	// pokemonName =  strings.Replace(pokemonName, "-", " ", 1)
	if dex, err := strconv.Atoi(pokemonName); err == nil {
		if pk, ok := dexMap[dex]; ok {
			pokemonName = pk
		}
	}

	pokemonName = strings.ToLower(pokemonName)
	if p, ok := pokemonMap[pokemonName]; ok {
		p.GetSprite()
		p.GetTypeEffects()
		return &p, nil
	} else {
		return nil, ERR_NOT_FOUND
	}
}

func (p *Pokemon) GetSprite() {
	if p.API.Sprites.Front == "" {
		p.API.Sprites.Front = HAYNESBOT_IMG + p.ID + ".png?raw=true"
		/*err := p.API.getSprite()
		if err != nil {
			fmt.Println(err.Error())
		}*/
	}
	return
}

func (p *Pokemon) GetShiny() (string, error) {
	if p.Icons.Shiny == "" {
		err := p.getShiny()
		if err != nil {
			log.Println(err.Error())
			return "", err
		}
	}
	return p.Icons.Shiny, nil
}

func (p *Pokemon) GetNormal() (string, error) {
	if p.Icons.Normal == "" {
		err := p.getNormal()
		if err != nil {
			log.Println(err.Error())
			return "", err
		}
	}
	return p.Icons.Normal, nil
}

// GetMaxCP returns the maximum CP of a pokemon
func (p *Pokemon) GetMaxCP() (cp int) {
	return p.MaxCP
}

func (p *Pokemon) GetCP(level float64, ivAttack int, ivDefense int, ivStamina int) (cp int) {
	attack := getStatValue(p.Stats.BaseAttack, ivAttack, level)
	defense := getStatValue(p.Stats.BaseDefense, ivDefense, level)
	stamina := getStatValue(p.Stats.BaseStamina, ivStamina, level)

	cp = calculateCP(attack, defense, stamina, level)
	return
}

func (p *Pokemon) GetHP(level float64, ivStamina int) (hp int) {
	stamina := getStatValue(p.Stats.BaseStamina, ivStamina, level)
	hp = calculateHP(stamina, level)
	return
}

func (p *Pokemon) GetRaidCPChart() ([]IVStat, string) {
	possibleIVs := []int{15, 14, 13, 12, 11, 10}

	//ivList := map[int]map[int]map[int]
	ivs := []IVStat{}

	str := "[ % ]Ak|Df|St[ 20 | 25 ]\n"
	str += "------------------------\n"
	for _, a := range possibleIVs {
		for _, d := range possibleIVs {
			for _, s := range possibleIVs {
				percent := round(float64((a+d+s)*100) / float64(45))

				cp15 := p.GetCP(15.0, a, d, s)
				cp20 := p.GetCP(20.0, a, d, s)
				cp25 := p.GetCP(25.0, a, d, s)
				iv := IVStat{
					Attack:  a,
					Defense: d,
					Stamina: s,
					CP15:    cp15,
					CP20:    cp20,
					CP25:    cp25,
					Percent: percent,
				}
				ivs = append(ivs, iv)
			}
		}
	}

	ivs = SortChart(ivs)
	chart := []string{}
	for _, iv := range ivs {
		chart = append(chart, iv.PrintChartRow())
	}

	return ivs, str + strings.Join(chart, "\n")
}

func (p *Pokemon) GetRaidCPRange() string {
	min20 := p.GetCP(20.0, 10, 10, 10)
	max20 := p.GetCP(20.0, 15, 15, 15)
	min25 := p.GetCP(25.0, 10, 10, 10)
	max25 := p.GetCP(25.0, 15, 15, 15)
	return fmt.Sprintf("Level 20: %v - **%v**\nLevel 25: %v - **%v**", min20, max20, min25, max25)
}

func (p *Pokemon) GetIV(cp int, hp int, level float64, stardust int, best string) ([]IVStat, string) {
	IVStat := &IVStat{
		Level:    level,
		CP:       cp,
		HP:       hp,
		Stardust: stardust,
		Best:     best,
	}
	return p.getIV(IVStat)
}

func (p *Pokemon) getIV(stats *IVStat) ([]IVStat, string) {
	possibleIVs := []int{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}

	possibleLevels := []float64{}
	if stats.Level != 0.0 {
		possibleLevels = append(possibleLevels, stats.Level)
	} else if stats.Stardust != 0 {
		if _, ok := stardustMap[stats.Stardust]; ok {
			possibleLevels = stardustMap[stats.Stardust]
		} else {
			for k := range multiplierMap {
				possibleLevels = append(possibleLevels, k)
			}
		}
	} else {
		for k := range multiplierMap {
			possibleLevels = append(possibleLevels, k)
		}
	}
	cp := stats.CP
	hp := stats.HP

	ivList := []IVStat{}

	message := "|Lvl | At | Df | St |%%%|   \n"
	message += "|----|----|----|----|---|   \n"

	for _, l := range possibleLevels {
		for _, a := range possibleIVs {
			for _, d := range possibleIVs {
				for _, s := range possibleIVs {
					calccp := p.GetCP(l, a, d, s)
					calchp := p.GetHP(l, s)

					if stats.Best != "" {
						beststr := ""
						vals := []int{a, d, s}
						sort.Ints(vals)
						highest := vals[2]
						if a == highest {
							beststr += "a"
						}
						if d == highest {
							beststr += "d"
						}
						if s == highest {
							beststr += "s"
						}
						if beststr != stats.Best {
							continue
						}
					}
					if cp == calccp && hp == calchp {
						perc := round(float64((a+d+s)*100) / float64(45))
						stat := IVStat{
							Level:   l,
							Attack:  a,
							Defense: d,
							Stamina: s,
							//CP20: cp20,
							//CP25: cp25,
							Percent: perc,
						}
						ivList = append(ivList, stat)
					}
				}
			}
		}
	}

	if ivList == nil || len(ivList) == 0 {
		return nil, ""
	}

	ivList = SortChart(ivList)
	chart := []string{}
	/*if len(ivList) > 50 {
	    ivList = ivList[0:50]
	}*/
	for i, s := range ivList {
		if i >= 30 {
			break
		}
		chart = append(chart, s.PrintIVRow())
	}

	return ivList, message + strings.Join(chart, "\n")
}

func (p *Pokemon) GetRaidIV(raidcp int) ([]IVStat, string) {
	possibleIVs := []int{15, 14, 13, 12, 11, 10}

	ivList := []IVStat{}

	//message := fmt.Sprintf("Possible IVs for **%s** with CP of **%d**:\n", p.Name, raidcp)

	message := "| At | Df | St | %%% | \n"
	message += "|----|----|----|-----| \n"
	for _, a := range possibleIVs {
		for _, d := range possibleIVs {
			for _, s := range possibleIVs {
				cp15 := p.GetCP(15.0, a, d, s)
				cp20 := p.GetCP(20.0, a, d, s)
				cp25 := p.GetCP(25.0, a, d, s)
				if raidcp == cp20 || raidcp == cp25 || raidcp == cp15 {
					lvl := 20.0
					if raidcp == cp25 {
						lvl = 25.0
					} else if raidcp == cp15 {
						lvl = 15.0
					}
					perc := round(float64((a+d+s)*100) / float64(45))
					stat := IVStat{
						Attack:  a,
						Defense: d,
						Stamina: s,
						Level:   lvl,
						//CP20: cp20,
						//CP25: cp25,
						Percent: perc,
					}
					ivList = append(ivList, stat)
				}
			}
		}
	}

	if ivList == nil || len(ivList) == 0 {
		return ivList, ""
	}

	ivList = SortChart(ivList)
	chart := []string{}
	for _, s := range ivList {
		chart = append(chart, s.PrintRaidIVRow())
	}

	return ivList, message + strings.Join(chart, "\n")
}

func (p *Pokemon) GetTypeRelations() (relations map[string]map[string]float64) {
	relations = make(map[string]map[string]float64)
	relations["attack"] = make(map[string]float64)
	relations["defense"] = make(map[string]float64)

	for _, pt := range p.Types {
		attackScalars := GetAttackTypeScalars(pt.ID)
		for tName, tScalar := range attackScalars {
			if _, ok := relations["attack"][tName]; !ok {
				relations["attack"][tName] = 1
			}
			relations["attack"][tName] = relations["attack"][tName] * tScalar
		}

		defenseScalars := GetDefenseTypeScalars(pt.ID)
		for tName, tScalar := range defenseScalars {
			if _, ok := relations["defense"][tName]; !ok {
				relations["defense"][tName] = 1
			}
			relations["defense"][tName] = relations["defense"][tName] * tScalar
		}
	}

	return
}

func (p *Pokemon) GetTypeEffects() {
	if p.SuperEffective.Len() != 0 {
		return
	}

	typeRelations := p.GetTypeRelations()

	//p.SuperEffective = []TypeRelation{}
	//p.NotEffective = []TypeRelationstring{}
	//p.Weakness = []TypeRelationstring{}
	//p.Resistance = TypeRelation}

	//Attack
	for ty, sc := range typeRelations["attack"] {
		if sc > 1.9 {
			p.SuperEffective = append(p.SuperEffective, ty+"(x2)")
		} else if sc >= 1.4 {
			p.SuperEffective = append(p.SuperEffective, ty)
		} else if sc <= .6 {
			p.NotEffective = append(p.NotEffective, ty+"(x2)")
		} else if sc <= .8 {
			p.NotEffective = append(p.NotEffective, ty)
		}
	}

	//Defense
	for ty, sc := range typeRelations["defense"] {
		if sc > 1.9 {
			p.Weakness = append(p.Weakness, ty+"(x2)")
		} else if sc >= 1.4 {
			p.Weakness = append(p.Weakness, ty)
		} else if sc <= .6 {
			p.Resistance = append(p.Resistance, ty+"(x2)")
		} else if sc <= .8 {
			p.Resistance = append(p.Resistance, ty)
		}
	}

	/*
	   msg := fmt.Sprintf("Type Effects for **%s** (%s):\n", p.Name, p.Types.Print())
	   /*if len(doubleEffective) > 0 {
	       msg += fmt.Sprintf("Double Effective Against: %s\n", strings.Join(doubleEffective, ", "))
	   }
	   if len(superEffective) > 0 {
	       msg += fmt.Sprintf("Super Effective Against: %s\n", strings.Join(superEffective, ", "))
	   }
	   if len(notEffective) > 0 {
	       msg += fmt.Sprintf("Not Very Effective Against: %s\n", strings.Join(notEffective, ", "))
	   }
	   if len(weakness) > 0 {
	       msg += fmt.Sprintf("Weak To: %s\n", strings.Join(weakness, ", "))
	   }
	   if len(resistance) > 0 {
	       msg += fmt.Sprintf("Resistant To: %s\n", strings.Join(resistance, ", "))
	   }*/

	return
}

/* for _, pt := range Pokemon.Types {
    if ty, ok := TypeMap[pt.ID]; ok {
        for _, t := range ty {
            scalars := p.GetTypeScalars()
            for tName, tScalar := range scalars {
                if _, ok := relations[tName]; !ok {
                    relations[tName] = 1
                }
                relations[t.Name] = relations[t.Name] *
            }

        }
    }
}*/

/*func PrintPokemonToDiscord(s *discordgo.Session, m *discordgo.MessageCreate, fields []string) error {
    if len(fields) < 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pokemon command should be in the following format: !poke mewtwo")

		return nil
	}

	pokemonName := strings.ToLower(fields[1])

	p, err := gokemon.GetPokemon(pokemonName)
	fmt.Println(err)
	fmt.Println(p.String())
	if err != nil {
	    _, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Pokemon %s not recognized", fields[1]))
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, p.String())

	return nil
}

func PrintWeaknessToDiscord(s *discordgo.Session, m *discordgo.MessageCreate, fields []string) error {
    /*if len(fields) < 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Weakness command should be in the following format: !weakness mewtwo")

		return nil
	}

	pokemonName := strings.ToLower(fields[1])

	if p, err := gokemon.GetPokemon(pokemonName); err != nil {
	    types := p.Types
	    message := fmt.Sprintf("Weaknesses for **%s:** %s", p.Name, weaknessString)
	    _, _ = s.ChannelMessageSend(m.ChannelID, message)
	} else {
	    _, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Pokemon %s not recognized", fields[1]))
	}

	return nil
} */

// ImageExists checks if an image exists in the image server folder
func ImageExists(name string) bool {
	if _, err := os.Stat(ICONS_FILE + name); err != nil {
		return false
	}
	return true
}

func AssetsImageExists(name string) bool {
	if _, err := os.Stat(ASSETS_FILE + name); err != nil {
		return false
	}
	return true
}

func RenameImage(old string, nw string) {
	if err := os.Rename(ICONS_FILE+old, ICONS_FILE+nw); err != nil {
		fmt.Println(err.Error())
	}
	return
}

func LinkFromAssets(old string, nw string) {
	if err := os.Link(ASSETS_FILE+old, ICONS_FILE+nw); err != nil {
		fmt.Println(err.Error())
	}
	return
}

func ImageLocation(name string) string {
	return ICONS_FILE + name
}

func init() {

	pokemonMap = make(map[string]Pokemon)
	dexMap = make(map[int]string)

	//Pokemon
	file, err := ioutil.ReadFile(JSON_LOCATION + POKEMON_FILE)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	writeFile, err := os.Create("pokemonNames.csv")
	defer writeFile.Close()

	pokemonList := []Pokemon{}
	err = json.Unmarshal(file, &pokemonList)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, poke := range pokemonList {

		if strings.Contains(strings.ToUpper(poke.ID), "PURIFIED") || strings.Contains(strings.ToUpper(poke.ID), "SHADOW") {
			continue
		}

		pokeID := strings.Replace(strings.ToLower(poke.ID), "_", "-", -1)
		poke.ID = pokeID
		pokemonMap[pokeID] = poke
		dexMap[poke.Dex] = pokeID

		// See if pokemon image exists
		if !ImageExists(fmt.Sprintf("%s.png", pokeID)) {
			fmt.Println("Don't have:", pokeID)
			// Try to get from pogo assets
			oldImage := fmt.Sprintf("pokemon_icon_%03d_%02d.png", poke.Dex, 0)

			newImage := fmt.Sprintf("%s.png", pokeID)
			if AssetsImageExists(oldImage) {
				LinkFromAssets(oldImage, newImage)
			} else {
			ImageLoop:
				for f := 0; f < 20; f++ {
					oldImage := fmt.Sprintf("pokemon_icon_%03d_%02d.png", poke.Dex, f)
					if AssetsImageExists(oldImage) {
						LinkFromAssets(oldImage, newImage)
						break ImageLoop
					}
				}
			}
		}
		if !ImageExists(fmt.Sprintf("%s-shiny.png", pokeID)) {
			oldImage := fmt.Sprintf("pokemon_icon_%03d_%02d_shiny.png", poke.Dex, 0)
			newImage := fmt.Sprintf("%s-shiny.png", pokeID)
			if AssetsImageExists(oldImage) {
				LinkFromAssets(oldImage, newImage)
			} else {
			ShinyImageLoop:
				for f := 0; f < 20; f++ {
					oldImage = fmt.Sprintf("pokemon_icon_%03d_%02d_shiny.png", poke.Dex, f)
					if AssetsImageExists(oldImage) {
						LinkFromAssets(oldImage, newImage)
						break ShinyImageLoop
					}
				}
			}
		}

		if poke.Forms.Len() > 0 {
			for _, form := range poke.Forms {
				thisForm := poke

				// See if any changes to name need to be done
				if len(form.ID) > 6 && form.ID[0:6] == "SPINDA" {
					lastNum, _ := strconv.Atoi(form.ID[8:9])
					lastNum++
					form.ID = fmt.Sprintf("SPINDA_%d", lastNum)
					form.Name = fmt.Sprintf("Spinda %d", lastNum)
				}
				formID := strings.Replace(strings.ToLower(form.ID), "_", "-", -1)
				thisForm.ID = formID
				thisForm.Name = form.Name
				//thisForm.Name = strings.Title(strings.ToLower(strings.Replace(form.ID, "_", " ", -1)))
				if _, ok := pokemonMap[formID]; !ok {
					pokemonMap[formID] = thisForm
				}
				// See if pokemon image exists
				if !ImageExists(fmt.Sprintf("%s.png", formID)) {
					// Try to get from pogo assets
					f := form.AssetBundleValue
					suffix := form.AssetBundleSuffix
					oldImage := ""
					if suffix == "" {
						oldImage = fmt.Sprintf("pokemon_icon_%03d_%02d.png", poke.Dex, f)
					} else {
						oldImage = fmt.Sprintf("pokemon_icon%s.png", suffix)
					}
					newImage := fmt.Sprintf("%s.png", formID)
					if AssetsImageExists(oldImage) {
						LinkFromAssets(oldImage, newImage)
					}
				}
				if !ImageExists(fmt.Sprintf("%s-shiny.png", formID)) {
					f := form.AssetBundleValue
					suffix := form.AssetBundleSuffix
					oldImage := ""
					if suffix == "" {
						oldImage = fmt.Sprintf("pokemon_icon_%03d_%02d_shiny.png", poke.Dex, f)
					} else {
						oldImage = fmt.Sprintf("pokemon_icon%s_shiny.png", suffix)
					}
					newImage := fmt.Sprintf("%s-shiny.png", formID)
					if AssetsImageExists(oldImage) {
						LinkFromAssets(oldImage, newImage)
					}
				}
			}
		}
	}

	for id, pokemon := range pokemonMap {
		img := ""
		if err := pokemon.getNormal(); err != nil {
			img = "No image found"
		}
		_, err := writeFile.WriteString(fmt.Sprintf("%v,%s,%s,%s\n", pokemon.Dex, pokemon.Name, id, img))
		if err != nil {
			fmt.Println(err)
		}
	}

	// Add aliases
	for p, alias := range pokemonAliases {
		for poke, data := range pokemonMap {
			if strings.Contains(poke, p) {
				newalias := fmt.Sprintf(alias, strings.Replace(poke, p, "", -1))
				pokemonMap[newalias] = data

			}
		}
	}

	return
}

func (p *API) getSprite() error {
	resp, err := http.Get(p.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(responseData, p)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pokemon) getNormal() error {
	fileName := fmt.Sprintf("%s.png", p.ID)
	if ImageExists(fileName) {
		p.Icons.Normal = ImageLocation(fileName)
	} else {
		return errors.New("no image found")
	}
	return nil
}
func (p *Pokemon) getShiny() error {
	fileName := fmt.Sprintf("%s-shiny.png", p.ID)
	if ImageExists(fileName) {
		p.Icons.Shiny = ImageLocation(fileName)
	} else {
		return errors.New("no image found")
	}
	return nil
}
