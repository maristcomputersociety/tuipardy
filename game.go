package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Game struct {
	s              tcell.Screen
	board          *Board
	phase          int
	teams          []*Team
	minTeams       int
	maxTeams       int
	prompt         string
	inputBuf       string // command buffer
	msg            string // status line
	cursorCol      int
	cursorRow      int
	curQ           *Question
	showAnswer     bool
	lastClick      time.Time
	lastCell       [2]int
	imageRenderer  *ImageRenderer
	imageSupported bool
	textToRender   string // text content to render via stdout
	textStyle      tcell.Style
	textAreaX      int
	textAreaY      int
	textAreaW      int
	textAreaH      int
}

func NewGame(b *Board) *Game {
	imageSupported := IsImageSupported()
	var imageRenderer *ImageRenderer
	if imageSupported {
		imageRenderer = NewImageRenderer()
	}

	return &Game{
		board:          b,
		phase:          PhaseSetupNumTeams,
		minTeams:       MinTeams,
		maxTeams:       MaxTeams,
		imageRenderer:  imageRenderer,
		imageSupported: imageSupported,
	}
}

func (g *Game) Run() error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := s.Init(); err != nil {
		return err
	}
	defer s.Fini()
	g.s = s
	g.prompt = fmt.Sprintf("enter number of teams (%d-%d): ", g.minTeams, g.maxTeams)

	for {
		g.draw()
		if ev := s.PollEvent(); ev != nil {
			switch e := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				if done := g.handleKey(e); done {
					return nil
				}
			case *tcell.EventMouse:
				g.handleMouse(e)
			}
		}
	}
}

func (g *Game) handleKey(e *tcell.EventKey) bool {
	key, r := e.Key(), e.Rune()
	switch g.phase {
	case PhaseSetupNumTeams:
		return g.handleSetupNumTeams(key, r)
	case PhaseSetupTeamNames:
		return g.handleSetupTeamNames(key, r)
	case PhaseBoard:
		return g.handleBoardKey(key, r)
	case PhaseQuestion:
		return g.handleQuestionKey(key, r)
	}
	return false
}

