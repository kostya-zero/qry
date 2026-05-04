package main

import "charm.land/lipgloss/v2"

// Colors
var (
	ColorRed        = lipgloss.Color("9")
	СolorBaseGray   = lipgloss.Color("242")
	ColorBrightGray = lipgloss.Color("248")
	ColorWhite      = lipgloss.Color("15")
	ColorYellow     = lipgloss.Color("11")
)

// Table
var (
	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Foreground(ColorWhite)
	TableCellStyle   = lipgloss.NewStyle().Padding(0, 1).Foreground(ColorBrightGray)
	TableBorderStyle = lipgloss.NewStyle().Foreground(СolorBaseGray)
)
