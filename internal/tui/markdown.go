package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
)

func renderMarkdown(content string, width int) string {
	if width < 10 {
		width = 10
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(styles.TokyoNightStyleConfig),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}

	result, err := r.Render(content)
	if err != nil {
		return content
	}

	// Glamour adds leading/trailing newlines - strip them
	result = strings.TrimLeft(result, "\n")
	result = strings.TrimRight(result, "\n ")

	return result
}
