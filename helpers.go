package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"time"
)

type TrackerState struct {
	TrackerFile string    `json:"tracker_file"`
	LastEdited  time.Time `json:"last_edited"`
}

func loadJSON(filename string, dest interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func loadTechniques(filename string) (map[string]map[string]int, error) {
	var techniques map[string]map[string]int
	err := loadJSON(filename, &techniques)
	if err != nil {
		return nil, err
	}

	return techniques, nil
}

func saveJSON(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func saveTechniques(filename string, techniques map[string]map[string]int) error {
	return saveJSON(filename, techniques)
}

func launchURL(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func launchMetronome() error {
	return launchURL("https://www.google.com/search?q=metronome")
}

func launchGallopPicking() error {
	return launchURL("https://www.youtube.com/watch?v=S-6Iq2wuf0A")
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
	return chosenMap, nil
}

func spawnTracker() error {
	if err := saveJSON("tracker.json", map[string]map[string]int{}); err != nil {
		return err
	}

	if err := updateTrackerJSON(); err != nil {
		return err
	}

	state := TrackerState{
		TrackerFile: "tracker.json",
		LastEdited:  time.Now(),
	}

	return saveJSON("trackerState.json", state)
}

func getLastUpdateTime() (time.Time, error) {
	var state TrackerState

	if err := loadJSON("trackerState.json", &state); err != nil {
		return time.Time{}, err
	}
	return state.LastEdited, nil
}

func updateTrackerJSON() error {
	// copy techniques into tracker
	tech, err := loadTechniques("techniques.json")
	if err != nil {
		return err
	}
	return saveTechniques("tracker.json", tech)
}
