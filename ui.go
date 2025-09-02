package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (g *Game) draw() {
	s := g.s
	w, h := s.Size()
	s.Clear()

	switch g.phase {
	case PhaseSetupNumTeams, PhaseSetupTeamNames:
		drawCenteredText(s, 0, 0, w, h/2, tcell.StyleDefault.Bold(true), g.prompt+g.inputBuf)
		if g.msg != "" {
			drawCenteredText(s, 0, h/2, w, h/2, tcell.StyleDefault, g.msg)
		}
	case PhaseBoard:
		g.drawBoard()
		g.drawTeams()
		g.drawStatus()
	case PhaseQuestion:
		g.drawQuestion()
		g.drawStatus()
	}

	s.Show()

	// hacky solution, render images after tcell has rendered the screen
	if g.phase == PhaseQuestion && g.curQ != nil && g.curQ.ImagePath != "" && g.imageSupported {
		g.renderImageAfterShow()
	}
}

func (g *Game) drawBoard() {
	s := g.s
	w, _ := s.Size()
	cols := len(g.board.Categories)
	colW := max(12, w/cols)
	categoryHeight := 3
	boxHeight := 7 // game board box height

	for c, cat := range g.board.Categories {
		x0 := c * colW
		fillBox(s, x0, 0, colW, categoryHeight, styleHeader())
		drawBox(s, x0, 0, colW, categoryHeight, styleHeader())
		drawCenteredText(s, x0, 1, colW, 1, styleHeader().Bold(true), strings.ToUpper(cat.Name))
		for r := range 5 {
			y := categoryHeight + r*boxHeight
			fillBox(s, x0, y, colW, boxHeight, styleCell())
			drawBox(s, x0, y, colW, boxHeight, styleCell())
			q := cat.Questions[r]
			label := fmt.Sprintf("$%d", q.Value)
			if q.Picked {
				label = "—"
			}
			st := styleCell().Bold(true)
			if c == g.cursorCol && r == g.cursorRow && g.phase == PhaseBoard {
				st = st.Reverse(true)
			}
			drawCenteredText(s, x0, y+3, colW, 1, st, label)
		}
	}
}

func (g *Game) drawTeams() {
	s := g.s
	_, h := s.Size()
	baseline := 3 + 5*7 + 1 // for box height, adjust when adjusting box height
	if baseline >= h {
		baseline = h - 3
	}
	for x := range 1_000_000 {
		setCell(s, x, baseline, '─', styleDim())
	}
	y := baseline + 1
	for i, t := range g.teams {
		line := fmt.Sprintf("%d) %s — %d", i+1, t.Name, t.Score)
		drawText(s, 1, y+i, styleTeam(), line)
	}
}

func (g *Game) drawStatus() {
	s := g.s
	w, h := s.Size()
	status := g.msg
	if g.phase == PhaseBoard && g.inputBuf != "" {
		status = fmt.Sprintf("score command: %s  (press enter to apply)", g.inputBuf)
	}
	pad := strings.Repeat(" ", max(0, w-len(status)-1))
	drawText(s, 0, h-1, styleStatus(), status+pad)
}

func (g *Game) drawQuestion() {
	s := g.s
	w, h := s.Size()
	if g.curQ == nil {
		return
	}
	fillBox(s, 0, 0, w, h-1, styleQuestion())
	drawBox(s, 0, 0, w, h-1, styleQuestion())

	title := fmt.Sprintf("%s — $%d", g.curQ.Category, g.curQ.Value)
	drawCenteredText(s, 0, 3, w, 1, styleQuestion().Bold(true), title)

	separatorY := 5
	for x := w / 4; x < 3*w/4; x++ {
		setCell(s, x, separatorY, '─', styleQuestion())
	}

	questionAreaY := QuestionAreaY
	questionAreaH := h - questionAreaY - 3

	textToShow, textStyle := g.curQ.Q, styleQuestion().Bold(true)
	if g.showAnswer {
		textToShow, textStyle = g.curQ.A, styleQuestion().Bold(true).Foreground(tcell.ColorLightGreen)
	}

	if g.curQ.ImagePath != "" && g.imageSupported && g.imageRenderer != nil {
		// split screen: image left, text right
		imageWidth := w / 2
		drawCenteredWrappedText(s, imageWidth+1, questionAreaY, w-imageWidth-1, questionAreaH, textStyle, textToShow)
	} else {
		// full width text
		drawCenteredWrappedText(s, 0, questionAreaY, w, questionAreaH, textStyle, textToShow)
	}
}
