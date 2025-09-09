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
	QuestionAreaY      = 7
	CategoryHeight     = 3
	CellHeight         = 7
	MinColumnWidth     = 12
	TeamBaselineOffset = 1
	StatusBarHeight    = 1
	ImageSplitRatio    = 2  // image takes 1/ImageSplitRatio of screen width (for vertical split)
	ImageHeightRatio   = 65 // image takes ImageHeightRatio% of question area height (for horizontal split)
	ImageTextPadding   = 2  // padding between image and text areas
)

// game configuration
const (
	MinTeams             = 2
	MaxTeams             = 8
	QuestionsPerCategory = 5
	ExpectedCategories   = 6
)
