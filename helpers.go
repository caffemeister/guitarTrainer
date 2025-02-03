package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
)

func loadTechniques(filename string) (map[string]map[string]int, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse the JSON data into a map
	var techniques map[string]map[string]int
	err = json.Unmarshal(data, &techniques)
	if err != nil {
		return nil, err
	}

	return techniques, nil
}

func saveTechniques(filename string, techniques map[string]map[string]int) error {
	// Convert techniques map to JSON
	data, err := json.MarshalIndent(techniques, "", "  ")
	if err != nil {
		return err
	}

	// Write the data to the file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func launchMetronome() error {
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://www.google.com/search?q=metronome").Start()
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func launchGallopPicking() error {
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://www.youtube.com/watch?v=S-6Iq2wuf0A").Start()
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func getRandNotes(times int) []string {
	var notes []string
	remainingNotes := append([]string(nil), NOTES...)

	for i := 0; i < times; i++ {
		if len(remainingNotes) == 0 {
			break
		}
		rIndex := rand.Intn(len(remainingNotes))
		notes = append(notes, remainingNotes[rIndex])
		remainingNotes = append(remainingNotes[:rIndex], remainingNotes[rIndex+1:]...)
	}

	return notes
}

func getKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	// Sort the keys to ensure consistent ordering
	sort.Strings(keys)
	return keys
}

func getFourExercises() (map[string]map[string]int, error) {
	var chosenMap = make(map[string]map[string]int)

	tech, err := loadTechniques("techniques.json")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// categories from which to pick exercises from
	categories := []string{"Alternate Picking", "Economy Picking", "Legato", "Sweep Picking"}

	for _, category := range categories {
		exercises, exists := tech[category]
		if !exists || len(exercises) == 0 {
			continue
		}

		exerciseKeys := getKeys(exercises)
		r := rand.Intn(len(exerciseKeys))
		randExerc := exerciseKeys[r]

		if chosenMap[category] == nil {
			chosenMap[category] = make(map[string]int)
		}

		bpm := tech[category][randExerc]
		chosenMap[category][randExerc] = bpm
	}

	fmt.Println(chosenMap)
	return chosenMap, nil
}
