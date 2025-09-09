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

	// hacky solution, put text to stdout for kitty text sizing
	if g.phase == PhaseQuestion && g.curQ != nil && g.textToRender != "" {
		g.renderTextAfterShow()
	}
}

func (g *Game) drawBoard() {
	s := g.s
	w, _ := s.Size()
	cols := len(g.board.Categories)
	colW := max(MinColumnWidth, w/cols)
	categoryHeight := CategoryHeight
	boxHeight := CellHeight

	for c, cat := range g.board.Categories {
		g.drawCategoryHeader(s, c, colW, categoryHeight, cat)
		g.drawCategoryCells(s, c, colW, categoryHeight, boxHeight, cat)
	}
}

// drawCategoryHeader renders a single category header
func (g *Game) drawCategoryHeader(s tcell.Screen, col, colW, categoryHeight int, cat *Category) {
	x0 := col * colW
	fillBox(s, x0, 0, colW, categoryHeight, styleHeader())
	drawBox(s, x0, 0, colW, categoryHeight, styleHeader())
	drawCenteredText(s, x0, 1, colW, 1, styleHeader().Bold(true), strings.ToUpper(cat.Name))
}

// drawCategoryCells renders all question cells for a category
func (g *Game) drawCategoryCells(s tcell.Screen, col, colW, categoryHeight, boxHeight int, cat *Category) {
	for r := range QuestionsPerCategory {
		g.drawQuestionCell(s, col, r, colW, categoryHeight, boxHeight, cat.Questions[r])
	}
}

// drawQuestionCell renders a single question cell
func (g *Game) drawQuestionCell(s tcell.Screen, col, row, colW, categoryHeight, boxHeight int, q *Question) {
	x0 := col * colW
	y := categoryHeight + row*boxHeight

	fillBox(s, x0, y, colW, boxHeight, styleCell())
	drawBox(s, x0, y, colW, boxHeight, styleCell())

	label := fmt.Sprintf("$%d", q.Value)
	if q.Picked {
		label = "—"
	}

	st := styleCell().Bold(true)
	if col == g.cursorCol && row == g.cursorRow && g.phase == PhaseBoard {
		st = st.Reverse(true)
	}

	drawCenteredText(s, x0, y+3, colW, 1, st, label)
}

func (g *Game) drawTeams() {
	s := g.s
	w, h := s.Size()
	baseline := CategoryHeight + QuestionsPerCategory*CellHeight + TeamBaselineOffset
	if baseline >= h {
		baseline = h - 3
	}
	for x := 0; x < w; x++ {
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
	drawText(s, 0, h-StatusBarHeight, styleStatus(), status+pad)
}

// drawQuestion renders the question/answer screen
func (g *Game) drawQuestion() {
	s := g.s
	w, h := s.Size()
	if g.curQ == nil {
		return
	}

	g.drawQuestionBackground(s, w, h)
	g.drawQuestionTitle(s, w)
	g.drawQuestionSeparator(s, w)
	g.drawQuestionContent(s, w, h)
}

// drawQuestionBackground fills and draws the background box for the question screen
func (g *Game) drawQuestionBackground(s tcell.Screen, w, h int) {
	fillBox(s, 0, 0, w, h-StatusBarHeight, styleQuestion())
	drawBox(s, 0, 0, w, h-StatusBarHeight, styleQuestion())
}

// drawQuestionTitle renders the category and value at the top of the question screen
func (g *Game) drawQuestionTitle(s tcell.Screen, w int) {
	title := fmt.Sprintf("%s — $%d", g.curQ.Category, g.curQ.Value)
	drawCenteredText(s, 0, 3, w, 1, styleQuestion().Bold(true), title)
}

// drawQuestionSeparator draws a horizontal separator line below the title
func (g *Game) drawQuestionSeparator(s tcell.Screen, w int) {
	separatorY := 5
	for x := w / 4; x < 3*w/4; x++ {
		setCell(s, x, separatorY, '─', styleQuestion())
	}
}

// drawQuestionContent renders the main question/answer text with optional image
func (g *Game) drawQuestionContent(s tcell.Screen, w, h int) {
	questionAreaY := QuestionAreaY
	questionAreaH := h - questionAreaY - 3

	textToShow, textStyle := g.curQ.Q, styleQuestion().Bold(true)
	if g.showAnswer {
		textToShow, textStyle = g.curQ.A, styleQuestion().Bold(true).Foreground(tcell.ColorLightGreen)
	}

	if g.curQ.ImagePath != "" && g.imageSupported && g.imageRenderer != nil {
		// Use horizontal split: top 65% for image, bottom 35% for text
		g.drawQuestionWithImage(s, w, questionAreaY, questionAreaH, textToShow, textStyle)
	} else {
		// Full screen for text when no image
		g.drawQuestionFullWidth(s, w, questionAreaY, questionAreaH, textToShow, textStyle)
	}
}

// drawQuestionWithImage renders question text below an image in a horizontal split
func (g *Game) drawQuestionWithImage(s tcell.Screen, w, questionAreaY, questionAreaH int, textToShow string, textStyle tcell.Style) {
	// split horizontally: top 65% for image, bottom 35% for text
	imageHeight := questionAreaH * ImageHeightRatio / 100
	textHeight := questionAreaH - imageHeight
	textY := questionAreaY + imageHeight

	textPadding := ImageTextPadding
	adjustedTextY := textY + textPadding
	adjustedTextHeight := textHeight - textPadding

	if adjustedTextHeight < 1 {
		adjustedTextHeight = 1
		adjustedTextY = textY
	}

	clearTextArea(s, 0, adjustedTextY, w, adjustedTextHeight, textStyle.Background(tcell.ColorBlack))

	g.textToRender = textToShow
	g.textStyle = textStyle
	g.textAreaX = 0
	g.textAreaY = adjustedTextY
	g.textAreaW = w
	g.textAreaH = adjustedTextHeight
}

// drawQuestionFullWidth renders question text across the full width of the screen
func (g *Game) drawQuestionFullWidth(s tcell.Screen, w, questionAreaY, questionAreaH int, textToShow string, textStyle tcell.Style) {
	clearTextArea(s, 0, questionAreaY, w, questionAreaH, textStyle.Background(tcell.ColorBlack))

	g.textToRender = textToShow
	g.textStyle = textStyle
	g.textAreaX = 0
	g.textAreaY = questionAreaY
	g.textAreaW = w
	g.textAreaH = questionAreaH
}
