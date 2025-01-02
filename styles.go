package main

import "github.com/charmbracelet/lipgloss"

var navGuideStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	Width(50)

var nameStyle = lipgloss.NewStyle().
	Bold(true).
	Width(20).
	Height(1).
	Align(lipgloss.Center)

var choicesStyle = lipgloss.NewStyle().
	Width(20).
	Align(lipgloss.Left).
	Foreground(lipgloss.Color("#FAFAFA"))
