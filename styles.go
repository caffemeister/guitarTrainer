package main

import "github.com/charmbracelet/lipgloss"

var navGuideStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FAFAFA")).
	Faint(true)

var nameStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	Bold(true)

var choicesStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("201"))

var successStyle = lipgloss.NewStyle().
	Align(lipgloss.Center).
	Background(lipgloss.Color("#04B575")).
	Underline(true).
	Width(30)
