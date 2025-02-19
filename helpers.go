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
	return chosenMap, nil
}

func spawnTracker() error {
	// create a new tracker
	newTracker, err := os.Create("tracker.json")
	if err != nil {
		return err
	}
	defer newTracker.Close()

	updateTrackerJSON()

	// create a trackerState file
	newTrackerStateFile, err := os.Create("trackerState.json")
	if err != nil {
		return err
	}
	defer newTrackerStateFile.Close()

	// add a state to trackerState
	state := TrackerState{
		TrackerFile: "tracker.json",
		LastEdited:  time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile("trackerState.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getLastUpdateTime() (time.Time, error) {
	var state TrackerState

	jsonFile, err := os.Open("trackerState.json")
	if err != nil {
		return time.Time{}, err
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&state); err != nil {
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
	err = saveTechniques("tracker.json", tech)
	if err != nil {
		return err
	}
	return nil
}
