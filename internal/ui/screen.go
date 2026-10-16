package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
	"github.com/dmsRosa6/glyph/primitive"

	"github.com/example/zed-cheatsheet/internal/data"
)

const (
	sidebarW = 20 // outer width of the category frame
	boxTop   = 3  // first row of both frames (rows 0-1 are title and search)
	minW     = 64
	minH     = 12
	catColW  = 10 // width of the category column in "All" and search views

	// minActionW is the least room the action column may be squeezed to
	// before the category column is hidden instead.
	minActionW = 24
)

// Screen is a custom Glyph widget (built on mixin.Node, as devguide.md
// describes for custom drawables) that paints the sidebar and the shortcut
// table. It lays itself out from the buffer's size on every frame, so it
// follows terminal resizes without any extra wiring.
//
// The title, search field and footer are ordinary Glyph widgets added to the
// canvas next to it (see app.go).
type Screen struct {
	mixin.FocusableNode // focusable: Tab moves focus between this and the search field

	cats   []data.Category
	search *primitive.TextInput
	chordW int // widest chord in the data set, so the action column lines up

	mu    sync.Mutex
	cat   int // 0 = "All", 1..n = cats[cat-1]
	sel   int // selected row
	top   int // first visible row
	query string
	visH  int // table rows that fit on screen, as of the last Draw
}

// row is one line of the table.
type row struct {
	keys   []string
	action string
	cat    string
}

// NewScreen builds the main widget. search is the TextInput whose live value
// filters the table.
func NewScreen(cats []data.Category, search *primitive.TextInput) (*Screen, error) {
	// Bounds are nominal: Draw sizes itself from the buffer. Keeping them
	// tiny means the root container never flags this widget as out of bounds.
	node, err := mixin.NewNode(geom.NewBounds(0, 0, 1, 1), framework.Anchor{},
		framework.Style{Bg: core.Transparent, Fg: core.Transparent}, 0, "Screen")
	if err != nil {
		return nil, err
	}

	s := &Screen{FocusableNode: mixin.NewFocusableNode(node), cats: cats, search: search}
	for _, c := range cats {
		for _, sh := range c.Shortcuts {
			if w := chordWidth(sh.Keys); w > s.chordW {
				s.chordW = w
			}
		}
	}
	s.bindKeys()
	return s, nil
}

// bindKeys wires browse-mode keys. They only fire while the Screen has focus;
// while the search field has focus, typing goes to it instead.
func (s *Screen) bindKeys() {
	act := func(fn func()) mixin.FocusableActionFunc {
		return func(mixin.FocusableActionContext) (bool, error) { fn(); return true, nil }
	}
	s.BindAction(framework.KeyLeft, act(func() { s.CycleCategory(-1) }))
	s.BindAction(framework.KeyRight, act(func() { s.CycleCategory(+1) }))
	s.BindAction(framework.KeyHome, act(func() { s.MoveTo(0) }))
	s.BindAction(framework.KeyEnd, act(func() { s.MoveTo(-1) }))
	// Up/Down are global (see app.go) so they also work from the search field.

	s.BindAction(framework.KeyRune, func(a mixin.FocusableActionContext) (bool, error) {
		return s.handleRune(a.Event().Rune), nil
	})
}

// handleRune implements the vim-style browse keys; it reports whether r was
// one of them.
func (s *Screen) handleRune(r rune) bool {
	switch r {
	case 'j':
		s.Move(+1)
	case 'k':
		s.Move(-1)
	case 'd':
		s.MovePage(+1)
	case 'u':
		s.MovePage(-1)
	case 'g':
		s.MoveTo(0)
	case 'G':
		s.MoveTo(-1)
	case 'l', ']':
		s.CycleCategory(+1)
	case 'h', '[':
		s.CycleCategory(-1)
	case '/':
		s.Context().Nav().Next() // focus the search field
	default:
		if r >= '1' && r <= '9' {
			s.SetCategory(int(r - '1'))
			return true
		}
		return false
	}
	return true
}

// ---- state changes (called from key handlers) ----

// Move shifts the selected row by delta, clamped to the table.
func (s *Screen) Move(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.rowsLocked(strings.TrimSpace(s.search.Value())))
	s.sel = clamp(s.sel+delta, 0, n-1)
}

