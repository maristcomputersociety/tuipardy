package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

func (g *Game) handleMouse(ev *tcell.EventMouse) {
	if g.phase != PhaseBoard {
		return
	}
	x, y := ev.Position()
	btn := ev.Buttons()
	if btn&tcell.Button1 == 0 {
		return
	}
	col, row, ok := g.hitTestCell(x, y)
	if !ok {
		return
	}
	now := time.Now()
	if g.lastCell == [2]int{col, row} && now.Sub(g.lastClick) < 350*time.Millisecond {
		g.cursorCol, g.cursorRow = col, row
		g.openSelected()
		g.lastClick = time.Time{}
		return
	}
	g.cursorCol, g.cursorRow = col, row
	g.lastCell = [2]int{col, row}
	g.lastClick = now
}

func (g *Game) hitTestCell(x, y int) (c, r int, ok bool) {
	w, _ := g.s.Size()
	cols := len(g.board.Categories)
	colW := max(MinColumnWidth, w/cols)
	if y < CategoryHeight || y >= CategoryHeight+QuestionsPerCategory*CellHeight {
		return 0, 0, false
	}
	r = (y - CategoryHeight) / CellHeight
	c = x / colW
	if c < 0 || c >= cols || r < 0 || r >= QuestionsPerCategory {
		return 0, 0, false
	}
	return c, r, true
}
