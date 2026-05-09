package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gaurav-Gosain/cider/pkg/fm"
)

const (
	maxContentWidth = 120
	editorHeight    = 6 // textarea rows + separator + bottom margin
	statusHeight    = 1
	headerHeight    = 1
	headerGap       = 1 // gap between header and chat
	appleLogo       = "\uF8FF"
)

type chatMessage struct {
	role    string
	content string
	// cached rendered output (invalidated on width change)
	rendered      string
	renderedWidth int
}

type streamChunkMsg struct{ text string }
type streamDoneMsg struct{ err error }

// Model is the bubbletea model for the cider chat TUI.
type Model struct {
	session      *fm.Session
	instructions string

	textarea textarea.Model
	anim     *anim

	messages     []*chatMessage
	generating   bool
	streaming    string
	err          error
	scrollOffset int

	width  int
	height int

	program  *tea.Program
	cancelFn context.CancelFunc
}

func New(session *fm.Session, instructions string) *Model {
	ta := textarea.New()
	ta.Placeholder = "Ask anything..."
	ta.ShowLineNumbers = false
	ta.CharLimit = -1
	ta.MaxHeight = 5
	ta.SetHeight(3)
	ta.SetVirtualCursor(false)
	ta.SetStyles(textareaStyles())
	// Rebind newline to shift+enter (we use enter to send)
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("shift+enter", "ctrl+j"))
	ta.Focus()

	a := newAnim(15, gradColorA, gradColorB, labelColor)
	a.setLabel("Thinking")

	return &Model{
		session:      session,
		instructions: instructions,
		textarea:     ta,
		anim:         a,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(m.width - 4) // minus left/right padding
		// Invalidate all caches on resize
		for _, msg := range m.messages {
			msg.rendered = ""
			msg.renderedWidth = 0
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	case tea.MouseWheelMsg:
		mouse := msg.Mouse()
		switch mouse.Button {
		case tea.MouseWheelUp:
			m.scrollOffset += 3
		case tea.MouseWheelDown:
			m.scrollOffset -= 3
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
		}
		return m, nil

	case streamChunkMsg:
		// FM stream yields cumulative snapshots, not deltas: replace, don't append.
		m.streaming = msg.text
		m.scrollOffset = 0
		return m, nil

	case streamDoneMsg:
		m.generating = false
		if msg.err != nil && msg.err != context.Canceled {
			m.err = msg.err
		}
		if m.streaming != "" {
			m.messages = append(m.messages, &chatMessage{role: "assistant", content: m.streaming})
		}
		m.streaming = ""
		m.cancelFn = nil
		return m, nil

	case animStepMsg:
		if m.generating {
			cmds = append(cmds, m.anim.animate(msg))
		}
	}

	if !m.generating {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if m.generating && m.cancelFn != nil {
			m.cancelFn()
			return m, nil
		}
		return m, tea.Quit

	case "esc":
		if m.generating && m.cancelFn != nil {
			m.cancelFn()
			return m, nil
		}
		return m, tea.Quit

	case "enter":
		if m.generating {
			return m, nil
		}
		text := strings.TrimSpace(m.textarea.Value())
		if text == "" {
			return m, nil
		}
		m.textarea.Reset()
		m.messages = append(m.messages, &chatMessage{role: "user", content: text})
		m.generating = true
		m.streaming = ""
		m.scrollOffset = 0
		m.err = nil
		// Create new anim for fresh entrance effect
		m.anim = newAnim(15, gradColorA, gradColorB, labelColor)
		m.anim.setLabel("Thinking")
		return m, tea.Batch(m.anim.start(), m.startStream(text))

	case "pgup":
		m.scrollOffset += 10
		return m, nil

	case "pgdown":
		m.scrollOffset -= 10
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		return m, nil
	}

	if !m.generating {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View composes the full screen: header → chat → separator → editor → status
func (m *Model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("")
	}

	chatHeight := max(m.height-headerHeight-headerGap-editorHeight-statusHeight, 1)

	var lines []string

	// 1. Header (1 line)
	lines = append(lines, m.viewHeader())

	// 2. Gap (1 line)
	lines = append(lines, "")

	// 3. Chat area
	chatLines := m.viewChatLines(chatHeight)
	lines = append(lines, chatLines...)

	// 4. Editor area (includes separator at top)
	editorStartLine := len(lines) + 2 // +2 for separator line + blank line
	editorLines := m.viewEditorLines()
	lines = append(lines, editorLines...)

	// 5. Status bar (1 line)
	lines = append(lines, m.viewStatus())

	// Trim to exact terminal height
	if len(lines) > m.height {
		lines = lines[:m.height]
	}

	v := tea.NewView(strings.Join(lines, "\n"))
	v.AltScreen = true
	v.BackgroundColor = bgBase
	v.MouseMode = tea.MouseModeCellMotion

	// Position the real cursor from the textarea
	if !m.generating {
		if c := m.textarea.Cursor(); c != nil {
			c.Position.X += 2 // account for "  " left indent
			c.Position.Y += editorStartLine
			// Clamp cursor within editor bounds
			maxY := editorStartLine + editorHeight - 2
			if c.Position.Y > maxY {
				c.Position.Y = maxY
			}
			v.Cursor = c
		}
	}

	return v
}

// viewHeader renders the compact header: logo + diagonal fill + details
func (m *Model) viewHeader() string {
	logo := headerCharm.Render("  "+appleLogo) + " " + headerTitle.Render("Cider") + " "

	// Right-side details
	msgCount := headerDetail.Render(fmt.Sprintf("%d messages", len(m.messages)))
	var status string
	if m.generating {
		status = headerDetail.Render(" · ") + lipgloss.NewStyle().Foreground(secondary).Render("generating")
	}
	right := msgCount + status + "  "

	logoW := lipgloss.Width(logo)
	rightW := lipgloss.Width(right)
	fillW := max(m.width-logoW-rightW, 0)
	fill := headerDiagStyle.Render(strings.Repeat("╱", fillW))

	return logo + fill + right
}

// contentWidth returns the capped content width for messages.
func (m *Model) contentWidth() int {
	w := m.width - 5 // account for prefix (border + padding)
	return max(min(w, maxContentWidth), 20)
}

// viewChatLines returns exactly chatHeight lines for the chat area.
func (m *Model) viewChatLines(chatHeight int) []string {
	cw := m.contentWidth()

	// Build all rendered message blocks (cached)
	var blocks []string

	// Welcome message if no messages
	if len(m.messages) == 0 && !m.generating {
		welcome := mutedStyle.PaddingLeft(3).Render("Send a message to start chatting.")
		blocks = append(blocks, welcome)
	}

	// Render completed messages (with caching)
	for _, msg := range m.messages {
		blocks = append(blocks, m.renderMsg(msg, cw))
	}

	if m.streaming != "" {
		label := roleLabelAssistant.Render("Cider")
		rendered := renderMarkdown(m.streaming, cw-4)
		boxed := responseBorder.Width(cw).Render(rendered)
		content := applyPrefix(boxed, assistantBlurredPrefix)
		blocks = append(blocks, label+"\n"+content)
	} else if m.generating {
		label := roleLabelAssistant.Render("Cider")
		spinView := assistantBlurredPrefix + m.anim.render()
		blocks = append(blocks, label+"\n"+spinView)
	}

	// Join blocks with 2 blank lines between messages for breathing room
	fullContent := strings.Join(blocks, "\n\n\n")
	allLines := strings.Split(fullContent, "\n")
	totalLines := len(allLines)

	maxScroll := max(totalLines-chatHeight, 0)
	m.scrollOffset = max(min(m.scrollOffset, maxScroll), 0)

	// Visible window, anchored to bottom.
	end := min(totalLines-m.scrollOffset, totalLines)
	start := max(end-chatHeight, 0)

	visible := make([]string, 0, chatHeight)
	// Pad top if content shorter than viewport
	for i := 0; i < chatHeight-(end-start); i++ {
		visible = append(visible, "")
	}
	visible = append(visible, allLines[start:end]...)

	return visible
}

// viewEditorLines renders the editor area: separator + textarea + bottom margin.
func (m *Model) viewEditorLines() []string {
	lines := make([]string, 0, editorHeight)

	// Thin separator line
	lines = append(lines, separator(m.width))

	if m.generating {
		// Show spinner below separator while generating
		lines = append(lines, "  "+m.anim.render())
		for len(lines) < editorHeight {
			lines = append(lines, "")
		}
		return lines
	}

	// Blank line above textarea for breathing room
	lines = append(lines, "")

	taLines := strings.Split(m.textarea.View(), "\n")
	for i, line := range taLines {
		taLines[i] = "  " + line
	}
	lines = append(lines, taLines...)

	// Pad/trim to exactly editorHeight
	for len(lines) < editorHeight {
		lines = append(lines, "")
	}
	if len(lines) > editorHeight {
		lines = lines[:editorHeight]
	}
	return lines
}

// viewStatus renders the bottom status bar.
func (m *Model) viewStatus() string {
	if m.err != nil {
		return "  " + errorStyle.Render("error: "+m.err.Error())
	}

	if m.generating {
		return mutedStyle.Render("  esc") + subtleStyle.Render(" cancel")
	}

	return mutedStyle.Render("  enter") + subtleStyle.Render(" send  ") +
		mutedStyle.Render("⇧ enter") + subtleStyle.Render(" newline  ") +
		mutedStyle.Render("esc") + subtleStyle.Render(" quit")
}

// renderMsg renders a message with role label and per-line prefix.
// Results are cached per message to avoid re-rendering unchanged content.
func (m *Model) renderMsg(msg *chatMessage, width int) string {
	if msg.rendered != "" && msg.renderedWidth == width {
		return msg.rendered
	}

	var result string
	switch msg.role {
	case "user":
		label := roleLabelUser.Render("You")
		content := applyPrefix(msg.content, userBlurredPrefix)
		result = label + "\n" + content

	case "assistant":
		label := roleLabelAssistant.Render("Cider")
		rendered := renderMarkdown(msg.content, width-4) // minus border + padding
		boxed := responseBorder.Width(width).Render(rendered)
		content := applyPrefix(boxed, assistantBlurredPrefix)
		result = label + "\n" + content

	default:
		result = msg.content
	}

	msg.rendered = result
	msg.renderedWidth = width
	return result
}

func applyPrefix(content, prefix string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func (m *Model) startStream(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		m.cancelFn = cancel

		chunkCh, errCh := m.session.StreamResponse(ctx, prompt)

		for chunk := range chunkCh {
			if m.program != nil {
				m.program.Send(streamChunkMsg{text: chunk})
			}
		}

		var streamErr error
		select {
		case err := <-errCh:
			streamErr = err
		default:
		}

		cancel()
		return streamDoneMsg{err: streamErr}
	}
}

func Run(session *fm.Session, instructions string) error {
	m := New(session, instructions)
	p := tea.NewProgram(m)
	m.program = p
	_, err := p.Run()
	return err
}
