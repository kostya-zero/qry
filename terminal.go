package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func PrintError(msg string) {
	fmt.Printf("%s %s\n", lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render("error:"), msg)
}

func PrintWarn(msg string) {
	fmt.Printf("%s %s\n", lipgloss.NewStyle().Foreground(ColorYellow).Bold(true).Render("warn:"), msg)
}
