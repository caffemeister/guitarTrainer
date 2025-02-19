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
// 1. Add some sorta style ✅
// 2. Add a "get 4 random exercises to work on today" button ✅
// 3. Hotkey to launch metronome in Google ✅
// 4. Integrate the "find notes" trainer thing ✅
// 5. Add tracker to track diff of scores per month

type model struct {
	cursor            int                       // Cursor for navigating lists
	currentLevel      string                    // Current level: "main" or "submenu"
	selectedTech      string                    // Currently selected technique in "main" menu
	techniques        map[string]map[string]int // Map of techniques to exercises and their BPMs
	trackerTechniques map[string]map[string]int
	keys              []string // Ordered keys for the current menu
	input             string   // Input buffer for editing
	showPopup         bool
	spinner           spinner.Model
	showSuccess       bool
	successTime       time.Time
	fourExercises     map[string]map[string]int
	exerciseKeys      []string
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
	oldValues, err := loadTechniques("tracker.json")
	if err != nil {
		log.Fatalln(err)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		techniques:        tech,
		trackerTechniques: oldValues,
		currentLevel:      "main",
		cursor:            0,
		keys:              getKeys(tech),
		spinner:           s,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		if m.currentLevel == "fourExercises" {
			switch key {
			case "q":
				return m, tea.Quit
			case "e":
				m.selectedTech = m.exerciseKeys[m.cursor]
				m.showPopup = true
				m.input = ""
			case "esc":
				if m.showPopup {
					m.showPopup = false
				} else {
					m.showPopup = false
					m.currentLevel = "main"
					m.cursor = 0
				}
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(m.exerciseKeys)-1 {
					m.cursor++
				}
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "enter":
				bpm, err := strconv.Atoi(strings.TrimSpace(m.input))
				if err == nil {
					exercise := m.fourExercises[m.selectedTech]
					for ex := range exercise {
						m.fourExercises[m.selectedTech][ex] = bpm
						m.techniques[m.selectedTech][ex] = bpm
					}

					err = saveTechniques("techniques.json", m.techniques)
					if err != nil {
						log.Println("Error saving techniques:", err)
					}
					m.showSuccess = true
					m.successTime = time.Now().Add(3 * time.Second)
				}

				m.showPopup = false
				m.input = ""
			default:
				if key >= "0" && key <= "9" {
					m.input += key
				}
			}
		}

		if m.currentLevel == "noteLocation" {
			if m.showPopup {
				switch key {
				case "esc":
					m.showPopup = false
					m.currentLevel = "main"
					m.cursor = 0
				case "q":
					return m, tea.Quit
				case "m":
					launchMetronome()
				case "enter":
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

		if m.currentLevel == "submenu" {
			if m.showPopup {
				switch key {
				case "enter":
					bpm, err := strconv.Atoi(strings.TrimSpace(m.input))
					if err == nil {
						exercise := m.keys[m.cursor]
						m.techniques[m.selectedTech][exercise] = bpm

						err = saveTechniques("techniques.json", m.techniques)
						if err != nil {
							log.Println("Error saving techniques:", err)
						}
						m.showSuccess = true
						m.successTime = time.Now().Add(3 * time.Second)
					}

					m.showPopup = false
					m.input = ""
				case "esc":
					m.showPopup = false
					m.input = ""
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
			} else {
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
					m.showPopup = true
					m.input = ""
				case ",":
					getFourExercises()
				case "m":
					launchMetronome()
				case "esc":
					m.currentLevel = "main"
					m.cursor = 0
					m.keys = getKeys(m.techniques) // Reset keys to main menu techniques
				}
			}
		}

		if m.currentLevel == "main" {
			switch key {
			case "q":
				return m, tea.Quit
			case ",":
				m.currentLevel = "fourExercises"
				m.cursor = 0
				exerc, err := getFourExercises()
				if err != nil {
					log.Fatal(err)
				}
				m.fourExercises = exerc
				m.exerciseKeys = getKeys(exerc)
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
					m.currentLevel = "noteLocation"
					m.showPopup = true
					m.cursor = 0
				} else {
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

	if m.currentLevel == "fourExercises" {
		if m.showSuccess {
			b.WriteString(successStyle.Render("BPM updated!"))
			b.WriteString("\n")
		}

		if !m.showPopup {
			b.WriteString(nameStyle.Render(fmt.Sprintf("The council has decided your fate... %s", m.spinner.View())))
			b.WriteString("\n\n")

			for i, key := range m.exerciseKeys {
				b.WriteString(fmt.Sprintf("%s:\n", key))

				exerc := m.fourExercises[key]
				exerciseCursor := " "
				if m.cursor == i {
					exerciseCursor = "->"
				}
				for ex, bpm := range exerc {
					b.WriteString(fmt.Sprintf("%s %s -- %d BPM\n", exerciseCursor, ex, bpm))
				}
			}
		}

		b.WriteString("\n")
		b.WriteString(navGuideStyle.Render("[up/down] Navigate • [esc] Back • [e] Edit BPM • [q] Quit\n"))
	}

	// Rendering for noteLocation
	if m.currentLevel == "noteLocation" {
		b.WriteString(nameStyle.Render(fmt.Sprintf("Let's practice locating notes! %s", m.spinner.View())))
		b.WriteString("\n")
		s := "Select number of notes:\n\n"

		if m.showPopup && !notesGotten {
			for i := 1; i <= 9; i++ {
				cursor := " "
				if m.cursor == i-1 {
					cursor = "->"
				}
				s += fmt.Sprintf("%s %d ", cursor, i)
			}
			b.WriteString("\n")
			b.WriteString(popupStyle.Render(s))
		}

		if m.showPopup && notesGotten {
			s := "Select number of notes:\n\n"
			for i := 1; i <= 9; i++ {
				cursor := " "
				if m.cursor == i-1 {
					cursor = "->"
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
		b.WriteString(navGuideStyle.Render("•"))
		b.WriteString(hotkeyStyle.Render(" [,] 4 Random Exercises "))
		b.WriteString(navGuideStyle.Render("• [q] Quit\n"))
	} else if m.currentLevel == "submenu" {
		// Submenu rendering
		if m.showSuccess {
			b.WriteString(successStyle.Render("BPM updated!"))
			b.WriteString("\n")
		}

		b.WriteString(nameStyle.Render(fmt.Sprintf("Exercises for %s:", m.selectedTech)))
		b.WriteString("\n")
		b.WriteString("\t\t\t\t\t\tLast month\n")
		exercises := m.techniques[m.selectedTech]
		trackerExercises := m.trackerTechniques[m.selectedTech]
		for i, exercise := range m.keys {
			cursor := " "
			if i == m.cursor {
				cursor = "->"
			}
			output := fmt.Sprintf("%s %s: %d BPM", cursor, exercise, exercises[exercise])
			lastMonthPosition := 51
			outputLength := len(output)

			paddingLength := lastMonthPosition - outputLength
			if paddingLength < 0 {
				paddingLength = 0
			}

			padding := strings.Repeat(" ", paddingLength)

			b.WriteString(output + padding + strconv.Itoa(trackerExercises[exercise]) + "\n")
		}
		b.WriteString("\n")
		b.WriteString(navGuideStyle.Render("[up/down] Navigate • [enter] Select • [e] Edit BPM •"))
		b.WriteString(hotkeyStyle.Render(" [m] Metronome "))
		b.WriteString(navGuideStyle.Render("• [q] Quit\n"))
	}

	// Render the popup for editing BPM (only if in submenu or noteLocation)
	if (m.currentLevel == "submenu" || m.currentLevel == "fourExercises") && m.showPopup {
		if m.currentLevel == "submenu" {
			popupContent := fmt.Sprintf(
				"Editing [%s] BPM\n\n%s\n\n[enter] Save • [esc] Cancel",
				m.keys[m.cursor],
				m.input,
			)
			b.WriteString("\n" + popupStyle.Render(popupContent))
		} else if m.currentLevel == "fourExercises" {
			popupContent := fmt.Sprintf(
				"Editing [%s] BPM\n\n%s\n\n[enter] Save • [esc] Cancel",
				m.exerciseKeys[m.cursor],
				m.input,
			)
			b.WriteString("\n" + popupStyle.Render(popupContent))
		}
	}

	return b.String()
}

func main() {
	// check if tracker exists, if not, create one
	_, err := os.Stat("tracker.json")
	if os.IsNotExist(err) {
		err := spawnTracker()
		if err != nil {
			log.Fatal(err)
		}
	}

	// if enough time has passed, update tracker.json
	currentTime := time.Now()
	lastUpdateTime, err := getLastUpdateTime()
	if err != nil {
		log.Fatal(err)
	}
	if currentTime.After(lastUpdateTime.AddDate(0, 1, 0)) {
		updateTrackerJSON()
	}

	// run program
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting app: %v", err)
		os.Exit(1)
	}
}
