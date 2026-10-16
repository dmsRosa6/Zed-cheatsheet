package ui

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/primitive"

	"github.com/example/zed-cheatsheet/internal/data"
)

// pen is a tiny drawing helper over a core.Buffer. All coordinates are
// relative to (ox, oy), which is where the owning widget sits on the canvas.
type pen struct {
	buf    *core.Buffer
	ox, oy int
}

func (p pen) set(x, y int, r rune, bg, fg core.Color) {
	p.buf.Set(p.ox+x, p.oy+y, r, bg, fg)
}

// text writes s starting at x, never drawing at or beyond maxX, and returns
// the x just after the last rune drawn.
func (p pen) text(x, y, maxX int, s string, bg, fg core.Color) int {
	for _, r := range s {
		if x >= maxX {
			break
		}
		p.set(x, y, r, bg, fg)
		x++
	}
	return x
}

func (p pen) fill(x, y, w, h int, bg core.Color) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			p.set(xx, yy, ' ', bg, colText)
		}
	}
}

// frame draws a rounded border (using Glyph's built-in primitive.Rounded
// runes) with an optional title set into the top edge.
func (p pen) frame(x, y, w, h int, title string, border, titleFg core.Color) {
	if w < 2 || h < 2 {
		return
	}
	bs := primitive.Rounded
	p.fill(x, y, w, h, colPanel)
	for xx := x + 1; xx < x+w-1; xx++ {
		p.set(xx, y, bs.Horizontal, colPanel, border)
		p.set(xx, y+h-1, bs.Horizontal, colPanel, border)
	}
	for yy := y + 1; yy < y+h-1; yy++ {
		p.set(x, yy, bs.Vertical, colPanel, border)
		p.set(x+w-1, yy, bs.Vertical, colPanel, border)
	}
	p.set(x, y, bs.TopLeft, colPanel, border)
	p.set(x+w-1, y, bs.TopRight, colPanel, border)
	p.set(x, y+h-1, bs.BottomLeft, colPanel, border)
	p.set(x+w-1, y+h-1, bs.BottomRight, colPanel, border)

	if title != "" {
		p.text(x+2, y, x+w-2, " "+title+" ", colPanel, titleFg)
	}
}

// keyLabel prettifies a key name for display on a key cap.
func keyLabel(k string) string {
	switch k {
	case "Up":
		return "↑"
	case "Down":
		return "↓"
	case "Left":
		return "←"
	case "Right":
		return "→"
	}
	return k
}

// chordWidth is the on-screen width of a chord drawn as key caps joined by
// "+": every cap is its label padded by one space on each side.
func chordWidth(keys []string) int {
	w := 0
	for i, k := range keys {
		switch {
		case k == data.Then:
			w += len(thenLabel)
		case i > 0 && keys[i-1] != data.Then:
			w += 1 + len([]rune(keyLabel(k))) + 2 // "+" joiner and the cap
		default:
			w += len([]rune(keyLabel(k))) + 2
		}
	}
	return w
}

const thenLabel = " then "

// chord draws key caps for keys at (x, y), clipped at maxX, and returns the
// x after the last cell. Simultaneous keys are joined with "+", the two steps
// of a sequence with a dim "then".
func (p pen) chord(x, y, maxX int, keys []string, rowBg, capBg core.Color) int {
	for i, k := range keys {
		switch {
		case k == data.Then:
			x = p.text(x, y, maxX, thenLabel, rowBg, colDim)
			continue
		case i > 0 && keys[i-1] != data.Then:
			x = p.text(x, y, maxX, "+", rowBg, colDim)
		}
		x = p.text(x, y, maxX, " "+keyLabel(k)+" ", capBg, colCapText)
	}
	return x
}
