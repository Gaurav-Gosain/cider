package tui

import (
	"image/color"
	"math/rand/v2"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	animFPS           = 20
	animInitialChar   = '.'
	animLabelGap      = " "
	animMaxBirthDelay = time.Second
	animEllipsisSpeed = 8 // frames per ellipsis step
)

var (
	animRunes      = []rune("0123456789abcdefABCDEF~!@#$£€%^&*()+=_")
	ellipsisCycle  = []string{".", "..", "...", ""}
)

var animID int64

// animStepMsg triggers the next animation frame.
type animStepMsg struct{ id int }

// anim is a gradient glitch spinner: a row of cycling characters
// painted with a moving HCL-blended colour ramp, plus a label and an
// animated trailing ellipsis. Frames are pre-rendered up front so each
// tick is just a string concat.
type anim struct {
	id               int
	size             int // number of cycling characters
	label            string
	labelColor       color.Color
	gradA, gradB     color.Color
	startTime        time.Time
	birthOffsets     []time.Duration
	initialized      bool
	step             int
	ellipsisStep     int

	// Pre-rendered frames
	initialFrames [][]string // [frame][charIndex] → styled string
	cyclingFrames [][]string // [frame][charIndex] → styled string
	labelRendered []string   // per-rune styled label
	ellipsisRendered []string // per-frame styled ellipsis
}

func newAnim(size int, gradA, gradB, labelColor color.Color) *anim {
	id := int(atomic.AddInt64(&animID, 1))
	a := &anim{
		id:         id,
		size:       size,
		labelColor: labelColor,
		gradA:      gradA,
		gradB:      gradB,
		startTime:  time.Now(),
	}

	// Random birth offsets for staggered entrance
	a.birthOffsets = make([]time.Duration, size)
	for i := range a.birthOffsets {
		a.birthOffsets[i] = time.Duration(rand.N(int64(animMaxBirthDelay)))
	}

	// Pre-render gradient ramp for cycling colors
	numFrames := size * 2
	ramp := makeGradient(size*3, gradA, gradB, gradA, gradB)

	// Pre-render initial dot frames
	a.initialFrames = make([][]string, numFrames)
	for f := range a.initialFrames {
		a.initialFrames[f] = make([]string, size)
		for j := range size {
			idx := j + f
			if idx >= len(ramp) {
				idx = idx % len(ramp)
			}
			a.initialFrames[f][j] = lipgloss.NewStyle().
				Foreground(ramp[idx]).
				Render(string(animInitialChar))
		}
	}

	// Pre-render cycling character frames
	a.cyclingFrames = make([][]string, numFrames)
	for f := range a.cyclingFrames {
		a.cyclingFrames[f] = make([]string, size)
		for j := range size {
			idx := j + f
			if idx >= len(ramp) {
				idx = idx % len(ramp)
			}
			r := animRunes[rand.IntN(len(animRunes))]
			a.cyclingFrames[f][j] = lipgloss.NewStyle().
				Foreground(ramp[idx]).
				Render(string(r))
		}
	}

	return a
}

func (a *anim) setLabel(label string) {
	a.label = label
	// Pre-render label runes
	a.labelRendered = make([]string, len([]rune(label)))
	for i, r := range []rune(label) {
		a.labelRendered[i] = lipgloss.NewStyle().
			Foreground(a.labelColor).
			Render(string(r))
	}
	// Pre-render ellipsis frames
	a.ellipsisRendered = make([]string, len(ellipsisCycle))
	for i, frame := range ellipsisCycle {
		a.ellipsisRendered[i] = lipgloss.NewStyle().
			Foreground(a.labelColor).
			Render(frame)
	}
}

func (a *anim) start() tea.Cmd {
	return a.tick()
}

func (a *anim) animate(msg animStepMsg) tea.Cmd {
	if msg.id != a.id {
		return nil
	}
	a.step++
	if a.step >= len(a.cyclingFrames) {
		a.step = 0
	}
	if a.initialized && len(a.labelRendered) > 0 {
		a.ellipsisStep++
		if a.ellipsisStep >= animEllipsisSpeed*len(ellipsisCycle) {
			a.ellipsisStep = 0
		}
	} else if !a.initialized && time.Since(a.startTime) >= animMaxBirthDelay {
		a.initialized = true
	}
	return a.tick()
}

func (a *anim) render() string {
	var b strings.Builder
	step := a.step
	if step >= len(a.cyclingFrames) {
		step = 0
	}

	for i := range a.size {
		if !a.initialized && time.Since(a.startTime) < a.birthOffsets[i] {
			// Not yet born: show initial dot
			b.WriteString(a.initialFrames[step][i])
		} else {
			// Cycling character
			b.WriteString(a.cyclingFrames[step][i])
		}
	}

	// Label
	if len(a.labelRendered) > 0 {
		b.WriteString(animLabelGap)
		for _, lr := range a.labelRendered {
			b.WriteString(lr)
		}
		// Animated ellipsis
		if a.initialized {
			frameIdx := a.ellipsisStep / animEllipsisSpeed
			if frameIdx < len(a.ellipsisRendered) {
				b.WriteString(a.ellipsisRendered[frameIdx])
			}
		}
	}

	return b.String()
}

func (a *anim) tick() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(animFPS), func(time.Time) tea.Msg {
		return animStepMsg{id: a.id}
	})
}

// makeGradient creates a color ramp blending between stops using HCL colorspace.
func makeGradient(size int, stops ...color.Color) []color.Color {
	if len(stops) < 2 || size < 2 {
		return nil
	}
	points := make([]colorful.Color, len(stops))
	for i, s := range stops {
		points[i], _ = colorful.MakeColor(s)
	}

	numSeg := len(stops) - 1
	result := make([]color.Color, 0, size)
	segSize := size / numSeg
	remainder := size % numSeg

	for i := range numSeg {
		n := segSize
		if i < remainder {
			n++
		}
		for j := range n {
			t := float64(j) / float64(n)
			c := points[i].BlendHcl(points[i+1], t)
			result = append(result, c)
		}
	}
	return result
}
