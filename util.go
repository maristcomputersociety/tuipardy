package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

func drawText(s tcell.Screen, x, y int, st tcell.Style, text string) {
	for i, ch := range text {
		s.SetContent(x+i, y, ch, nil, st)
	}
}

func drawCenteredText(s tcell.Screen, x, y, w, h int, st tcell.Style, text string) {
	lines := strings.Split(text, "\n")
	cy := y + h/2 - len(lines)/2
	for i, line := range lines {
		cx := x + w/2 - len(line)/2
		if cx < x {
			cx = x
		}
		drawText(s, cx, cy+i, st, line)
	}
}

func setCell(s tcell.Screen, x, y int, ch rune, st tcell.Style) {
	s.SetContent(x, y, ch, nil, st)
}

func drawBox(s tcell.Screen, x, y, w, h int, st tcell.Style) {
	if w <= 0 || h <= 0 {
		return
	}

	// top/bottom
	for i := range w {
		s.SetContent(x+i, y, '─', nil, st)
		s.SetContent(x+i, y+h-1, '─', nil, st)
	}

	// sides
	for j := range h {
		s.SetContent(x, y+j, '│', nil, st)
		s.SetContent(x+w-1, y+j, '│', nil, st)
	}

	// corners
	s.SetContent(x, y, '┌', nil, st)
	s.SetContent(x+w-1, y, '┐', nil, st)
	s.SetContent(x, y+h-1, '└', nil, st)
	s.SetContent(x+w-1, y+h-1, '┘', nil, st)
}

func fillBox(s tcell.Screen, x, y, w, h int, st tcell.Style) {
	if w <= 0 || h <= 0 {
		return
	}

	for j := range h {
		for i := range w {
			s.SetContent(x+i, y+j, ' ', nil, st)
		}
	}
}

func trimSpaces(s string) string { return strings.TrimSpace(strings.Join(strings.Fields(s), " ")) }

func drawCenteredWrappedText(s tcell.Screen, x, y, w, h int, st tcell.Style, text string) {
	words := strings.Fields(text)
	lines := []string{}
	line := ""

	for _, word := range words {
		if len(line)+len(word)+1 > w-4 {
			if line != "" {
				lines = append(lines, line)
				line = word
			} else {
				lines = append(lines, word)
				line = ""
			}
		} else {
			if line == "" {
				line = word
			} else {
				line += " " + word
			}
		}
	}
	if line != "" {
		lines = append(lines, line)
	}

	startY := y + (h-len(lines)*2)/2
	if startY < y {
		startY = y
	}

	for i, textLine := range lines {
		lineY := startY + i*2
		if lineY >= y+h {
			break
		}
		drawCenteredText(s, x, lineY, w, 1, st, textLine)
	}
}