// MoveTo selects row pos (0 = first); a negative pos selects the last row.
func (s *Screen) MoveTo(pos int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.rowsLocked(strings.TrimSpace(s.search.Value())))
	if pos < 0 {
		pos = n - 1
	}
	s.sel = clamp(pos, 0, n-1)
}

// MovePage moves the selection half a screen up or down.
func (s *Screen) MovePage(dir int) {
	s.mu.Lock()
	step := max(1, s.visH/2)
	s.mu.Unlock()
	s.Move(dir * step)
}

// SetCategory jumps to category i (0 = All); out-of-range indexes are ignored.
func (s *Screen) SetCategory(i int) {
	s.mu.Lock()
	if i < 0 || i > len(s.cats) {
		s.mu.Unlock()
		return
	}
	s.cat = i
	s.sel, s.top = 0, 0
	s.mu.Unlock()
	s.ClearSearch()
}

// CycleCategory moves to the next (dir > 0) or previous (dir < 0) category,
// wrapping around, and clears any active search so the change is visible.
func (s *Screen) CycleCategory(dir int) {
	s.mu.Lock()
	n := len(s.cats) + 1
	s.cat = ((s.cat+dir)%n + n) % n
	s.sel, s.top = 0, 0
	s.mu.Unlock()
	s.ClearSearch()
}

// ClearSearch empties the search field.
func (s *Screen) ClearSearch() {
	if s.search.Len() > 0 {
		s.search.SetValue("")
	}
}

// ---- data ----

// rowsLocked returns the rows for the current view. Caller holds s.mu.
func (s *Screen) rowsLocked(query string) []row {
	if query != "" {
		ms := data.Search(s.cats, query)
		out := make([]row, len(ms))
		for i, m := range ms {
			out[i] = row{m.Keys, m.Action, m.Category}
		}
		return out
	}
	var out []row
	for i, c := range s.cats {
		if s.cat != 0 && s.cat-1 != i {
			continue
		}
		for _, sh := range c.Shortcuts {
			r := row{keys: sh.Keys, action: sh.Action}
			if s.cat == 0 {
				r.cat = c.Name
			}
			out = append(out, r)
		}
	}
	return out
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ---- drawing ----

// Draw paints sidebar and table. It is the only place that touches the
// buffer, and the only reader of the search field's value besides Move.
func (s *Screen) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := s.ComputedPos()
	p := pen{buf: buf, ox: vec.X + pos.X, oy: vec.Y + pos.Y}
	W, H := buf.W-p.ox, buf.H-p.oy

	p.fill(0, 0, W, H, colBg)

	if W < minW || H < minH {
		msg := fmt.Sprintf("Terminal too small (%dx%d) - need at least %dx%d", W, H, minW, minH)
		p.text(2, 4, W, msg, colBg, colWarn)
		return
	}

	query := strings.TrimSpace(s.search.Value())

	// Search row: the label lights up while the field has focus.
	searching := s.search.IsFocused()
	labelFg, hint := colDim, "press / to search"
	switch {
	case searching:
		labelFg, hint = colAccent, "Enter: browse results   Esc: clear"
	case query != "":
		hint = "/ edit search   Esc clear"
	}
	p.text(2, 1, searchX-1, "Search", colBg, labelFg)
	p.text(searchX+searchWidth+2, 1, W-1, hint, colBg, colDim)

	s.mu.Lock()
	defer s.mu.Unlock()

	if query != s.query { // the filter changed: start from the top
		s.query = query
		s.sel, s.top = 0, 0
	}
	rows := s.rowsLocked(query)

	boxH := H - boxTop - 1 // leave the last row for the footer
	s.drawSidebar(p, 1, boxTop, sidebarW, boxH, query)
	s.drawTable(p, sidebarW+2, boxTop, W-sidebarW-3, boxH, rows, query)
}

