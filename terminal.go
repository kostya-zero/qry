package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func PrintError(msg string) {
	fmt.Printf("%s %s\n", lipgloss.NewStyle().Foreground(ColorRed).Render("error:"), msg)
}
