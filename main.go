package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO:
// 1. Add style
// 2. Add a "get 4 random exercises to work on today" button like JP said
// 3. Hotkey to launch metronome in Google

type model struct {
	cursor       int                       // Cursor for navigating lists
	currentLevel string                    // Current level: "main" or "submenu"
	selectedTech string                    // Currently selected technique in "main" menu
	techniques   map[string]map[string]int // Map of techniques to exercises and their BPMs
	keys         []string                  // Ordered keys for the current menu
	input        string                    // Input buffer for editing
	mode         string                    // Current mode: "view" or "edit"
	showPopup    bool
}

var popupStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(1, 2).
	Align(lipgloss.Center).
	Width(40)

func (m model) Init() tea.Cmd {
	return nil
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

func initialModel() model {
	tech, err := loadTechniques("techniques.json")
	if err != nil {
		log.Fatalln(err)
	}
	return model{
		techniques:   tech,
		currentLevel: "main",
		cursor:       0,
		keys:         getKeys(tech),
		mode:         "view",
	}
}

func launchMetronome() error {
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://www.google.com/search?q=metronome").Start()
	if err != nil {
		log.Fatalln(err)
	}
	return err
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Handle the popup mode
		if m.showPopup {
			switch key {
			case "enter":
				// Save the new BPM
				bpm, err := strconv.Atoi(strings.TrimSpace(m.input))
				if err == nil {
					exercise := m.keys[m.cursor]
					m.techniques[m.selectedTech][exercise] = bpm

					// Save the changes to the JSON file
					err = saveTechniques("techniques.json", m.techniques)
					if err != nil {
						log.Println("Error saving techniques:", err)
					}
				}
				// Close the popup
				m.showPopup = false
				m.input = ""
			case "esc":
				// Close the popup without saving
				m.showPopup = false
				m.input = ""
			case "backspace":
				// Remove the last character from the input
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			default:
				// Allow only numerical input (0-9)
				if key >= "0" && key <= "9" {
					m.input += key
				}
			}
			return m, nil
		}

		// Main menu navigation
		if m.currentLevel == "main" {
			switch key {
			case "q":
				return m, tea.Quit
			case "m":
				launchMetronome()
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(m.keys)-1 {
					m.cursor++
				}
			case "enter":
				// Enter submenu for the selected technique
				m.selectedTech = m.keys[m.cursor]
				m.currentLevel = "submenu"
				m.cursor = 0
				m.keys = getKeys(m.techniques[m.selectedTech]) // Update keys for the selected technique's exercises
			}
		} else if m.currentLevel == "submenu" {
			// Submenu navigation
			switch key {
			case "q":
				return m, tea.Quit
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(m.keys)-1 {
					m.cursor++
				}
			case "e":
				// Open the popup for editing BPM
				m.showPopup = true
				m.input = ""
			case "esc":
				// Return to main menu
				m.currentLevel = "main"
				m.cursor = 0
				m.keys = getKeys(m.techniques) // Reset keys to main menu techniques
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	if m.currentLevel == "main" {
		b.WriteString("What are we working on?:\n\n")
		for i, technique := range m.keys {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, technique))
		}
		b.WriteString("\n[up/down] Navigate • [enter] Select • [m] Metronome • [q] Quit\n")
	} else if m.currentLevel == "submenu" {
		b.WriteString(fmt.Sprintf("Exercises for %s:\n\n", m.selectedTech))
		exercises := m.techniques[m.selectedTech]
		for i, exercise := range m.keys {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("%s %s: %d BPM\n", cursor, exercise, exercises[exercise]))
		}
		b.WriteString("\n[up/down] Navigate • [e] Edit BPM • [m] Metronome • [esc] Back • [q] Quit\n")
	}

	// Render the popup if it's visible
	if m.showPopup {
		popupContent := fmt.Sprintf(
			"Editing [%s] BPM\n\n%s\n\n[enter] Save • [esc] Cancel",
			m.keys[m.cursor],
			m.input,
		)
		b.WriteString("\n" + popupStyle.Render(popupContent))
	}

	return b.String()
}

func main() {

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting app: %v", err)
		os.Exit(1)
	}
}