func (g *Game) handleSetupNumTeams(key tcell.Key, r rune) bool {
	switch key {
	case tcell.KeyEsc, tcell.KeyCtrlC:
		return true
	case tcell.KeyEnter:
		if n, err := strconv.Atoi(g.inputBuf); err == nil && n >= g.minTeams && n <= g.maxTeams {
			g.teams = make([]*Team, 0, n)
			g.inputBuf = ""
			g.prompt = fmt.Sprintf("enter name for Team %d: ", len(g.teams)+1)
			g.phase = PhaseSetupTeamNames
		} else {
			g.flashMsg("invalid number; please enter %d-%d", g.minTeams, g.maxTeams)
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(g.inputBuf) > 0 {
			g.inputBuf = g.inputBuf[:len(g.inputBuf)-1]
		}
	default:
		if r >= '0' && r <= '9' {
			g.inputBuf += string(r)
		}
	}
	return false
}

func (g *Game) handleSetupTeamNames(key tcell.Key, r rune) bool {
	switch key {
	case tcell.KeyEsc, tcell.KeyCtrlC:
		return true
	case tcell.KeyEnter:
		name := trimSpaces(g.inputBuf)
		if name == "" {
			name = fmt.Sprintf("Team %d", len(g.teams)+1)
		}
		g.teams = append(g.teams, &Team{Name: name})
		g.inputBuf = ""
		if len(g.teams) == cap(g.teams) {
			g.phase = PhaseBoard
			g.msg = "arrows to move, enter to open, space to reveal, <teamnum><+ | -><score> to modify score"
		} else {
			g.prompt = fmt.Sprintf("enter name for Team %d: ", len(g.teams)+1)
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(g.inputBuf) > 0 {
			g.inputBuf = g.inputBuf[:len(g.inputBuf)-1]
		}
	default:
		if r != 0 {
			g.inputBuf += string(r)
		}
	}
	return false
}

func (g *Game) handleBoardKey(key tcell.Key, r rune) bool {
	switch key {
	case tcell.KeyCtrlC:
		return true
	case tcell.KeyRune:
		if r == 'q' || r == 'Q' {
			return true
		}
		if r == 'h' {
			g.move(-1, 0)
			return false
		}
		if r == 'l' {
			g.move(+1, 0)
			return false
		}
		if r == 'k' {
			g.move(0, -1)
			return false
		}
		if r == 'j' {
			g.move(0, +1)
			return false
		}
		if (r >= '0' && r <= '9') || r == '+' || r == '-' {
			g.inputBuf += string(r)
			return false
		}
	case tcell.KeyEnter:
		if g.inputBuf != "" {
			if applyScoreCommand(g, g.inputBuf) {
				g.inputBuf = ""
			} else {
				g.flashMsg("invalid score command. use format: <team 1-%d><+|-><value>", len(g.teams))
				g.inputBuf = ""
			}
		} else {
			g.openSelected()
		}
		return false
	case tcell.KeyEsc:
		g.inputBuf = ""
		return false
	case tcell.KeyUp:
		g.move(0, -1)
	case tcell.KeyDown:
		g.move(0, +1)
	case tcell.KeyLeft:
		g.move(-1, 0)
	case tcell.KeyRight:
		g.move(+1, 0)
	}
	return false
}

func (g *Game) handleQuestionKey(key tcell.Key, r rune) bool {
	switch key {
	case tcell.KeyCtrlC:
		return true
	case tcell.KeyEsc:
		g.clearImage()
		g.curQ = nil
		g.showAnswer = false
		g.phase = PhaseBoard
	case tcell.KeyEnter, tcell.KeyRune:
		if key == tcell.KeyRune && r != ' ' {
			return false
		}
		g.showAnswer = !g.showAnswer
		if g.showAnswer {
			g.msg = "showing answer. press space/enter to show question again, esc to return."
		} else {
			g.msg = "press space/enter to reveal answer, esc to return."
		}
	}
	return false
}

func (g *Game) move(dc, dr int) {
	cols := len(g.board.Categories)
	rows := QuestionsPerCategory
	g.cursorCol = (g.cursorCol + dc + cols) % cols
	g.cursorRow = (g.cursorRow + dr + rows) % rows
}

func (g *Game) openSelected() {
	cat := g.board.Categories[g.cursorCol]
	q := cat.Questions[g.cursorRow]
	if q.Picked {
		g.flashMsg("Already taken.")
		return
	}
	g.curQ = q
	g.showAnswer = false
	q.Picked = true
	g.phase = PhaseQuestion
	g.msg = "press space/enter to reveal answer, esc to return."
}

// Regex for score commands: <teamNum><+|-><value>
var scoreCmdRe = regexp.MustCompile(`^([1-8])([\+\-])(\d+)$`)

// applyScoreCommand parses and applies score adjustment
func applyScoreCommand(g *Game, buf string) bool {
	m := scoreCmdRe.FindStringSubmatch(buf)
	if m == nil {
		return false
	}

	idx, _ := strconv.Atoi(m[1])
	sign := m[2]
	val, _ := strconv.Atoi(m[3])

	teamIdx := idx - 1
	if teamIdx < 0 || teamIdx >= len(g.teams) {
		return false
	}

	if sign == "+" {
		g.teams[teamIdx].Score += val
	} else {
		g.teams[teamIdx].Score -= val
	}

	g.flashMsg("adjusted %s: %c%d", g.teams[teamIdx].Name, sign[0], val)
	return true
}

func (g *Game) flashMsg(format string, args ...any) {
	g.msg = fmt.Sprintf(format, args...)
}

// renderImageAfterShow renders image using Kitty protocol
func (g *Game) renderImageAfterShow() {
	if g.imageRenderer == nil || g.curQ == nil || g.curQ.ImagePath == "" {
		return
	}

	w, h := g.s.Size()
	questionAreaY := QuestionAreaY
	questionAreaH := h - questionAreaY - StatusBarHeight
	imageHeight := questionAreaH * ImageHeightRatio / 100

	// define the image area boundaries
	imageAreaY := questionAreaY
	imageAreaHeight := imageHeight
	imageAreaWidth := w - 4 // leave some padding on sides

	// calculate the midpoint of the designated image area
	areaMidX := w / 2
	areaMidY := imageAreaY + imageAreaHeight/2

	imgWidth, imgHeight, err := g.imageRenderer.GetImageBounds(g.curQ.ImagePath)
	if err != nil {
		return
	}

	estimatedCellWidth := imgWidth / 10
	estimatedCellHeight := imgHeight / 20

	// if estimated size exceeds the available area, scale down proportionally
	if estimatedCellWidth > imageAreaWidth {
		scale := float64(imageAreaWidth) / float64(estimatedCellWidth)
		estimatedCellWidth = imageAreaWidth
		estimatedCellHeight = int(float64(estimatedCellHeight) * scale)
	}
	if estimatedCellHeight > imageAreaHeight {
		scale := float64(imageAreaHeight) / float64(estimatedCellHeight)
		estimatedCellHeight = imageAreaHeight
		estimatedCellWidth = int(float64(estimatedCellWidth) * scale)
	}

	// calculate cursor position to center the image's midpoint on the area's midpoint
	cursorX := areaMidX - (estimatedCellWidth / 2)
	cursorY := areaMidY - (estimatedCellHeight / 2)

	if cursorX < 2 {
		cursorX = 2
	}
	if cursorY < imageAreaY {
		cursorY = imageAreaY
	}
	if cursorX+estimatedCellWidth > w-2 { // ensure image doesn't go off right edge
		cursorX = w - 2 - estimatedCellWidth
	}
	if cursorY+estimatedCellHeight > imageAreaY+imageAreaHeight { // ensure image doesn't go off bottom
		cursorY = imageAreaY + imageAreaHeight - estimatedCellHeight
	}

	fmt.Printf("\x1b[s")
	fmt.Printf("\x1b[%d;%dH", cursorY, cursorX)

	imageData, err := g.imageRenderer.RenderImageToString(g.curQ.ImagePath, 0, 0)
	if err == nil && imageData != "" {
		fmt.Print(imageData)
	}

	fmt.Printf("\x1b[u")
}

func (g *Game) clearImage() {
	if g.imageRenderer == nil {
		return
	}

	fmt.Print("\x1b_Ga=d\x1b\\")
}

// renderTextAfterShow renders text using Kitty protocol
func (g *Game) renderTextAfterShow() {
	if g.textToRender == "" {
		return
	}

	maxLineWidth := g.textAreaW/2 - 4
	lines := wrapText(g.textToRender, maxLineWidth)

	lineSpacing := 2
	startY := g.textAreaY + (g.textAreaH-len(lines)*lineSpacing)/2
	if startY < g.textAreaY {
		startY = g.textAreaY
	}

	fmt.Printf("\x1b[s")

	// to stdout
	for i, textLine := range lines {
		lineY := startY + i*lineSpacing
		if lineY >= g.textAreaY+g.textAreaH {
			break
		}

		scaledTextWidth := len(textLine) * 2
		cx := g.textAreaX + g.textAreaW/2 - scaledTextWidth/2
		if cx < g.textAreaX {
			cx = g.textAreaX
		}

		fmt.Printf("\x1b[%d;%dH", lineY, cx)

		// extract colors from tcell style
		fg, _, _ := g.textStyle.Decompose()

		fmt.Print("\x1b[48;5;234m") // dark gray background
		if fg == tcell.ColorLightGreen || fg == tcell.ColorGreen {
			fmt.Print("\x1b[38;5;46m") // bright green for answers
		} else {
			fmt.Print("\x1b[38;5;231m") // bright white for questions
		}

		// kitty text scaling
		fmt.Printf("\x1b]66;s=2;%s\x07", textLine)

		fmt.Print("\x1b[0m")
	}

	fmt.Printf("\x1b[u")

	g.textToRender = ""
}
