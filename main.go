package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO:
// 1. Add style ✅
// 2. Add a "get 4 random exercises to work on today" button like JP said
// 3. Hotkey to launch metronome in Google ✅
// 4. Integrate the "find notes" trainer thing ✅

type model struct {
	cursor       int                       // Cursor for navigating lists
	currentLevel string                    // Current level: "main" or "submenu"
	selectedTech string                    // Currently selected technique in "main" menu
	techniques   map[string]map[string]int // Map of techniques to exercises and their BPMs
	keys         []string                  // Ordered keys for the current menu
	input        string                    // Input buffer for editing
	mode         string                    // Current mode: "view" or "edit"
	showPopup    bool
	spinner      spinner.Model
	showSuccess  bool
	successTime  time.Time
}

// Add the names of all techniques that have hotkeys assigned to them here
var HOTKEYS = []string{"Gallop picking rhythms"}
var NOTES = []string{"A", "B", "C", "D", "E", "F", "G", "A#", "C#", "D#", "G#"}
var notesGotten bool
var notes []string

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func initialModel() model {
	tech, err := loadTechniques("techniques.json")
	if err != nil {
		log.Fatalln(err)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		techniques:   tech,
		currentLevel: "main",
		cursor:       0,
		keys:         getKeys(tech),
		mode:         "view",
		spinner:      s,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Handle interaction in the "noteLocation" level
		if m.currentLevel == "noteLocation" {
			if m.showPopup {
				// Handle the input inside the noteLocation popup
				switch key {
				case "esc":
					// Cancel the mini-game and go back to the main menu
					m.showPopup = false
					m.currentLevel = "main"
					m.cursor = 0
				case "q":
					return m, tea.Quit
				case "m":
					launchMetronome()
				case "enter":
					// Save the selected number
					selectedNumber := m.cursor + 1 // Convert index to number (1-9)
					notes = getRandNotes(selectedNumber)
					if len(notes) > 0 {
						notesGotten = true
					}
				case "left":
					if m.cursor > 0 {
						m.cursor--
					}
				case "right":
					if m.cursor < 8 {
						m.cursor++
					}
				}
				return m, nil
			}
			return m, nil
		}

		// Handle popup mode (submenu and noteLocation)
		if m.currentLevel == "submenu" {
			if m.showPopup {
				switch key {
				case "enter":
					// Save BPM if in edit mode
					bpm, err := strconv.Atoi(strings.TrimSpace(m.input))
					if err == nil {
						exercise := m.keys[m.cursor]
						m.techniques[m.selectedTech][exercise] = bpm

						// Save to JSON file
						err = saveTechniques("techniques.json", m.techniques)
						if err != nil {
							log.Println("Error saving techniques:", err)
						}
						m.showSuccess = true
						m.successTime = time.Now().Add(3 * time.Second)
					}

					// Close the popup
					m.showPopup = false
					m.input = ""
				case "esc":
					// Close the popup without saving
					m.showPopup = false
					m.input = ""
				case ",":
					getFourExercises()
				case "m":
					launchMetronome()
				case "backspace":
					// Remove last character from the input
					if len(m.input) > 0 {
						m.input = m.input[:len(m.input)-1]
					}
				default:
					// Allow only numerical input
					if key >= "0" && key <= "9" {
						m.input += key
					}
				}
				return m, nil
			}
		}

		// Main menu navigation
		if m.currentLevel == "main" {
			switch key {
			case "q":
				return m, tea.Quit
			case ",":
				getFourExercises()
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
				if m.keys[m.cursor] == "Gallop picking rhythms" {
					launchGallopPicking()
				} else if m.keys[m.cursor] == "Note Location" {
					// Enter the Note Location mini-game
					m.currentLevel = "noteLocation"
					m.showPopup = true
					m.cursor = 0
				} else {
					// Enter submenu for the selected technique
					m.selectedTech = m.keys[m.cursor]
					m.currentLevel = "submenu"
					m.cursor = 0
					m.keys = getKeys(m.techniques[m.selectedTech]) // Update keys for the submenu
				}
			default:
				var cmd tea.Cmd
				m.spinner, cmd = m.spinner.Update(msg)
				return m, cmd
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
			case ",":
				getFourExercises()
			case "m":
				launchMetronome()
			case "esc":
				// Return to main menu
				m.currentLevel = "main"
				m.cursor = 0
				m.keys = getKeys(m.techniques) // Reset keys to main menu techniques
			}
		}
		if m.showSuccess && time.Now().After(m.successTime) {
			m.showSuccess = false
		}
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)

	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	// Rendering for noteLocation mini-game
	if m.currentLevel == "noteLocation" {
		b.WriteString(nameStyle.Render(fmt.Sprintf("Let's practice locating notes! %s", m.spinner.View())))
		b.WriteString("\n")
		s := "Select number of notes:\n\n"

		if m.showPopup && !notesGotten {
			for i := 1; i <= 9; i++ {
				cursor := " "
				if m.cursor == i-1 {
					cursor = "->" // Highlight the selected number
				}
				s += fmt.Sprintf("%s %d ", cursor, i)
			}
			// Render the popup style around the content
			b.WriteString("\n")
			b.WriteString(popupStyle.Render(s)) // Apply the popup style to the content
		}

		if m.showPopup && notesGotten {
			s := "Select number of notes:\n\n"
			for i := 1; i <= 9; i++ {
				cursor := " "
				if m.cursor == i-1 {
					cursor = "->" // Highlight the selected number
				}
				s += fmt.Sprintf("%s %d ", cursor, i)
			}
			s += "\n--------------------------------\n\n"
			for _, note := range notes {
				s += note + "  "
			}
			b.WriteString(popupStyle.Render(s))
		}

		b.WriteString("\n")
		b.WriteString(navGuideStyle.Render("[left/right] Navigate • [enter] Select •"))
		b.WriteString(hotkeyStyle.Render(" [m] Metronome "))
		b.WriteString(navGuideStyle.Render("• [q] Quit\n"))
	}

	// Main menu rendering
	if m.currentLevel == "main" {
		if m.showSuccess {
			b.WriteString(successStyle.Render("BPM updated!"))
			b.WriteString("\n")
		}

		b.WriteString(nameStyle.Render(fmt.Sprintf("What are we working on? %s", m.spinner.View())))
		b.WriteString("\n\n")
		for i, technique := range m.keys {
			cursor := " "
			if i == m.cursor {
				cursor = "->"
			}
			techniqueIsHotkey := false
			for _, h := range HOTKEYS {
				if h == technique {
					techniqueIsHotkey = true
					break
				}
			}
			if techniqueIsHotkey {
				b.WriteString(hotkeyStyle.Render(fmt.Sprintf("%s %s", cursor, technique)))
				b.WriteString("\n")
			} else if technique == "Note Location" {
				b.WriteString(noteLocationStyle.Render(fmt.Sprintf("%s %s", cursor, technique)))
				b.WriteString("\n")
			} else {
				b.WriteString(fmt.Sprintf("%s %s\n", cursor, technique))
			}
		}
		b.WriteString("\n")
		b.WriteString(navGuideStyle.Render("[up/down] Navigate • [enter] Select •"))
		b.WriteString(hotkeyStyle.Render(" [m] Metronome "))
		b.WriteString(navGuideStyle.Render("• [q] Quit\n"))
	} else if m.currentLevel == "submenu" {
		// Submenu rendering
		if m.showSuccess {
			b.WriteString(successStyle.Render("BPM updated!"))
			b.WriteString("\n")
		}

		b.WriteString(nameStyle.Render(fmt.Sprintf("Exercises for %s:", m.selectedTech)))
		b.WriteString("\n\n")
		exercises := m.techniques[m.selectedTech]
		for i, exercise := range m.keys {
			cursor := " "
			if i == m.cursor {
				cursor = "->"
			}
			b.WriteString(fmt.Sprintf("%s %s: %d BPM\n", cursor, exercise, exercises[exercise]))
		}
		b.WriteString("\n")
		b.WriteString(navGuideStyle.Render("[up/down] Navigate • [enter] Select • [e] Edit BPM •"))
		b.WriteString(hotkeyStyle.Render(" [m] Metronome "))
		b.WriteString(navGuideStyle.Render("• [q] Quit\n"))
	}

	// Render the popup for editing BPM (only if in submenu or noteLocation)
	if (m.currentLevel == "submenu" || m.currentLevel != "noteLocation") && m.showPopup {
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
