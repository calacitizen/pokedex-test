package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/calacitizen/pokedexcli/internal/pokecache"
)

const LIMIT = 20
const OFFSET = 20

const MAIN_URL = "https://pokeapi.co/api/v2/location-area/"
const POKEMON_URL = "https://pokeapi.co/api/v2/pokemon/"

var defaultU = fmt.Sprintf("%s?limit=%d", MAIN_URL, LIMIT)

var prev *string
var next *string
var ff bool = true

type Location struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type requestBody struct {
	Count    int        `json:"count"`
	Next     *string    `json:"next"`
	Previous *string    `json:"previous"`
	Results  []Location `json:"results"`
}

type Poke struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PokeRes struct {
	Pokemon Poke `json:"pokemon"`
}

type requestSingleBody struct {
	PokemonEncounters []PokeRes `json:"pokemon_encounters"`
}

type StatInner struct {
	Name string `json:"name"`
}

type PokemonStat struct {
	BaseStat int       `json:"base_stat"`
	Effort   int       `json:"effort"`
	Stat     StatInner `json:"stat"`
}

type PokemonType struct {
	Type StatInner `json:"type"`
}

type Pokemon struct {
	Id             int           `json:"id"`
	Name           string        `json:"name"`
	BaseExperience int           `json:"base_experience"`
	Height         int           `json:"height"`
	Weight         int           `json:"weight"`
	Stats          []PokemonStat `json:"stats"`
	Types          []PokemonType `json:"types"`
}

func getData(u string, ch pokecache.Cache) ([]byte, error) {
	req, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	var data []byte
	if val, exists := ch.Get(u); exists {
		data = val
	} else {
		d, errB := io.ReadAll(req.Body)
		if errB != nil {
			return nil, errB
		}
		ch.Add(u, d)
		data = d
	}
	return data, nil
}

func GetLocations(u string, ch pokecache.Cache) (*requestBody, error) {
	data, err := getData(u, ch)
	if err != nil {
		return nil, err
	}
	rBody := requestBody{}
	errL := json.Unmarshal(data, &rBody)
	if errL != nil {
		return nil, errL
	}
	return &rBody, nil
}

func GetPrev(ch pokecache.Cache) (*requestBody, error) {
	if prev == nil {
		return nil, errors.New("Can't fetch prev list!")
	}
	ls, err := GetLocations(*prev, ch)
	if err != nil {
		return nil, err
	}
	prev = ls.Previous
	next = ls.Next
	return ls, nil
}

func GetNext(ch pokecache.Cache) (*requestBody, error) {
	var fUrl string
	if next == nil && ff == false {
		return nil, errors.New("Can't fetch next list!")
	}
	if ff == true {
		fUrl = defaultU
		ff = false
	} else {
		if next == nil {
			return nil, errors.New("No next page on the list!")
		}
		fUrl = *next
	}
	ls, err := GetLocations(fUrl, ch)
	if err != nil {
		return nil, err
	}
	prev = ls.Previous
	next = ls.Next
	return ls, nil
}

func GetPokemons(locName string, ch pokecache.Cache) (*[]PokeRes, error) {
	data, err := getData(MAIN_URL+locName, ch)
	if err != nil {
		return nil, err
	}
	rBody := requestSingleBody{}
	errL := json.Unmarshal(data, &rBody)
	if errL != nil {
		return nil, errL
	}
	return &rBody.PokemonEncounters, nil
}

func GetPokemon(pokName string, ch pokecache.Cache) (*Pokemon, error) {
	data, err := getData(POKEMON_URL+pokName, ch)
	if err != nil {
		return nil, err
	}
	rBody := Pokemon{}
	errP := json.Unmarshal(data, &rBody)
	if errP != nil {
		return nil, errP
	}
	return &rBody, nil
}
