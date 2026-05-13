package main

import (
	"os"

	"charm.land/lipgloss/v2"
)

// Colors
var (
	ColorPrimary    = lipgloss.Color("#5F87AF")
	ColorRed        = lipgloss.Color("#FF5F5F")
	ColorGreen      = lipgloss.Color("#5FAF5F")
	ColorYellow     = lipgloss.Color("#FFAF00")
	ColorBaseGray   = lipgloss.Color("#767676")
	ColorBrightGray = lipgloss.Color("#D0D0D0")
	ColorWhite      = lipgloss.Color("#FFFFFF")
)

// Table
var (
	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Foreground(ColorWhite).Padding(0, 1)
	TableCellStyle   = lipgloss.NewStyle().Padding(0, 1).Foreground(ColorBrightGray)
	TableBorderStyle = lipgloss.NewStyle().Foreground(ColorBaseGray)
)

var (
	PromptStyle      = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	WelcomeStyle     = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	SubtextStyle     = lipgloss.NewStyle().Foreground(ColorBaseGray)
	InternalCmdStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	DescStyle        = lipgloss.NewStyle().Foreground(ColorBrightGray)
	SuccessStyle     = lipgloss.NewStyle().Foreground(ColorGreen)
)

func setupColors() {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		ColorPrimary = lipgloss.NoColor{}
		ColorRed = lipgloss.NoColor{}
		ColorGreen = lipgloss.NoColor{}
		ColorYellow = lipgloss.NoColor{}
		ColorBaseGray = lipgloss.NoColor{}
		ColorBrightGray = lipgloss.NoColor{}
		ColorWhite = lipgloss.NoColor{}

		TableHeaderStyle = lipgloss.NewStyle().Align(lipgloss.Center).Padding(0, 1)
		TableCellStyle = lipgloss.NewStyle().Padding(0, 1)
		TableBorderStyle = lipgloss.NewStyle()

		PromptStyle = lipgloss.NewStyle()
		WelcomeStyle = lipgloss.NewStyle()
		SubtextStyle = lipgloss.NewStyle()
		InternalCmdStyle = lipgloss.NewStyle()
		DescStyle = lipgloss.NewStyle()
		SuccessStyle = lipgloss.NewStyle()
	}
}
