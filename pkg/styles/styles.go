package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base style that all other styles inherit from
	Base = lipgloss.NewStyle()

	// TableSpacing adds padding to table cells
	TableSpacing = Base.Padding(0, 1) //.MarginBottom(1)

	// Header style for table headers with gray color
	Header = Base.Foreground(GrayColor)

	// DefaultText style for regular text content
	DefaultText = Base.Foreground(TextColor)

	// Border style for table borders
	Border = Base.Foreground(BorderColor)

	// Id style for displaying IDs in gray
	Id = Base.Foreground(GrayColor)

	// Center style for center-aligned content
	Center = Base.Align(lipgloss.Center)

	// Right style for right-aligned content
	Right = Base.AlignHorizontal(lipgloss.Right)

	// Debug style for debug messages in gray
	Debug = Base.Foreground(GrayColor)

	// Error style for error messages in red
	Error = Base.Foreground(ErrorColor)

	// Important style for highlighting important text in green
	Important = Base.Foreground(GreenColor)

	// Working style for in-progress states
	Working = Base.Foreground(WorkingColor)

	// Warning style for warning messages
	Warning = Base.Foreground(WarningColor)

	// Starting style for starting messages
	Pending = Base.Foreground(DarkGrayColor)

	// Hint style with custom foreground and background colors
	Hint = Base.Foreground(HintForegroundColor).Background(HintBackgroundColor)

	// Color definitions for light and dark themes
	BorderColor         = lipgloss.AdaptiveColor{Light: "#dddddd", Dark: "#353535"}
	TextColor           = lipgloss.AdaptiveColor{Light: "#2b2b2b", Dark: "#d2d2d2"}
	GrayColor           = lipgloss.AdaptiveColor{Light: "#373636", Dark: "#949494"}
	DarkGrayColor       = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#555555"}
	ErrorColor          = lipgloss.AdaptiveColor{Light: "#c63f48", Dark: "#f58c93"}
	WorkingColor        = lipgloss.AdaptiveColor{Light: "#d7b237", Dark: "#ffe89b"}
	GreenColor          = lipgloss.AdaptiveColor{Light: "#309352", Dark: "#37b965"}
	WarningColor        = lipgloss.AdaptiveColor{Light: "#d7b237", Dark: "#f4cf56"}
	HintBackgroundColor = lipgloss.AdaptiveColor{Light: "#fff", Dark: "#000"}
	HintForegroundColor = lipgloss.AdaptiveColor{Light: "#6b591f", Dark: "#f4cf56"}

	// Divider renders a bottom border with margin
	Divider = Base.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(BorderColor).
		MarginBottom(1).
		Render
)
