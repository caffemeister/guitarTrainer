package main

import "github.com/charmbracelet/lipgloss"

var navGuideStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FAFAFA")).
	Faint(true)

var nameStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	Bold(true)

var successStyle = lipgloss.NewStyle().
	Align(lipgloss.Center).
	Background(lipgloss.Color("#04B575")).
	Underline(true).
	Width(30)

var hotkeyStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("6"))

var noteLocationStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3"))

var popupStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(1, 2).
	Align(lipgloss.Center).
	Width(40)
