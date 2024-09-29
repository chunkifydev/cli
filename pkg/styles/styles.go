package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Base         = lipgloss.NewStyle()
	TableSpacing = Base.Padding(0, 1) //.MarginBottom(1)
	Header       = Base.Foreground(GrayColor)
	DefaultText  = Base.Foreground(TextColor)
	Border       = Base.Foreground(BorderColor)
	Id           = Base.Foreground(GrayColor)
	Center       = Base.Align(lipgloss.Center)
	Right        = Base.AlignHorizontal(lipgloss.Right)
	Debug        = Base.Foreground(GrayColor)
	Error        = Base.Foreground(ErrorColor)
	Important    = Base.Foreground(GreenColor)
	Working      = Base.Foreground(WorkingColor)

	BorderColor  = lipgloss.Color("#353535")
	TextColor    = lipgloss.Color("#d2d2d2")
	GrayColor    = lipgloss.Color("#949494")
	ErrorColor   = lipgloss.Color("#f58c93")
	WorkingColor = lipgloss.Color("#ffe89b")
	GreenColor   = lipgloss.Color("#73F59F")

	Divider = Base.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(BorderColor).
		MarginBottom(1).
		Render
)
