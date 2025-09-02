package main

import "github.com/gdamore/tcell/v2"

func styleHeader() tcell.Style {
	return tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
}
func styleCell() tcell.Style {
	return tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorYellow)
}
func styleDim() tcell.Style  { return tcell.StyleDefault.Foreground(tcell.ColorSilver) }
func styleTeam() tcell.Style { return tcell.StyleDefault.Foreground(tcell.ColorWhite) }
func styleStatus() tcell.Style {
	return tcell.StyleDefault.Background(tcell.ColorSilver).Foreground(tcell.ColorBlack)
}
func styleQuestion() tcell.Style {
	return tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
}
