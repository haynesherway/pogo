package pogo

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "strings"
)

var typeMap = map[string]Type{}
var typeToID = map[string]string{}

var (
    ERR_TYPE_NOT_FOUND = errors.New("Type not found.")
)

type Type struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Damage []*TypeDamage `json:"damage"`
    Thumbnail string
    TypeRelations
}

type TypeList []*PokemonType

type PokemonType struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
}

type TypeDamage struct {
    ID string `json:"id"`
    Scalar float64 `json:"attackScalar"`
}

type TypeRelations struct {
    SuperEffective TypeRelation
    NotEffective TypeRelation
    Weakness TypeRelation
    Resistance TypeRelation
}

type TypeRelation []string

func GetType(t string) (*Type, error) {
    t = strings.ToLower(t)
    if ty, ok := typeMap[typeToID[t]]; ok {
        ty.GetTypeEffects()
        return &ty, nil
    } else {
        return &Type{}, ERR_TYPE_NOT_FOUND
    }
}

func (typeList TypeList) Print() string {
    types := []string{}
    for _, t := range typeList {
        types = append(types, t.Name)
    }
    return strings.Join(types, ", ")
}
func (t *Type) GetTypeEffects() {
        if t.SuperEffective.Len() != 0 {
            return
        }
    
        typeRelations := make(map[string]map[string]float64)
        typeRelations["attack"] = GetAttackTypeScalars(t.ID)
        typeRelations["defense"] = GetDefenseTypeScalars(t.ID)
        
        //Attack
        for ty, sc := range typeRelations["attack"] {
             if sc > 1.9 {
                t.SuperEffective = append(t.SuperEffective, ty + "(x2)")  
            } else if sc >= 1.4 {
                t.SuperEffective = append(t.SuperEffective, ty)
            } else if sc <= .6 {
                t.NotEffective = append(t.NotEffective, ty + "(x2)")
            } else if sc <= .8 {
                t.NotEffective = append(t.NotEffective, ty)
            }
        }
        
        //Defense 
        for ty, sc := range typeRelations["defense"] {
           if sc > 1.9 {
                t.Weakness = append(t.Weakness, ty + "(x2)")
            } else if sc >= 1.4 {
                t.Weakness = append(t.Weakness, ty)
            } else if sc <= .6 {
                t.Resistance = append(t.Resistance, ty + "(x2)")
            } else if sc <= .8 {
                t.Resistance = append(t.Resistance, ty)
            }
        }
    
        return
    }

func GetAttackTypeScalars(id string) (map[string]float64) {
    typeScalars := map[string]float64{}
    if ty, ok := typeMap[id]; ok {
    
        for _, damage := range ty.Damage {
            typeScalars[typeMap[damage.ID].Name] = damage.Scalar
        }
    }
    
    return typeScalars
}

func GetDefenseTypeScalars(id string) (map[string]float64) {
    if _, ok := typeMap[id]; !ok {
        return nil
    } 
    
    typeScalars := map[string]float64{}
    for _, ty := range typeMap {
        for _, typeDamage := range ty.Damage {
            if typeDamage.ID == id {
                typeScalars[ty.Name] = typeDamage.Scalar
            }
        }
    }
    
    return typeScalars
}

func init() {
    typeMap = make(map[string]Type)
    
    //Types
	file, err := ioutil.ReadFile(JSON_LOCATION+TYPES_FILE)
	if err != nil {
	    fmt.Println(err.Error())
	    return
	}
	
	typeList := []Type{}
	err = json.Unmarshal(file, &typeList)
	if err != nil {
	    fmt.Println(err.Error())
	    return
	}
	
	for _, ty := range typeList {
	    ty.Thumbnail = "https://github.com/haynesherway/pogo/blob/master/pics/"+strings.ToLower(ty.Name)+".png?raw=true"
	    //ty.Thumbnail = "https://github.com/PokeAPI/sprites/blob/master/sprites/items/"+strings.ToLower(ty.Name)+"-gem.png?raw=true"
	    typeMap[ty.ID] = ty
	    typeToID[strings.ToLower(ty.Name)] = ty.ID
	}
}