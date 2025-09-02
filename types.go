package main

type Question struct {
	Category  string
	Value     int
	Q         string
	A         string
	ImagePath string // optional
	Picked    bool
}

type Category struct {
	Name      string
	Questions []*Question
}

type Board struct {
	Categories []*Category // len == 6
}

type Team struct {
	Name  string
	Score int
}

// game phases
const (
	PhaseSetupNumTeams = iota
	PhaseSetupTeamNames
	PhaseBoard
	PhaseQuestion
)

// UI constants
const (
	QuestionAreaY = 7
)
