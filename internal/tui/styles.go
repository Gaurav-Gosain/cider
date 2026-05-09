package tui

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Color palette - actual charmtone hex values
var (
	primary   = lipgloss.Color("#6B50FF") // Charple
	secondary = lipgloss.Color("#FF60FF") // Dolly
	tertiary  = lipgloss.Color("#68FFD6") // Bok
	bgBase    = lipgloss.Color("#201F26") // Pepper
	fgBase    = lipgloss.Color("#DFDBDD") // Ash
	fgMuted   = lipgloss.Color("#858392") // Squid
	fgSubtle  = lipgloss.Color("#605F6B") // Oyster
	red       = lipgloss.Color("#E7766C") // Coral
	border    = lipgloss.Color("#2E2D35") // Charcoal (subtle border)
)

// Gradient colors for the spinner animation
var (
	gradColorA color.Color = color.RGBA{R: 0x6B, G: 0x50, B: 0xFF, A: 0xFF} // Charple
	gradColorB color.Color = color.RGBA{R: 0xFF, G: 0x60, B: 0xFF, A: 0xFF} // Dolly
	labelColor color.Color = color.RGBA{R: 0xDF, G: 0xDB, B: 0xDD, A: 0xFF} // Ash
)

// Message prefixes: render an empty styled string to harvest the ANSI
// prefix codes, then prepend them line-by-line.
var (
	userBlurredPrefix = lipgloss.NewStyle().
				PaddingLeft(1).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(primary).
				Render()

	assistantBlurredPrefix = lipgloss.NewStyle().
				PaddingLeft(3).
				Render()
)

// Text styles
var (
	roleLabelUser = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			PaddingLeft(2)

	roleLabelAssistant = lipgloss.NewStyle().
				Foreground(fgBase).
				Bold(true).
				PaddingLeft(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	mutedStyle = lipgloss.NewStyle().
			Foreground(fgMuted)

	subtleStyle = lipgloss.NewStyle().
			Foreground(fgSubtle)

	headerCharm = lipgloss.NewStyle().
			Foreground(fgMuted)

	headerTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary)

	headerDiagStyle = lipgloss.NewStyle().
			Foreground(fgSubtle)

	headerDetail = lipgloss.NewStyle().
			Foreground(fgMuted)

	separatorStyle = lipgloss.NewStyle().
			Foreground(border)

	// Rounded border box for assistant responses
	responseBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			PaddingLeft(1).
			PaddingRight(1)
)

// separator returns a thin horizontal line for visual separation.
func separator(width int) string {
	return separatorStyle.Render(strings.Repeat("─", width))
}

func textareaStyles() textarea.Styles {
	base := lipgloss.NewStyle().Foreground(fgBase)
	return textarea.Styles{
		Focused: textarea.StyleState{
			Base:             base,
			Text:             base,
			LineNumber:       base.Foreground(fgSubtle),
			CursorLine:       base,
			CursorLineNumber: base.Foreground(fgSubtle),
			Placeholder:      base.Foreground(fgSubtle),
			Prompt:           base.Foreground(tertiary),
		},
		Blurred: textarea.StyleState{
			Base:             base,
			Text:             base.Foreground(fgMuted),
			LineNumber:       base.Foreground(fgMuted),
			CursorLine:       base,
			CursorLineNumber: base.Foreground(fgMuted),
			Placeholder:      base.Foreground(fgSubtle),
			Prompt:           base.Foreground(fgMuted),
		},
		Cursor: textarea.CursorStyle{
			Color: secondary,
			Shape: tea.CursorBlock,
			Blink: true,
		},
	}
}
