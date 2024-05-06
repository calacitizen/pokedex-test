package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/calacitizen/pokedexcli/internal/api"
	"github.com/calacitizen/pokedexcli/internal/pokecache"
)

var CaughtPokemons = make(map[string]api.Pokemon, 0)

var cln = "Pokedex"

var ch = pokecache.NewCache(7 * time.Second)

func pP() {
	fmt.Print(cln, " > ")
}

func pU(text string) {
	fmt.Print(text, ": command not found")
}

type cliCommand struct {
	name        string
	description string
	callback    func(*string) error
}

func displayHelp(s *string) error {
	fmt.Printf("\nWelcome to the %v!\n", cln)
	fmt.Println("Usage:")
	fmt.Println()
	for _, cm := range ccm() {
		fmt.Printf("%s: %s\n", cm.name, cm.description)
	}
	fmt.Println()
	return nil
}

func exitProgramm(s *string) error {
	os.Exit(0)
	return nil
}

func showLocations(results []api.Location) {
	for _, l := range results {
		fmt.Println(l.Name)
	}
}

func showNext(s *string) error {
	body, err := api.GetNext(ch)
	if err != nil {
		return err
	}
	showLocations(body.Results)
	return nil
}

func showPrev(s *string) error {
	body, err := api.GetPrev(ch)
	if err != nil {
		return err
	}
	showLocations(body.Results)
	return nil
}

func cleanInput(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func showPokemons(s *string) error {
	if s == nil {
		return errors.New("No argument of location provided!")
	}
	fmt.Printf("Exploring %s...\n", *s)
	res, err := api.GetPokemons(*s, ch)
	if err != nil {
		return err
	}
	fmt.Println("Found Pokemon:")
	for _, p := range *res {
		fmt.Printf(" - %s\n", p.Pokemon.Name)
	}
	return nil
}

func attemptCatch(xp int) bool {
	// Difficulty based on base experience (Higher XP = harder to catch)
	catchDifficulty := xp / 25 // Arbitrary scaling factor

	// Generate random number for catch chance
	catchChance := rand.Intn(100) + 1 // 1-100

	return catchChance <= (100 - catchDifficulty)
}

func catchPokemon(s *string) error {
	if s == nil {
		return errors.New("No pokemon name provided provided!")
	}
	fmt.Printf("Throwing a Pokeball at  %s...\n", *s)
	res, err := api.GetPokemon(*s, ch)
	if err != nil {
		return err
	}
	catchSuccess := attemptCatch(res.BaseExperience)
	if catchSuccess {
		CaughtPokemons[*s] = *res
		fmt.Printf("%s was caught!\n", res.Name)
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		fmt.Printf("%s escaped!\n", res.Name)
	}

	return nil
}

func inspectPokemon(s *string) error {
	if s == nil {
		return errors.New("No pokemon name provided provided!")
	}

	if pok, exists := CaughtPokemons[*s]; exists {
		fmt.Printf("Name: %s\n", pok.Name)
		fmt.Printf("Height: %d\n", pok.Height)
		fmt.Printf("Weight: %d\n", pok.Weight)
		fmt.Println("Stats:")
		for _, s := range pok.Stats {
			fmt.Printf("  -%s: %v\n", s.Stat.Name, s.BaseStat)
		}
		fmt.Println("Types:")
		for _, t := range pok.Types {
			fmt.Printf("  - %s\n", t.Type.Name)
		}
	} else {
		fmt.Println("you have not caught that pokemon")
	}

	return nil

}

func pokedexCmd(s *string) error {
	fmt.Println("Your Pokedex:")
	for _, p := range CaughtPokemons {
		fmt.Printf(" - %s\n", p.Name)
	}
	return nil
}

func ccm() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    displayHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    exitProgramm,
		},
		"map": {
			name:        "Map",
			description: "The map command displays the names of 20 location areas in the Pokemon world. Each subsequent call to map should display the next 20 locations, and so on. The idea is that the map command will let us explore the world of Pokemon.",
			callback:    showNext,
		},
		"mapb": {
			name:        "Map Back",
			description: "The map command displays the names of 20 location areas in the Pokemon world. Each subsequent call to map should display the next 20 locations, and so on. The idea is that the map command will let us explore the world of Pokemon.",
			callback:    showPrev,
		},
		"explore": {
			name:        "Explore",
			description: "Get pokemons from location area",
			callback:    showPokemons,
		},
		"catch": {
			name:        "Catch",
			description: "Try to catch pokemon!",
			callback:    catchPokemon,
		},
		"inspect": {
			name:        "Inspect",
			description: "Inspect caught pokemons!",
			callback:    inspectPokemon,
		},
		"pokedex": {
			name:        "Pokedex",
			description: "Show all caught pokemons",
			callback:    pokedexCmd,
		},
	}
}

func printUnknown(text string) {
	fmt.Println(text, ": command not found")
}

func handleInvalidCmd(text string) {
	defer printUnknown(text)
}

func handleCmd(text string) {
	handleInvalidCmd(text)
}

func main() {
	commands := ccm()
	reader := bufio.NewScanner(os.Stdin)
	pP()
	for reader.Scan() {
		text := cleanInput(reader.Text())
		words := strings.Fields(text)
		var cmName string
		if len(words) > 1 {
			cmName = words[0]
		} else {
			cmName = text
		}
		if cm, exists := commands[cmName]; exists {
			var err error
			if len(words) > 1 {
				err = cm.callback(&words[1])
			} else {
				err = cm.callback(nil)
			}
			if err != nil {
				fmt.Printf("Error occured while running command. %v", err)
				fmt.Println()
			}
		} else {
			handleCmd(text)
		}
		pP()
	}
	fmt.Println()
}
