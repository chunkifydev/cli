package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Base                = lipgloss.NewStyle()
	TableSpacing        = Base.Padding(0, 1) //.MarginBottom(1)
	Header              = Base.Foreground(GrayColor)
	DefaultText         = Base.Foreground(TextColor)
	Border              = Base.Foreground(BorderColor)
	Id                  = Base.Foreground(GrayColor)
	Center              = Base.Align(lipgloss.Center)
	Right               = Base.AlignHorizontal(lipgloss.Right)
	Debug               = Base.Foreground(GrayColor)
	Error               = Base.Foreground(ErrorColor)
	Important           = Base.Foreground(GreenColor)
	Working             = Base.Foreground(WorkingColor)
	Warning             = Base.Foreground(WarningColor)
	Hint                = Base.Foreground(HintForegroundColor).Background(HintBackgroundColor)
	BorderColor         = lipgloss.AdaptiveColor{Light: "#dddddd", Dark: "#353535"}
	TextColor           = lipgloss.AdaptiveColor{Light: "#2b2b2b", Dark: "#d2d2d2"}
	GrayColor           = lipgloss.AdaptiveColor{Light: "#373636", Dark: "#949494"}
	ErrorColor          = lipgloss.AdaptiveColor{Light: "#c63f48", Dark: "#f58c93"}
	WorkingColor        = lipgloss.AdaptiveColor{Light: "#d7b237", Dark: "#ffe89b"}
	GreenColor          = lipgloss.AdaptiveColor{Light: "#309352", Dark: "#37b965"}
	WarningColor        = lipgloss.AdaptiveColor{Light: "#d7b237", Dark: "#f4cf56"}
	HintBackgroundColor = lipgloss.AdaptiveColor{Light: "#fff", Dark: "#000"}
	HintForegroundColor = lipgloss.AdaptiveColor{Light: "#6b591f", Dark: "#f4cf56"}

	Divider = Base.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(BorderColor).
		MarginBottom(1).
		Render
)