func (s *Screen) drawSidebar(p pen, x, y, w, h int, query string) {
	p.frame(x, y, w, h, "Categories", colBorder, colDim)

	counts := make([]int, len(s.cats)+1)
	for i, c := range s.cats {
		counts[i+1] = len(c.Shortcuts)
		counts[0] += len(c.Shortcuts)
	}
	if query != "" { // while searching, show matches per category
		for i := range counts {
			counts[i] = 0
		}
		for _, m := range data.Search(s.cats, query) {
			counts[0]++
			for i, c := range s.cats {
				if c.Name == m.Category {
					counts[i+1]++
				}
			}
		}
	}

	for i := 0; i <= len(s.cats); i++ {
		yy := y + 1 + i
		if yy >= y+h-1 {
			break
		}
		name := "All"
		if i > 0 {
			name = s.cats[i-1].Name
		}

		active := i == s.cat && query == ""
		bg, fg, mark := colPanel, colText, ' '
		switch {
		case active:
			bg, fg, mark = colSelBg, colAccent, '▌'
		case query != "" && counts[i] == 0:
			fg = colDim
		}

		p.fill(x+1, yy, w-2, 1, bg)
		p.set(x+1, yy, mark, bg, colAccent)
		p.text(x+3, yy, x+w-2, name, bg, fg)

		num := fmt.Sprint(counts[i])
		numFg := colDim
		if active {
			numFg = colAccent
		}
		p.text(x+w-2-len(num), yy, x+w-1, num, bg, numFg)
	}
}

func (s *Screen) drawTable(p pen, x, y, w, h int, rows []row, query string) {
	title := "All shortcuts"
	switch {
	case query != "":
		title = fmt.Sprintf("Search “%s”", query)
	case s.cat > 0:
		title = s.cats[s.cat-1].Name
	}
	border := colAccentDim
	if s.IsFocused() {
		border = colAccent // the table has keyboard focus
	}
	p.frame(x, y, w, h, title, border, colAccent)

	innerL, innerR := x+2, x+w-2 // text may use [innerL, innerR)
	visH := h - 4                // minus borders, header and rule
	s.visH = visH
	if visH < 1 {
		return
	}

	// Keep the selection on screen.
	s.sel = clamp(s.sel, 0, len(rows)-1)
	if s.sel < s.top {
		s.top = s.sel
	}
	if s.sel >= s.top+visH {
		s.top = s.sel - visH + 1
	}
	s.top = clamp(s.top, 0, len(rows)-visH)

	showCat := query != "" || s.cat == 0
	actionX := innerL + s.chordW + 2
	actionMax := innerR
	if showCat {
		if innerR-catColW-1-actionX >= minActionW {
			actionMax = innerR - catColW - 1
		} else {
			showCat = false // narrow terminal: the action text matters more
		}
	}

	// Column headers and rule.
	p.text(innerL, y+1, innerR, "SHORTCUT", colPanel, colDim)
	p.text(actionX, y+1, actionMax, "ACTION", colPanel, colDim)
	if showCat {
		p.text(innerR-catColW, y+1, innerR, "CATEGORY", colPanel, colDim)
	}
	for xx := x + 1; xx < x+w-1; xx++ {
		p.set(xx, y+2, '─', colPanel, colBorder)
	}

	if len(rows) == 0 {
		p.text(innerL, y+4, innerR, "No shortcuts match.", colPanel, colDim)
		p.text(innerL, y+5, innerR, "Try an action (“split”), a key (“shift”) or a category.", colPanel, colDim)
		return
	}

	for i := 0; i < visH && s.top+i < len(rows); i++ {
		idx := s.top + i
		r := rows[idx]
		yy := y + 3 + i
		selected := idx == s.sel

		bg, actFg, capBg := colPanel, colText, colCap
		if selected {
			bg, capBg = colSelBg, colCapSel
			actFg = colCapText
		}
		p.fill(x+1, yy, w-2, 1, bg)
		if selected {
			p.set(x+1, yy, '▌', bg, colAccent)
		}

		p.chord(innerL, yy, innerR, r.keys, bg, capBg)

		p.text(actionX, yy, actionMax, fit(r.action, actionMax-actionX), bg, actFg)
		if showCat {
			p.text(innerR-catColW, yy, innerR, r.cat, bg, colDim)
		}
	}

	// Scrollbar thumb on the right border, plus a position read-out.
	if len(rows) > visH {
		thumb := max(1, visH*visH/len(rows))
		offset := 0
		if span := len(rows) - visH; span > 0 {
			offset = s.top * (visH - thumb) / span
		}
		for i := 0; i < thumb; i++ {
			p.set(x+w-1, y+3+offset+i, '┃', colPanel, colAccent)
		}
	}
	pos := fmt.Sprintf(" %d/%d ", s.sel+1, len(rows))
	p.text(x+w-2-len(pos), y+h-1, x+w-1, pos, colPanel, colDim)
}

// fit shortens s to at most w cells, ending in an ellipsis when it is cut.
func fit(s string, w int) string {
	r := []rune(s)
	if w <= 0 || len(r) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	return string(r[:w-1]) + "…"
}
