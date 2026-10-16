package ui

import (
	"strings"
	"testing"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"

	"github.com/example/zed-cheatsheet/internal/data"
)

func newTestScreen(t *testing.T) (*Screen, *primitive.TextInput) {
	t.Helper()
	in, err := primitive.NewTextInput(&geom.Point{X: searchX, Y: 1}, searchWidth, primitive.TextInputConfig{
		Style: framework.Style{Bg: colInputBg, Fg: colText},
	})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewScreen(data.All(), in)
	if err != nil {
		t.Fatal(err)
	}
	return s, in
}

// renderLines draws the screen into a w×h buffer and returns it as text lines.
func renderLines(s *Screen, w, h int) []string {
	buf := core.NewBuffer(w, h, colText, colBg)
	s.Draw(buf, geom.Vector{})
	lines := make([]string, h)
	for y := 0; y < h; y++ {
		var b strings.Builder
		for x := 0; x < w; x++ {
			b.WriteRune(buf.Get(x, y).Ch)
		}
		lines[y] = b.String()
	}
	return lines
}

func screenText(s *Screen, w, h int) string { return strings.Join(renderLines(s, w, h), "\n") }

func TestDefaultViewListsEverything(t *testing.T) {
	s, _ := newTestScreen(t)
	out := screenText(s, 100, 30)

	for _, want := range []string{"Categories", "All shortcuts", "Open settings", "SHORTCUT", "CATEGORY"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// Sidebar shows every category with its count.
	for _, want := range []string{"Zed", "File", "Edit", "Selection", "View", "Go", "56"} {
		if !strings.Contains(out, want) {
			t.Errorf("sidebar missing %q", want)
		}
	}
}

func TestCycleCategoryFiltersAndWraps(t *testing.T) {
	s, _ := newTestScreen(t)

	s.CycleCategory(+1) // All -> Zed
	out := screenText(s, 100, 30)
	if !strings.Contains(out, "Quit Zed") || strings.Contains(out, "Save All") {
		t.Fatalf("Zed category should show only Zed rows:\n%s", out)
	}
	if strings.Contains(out, "CATEGORY") {
		t.Error("single-category view should not show the category column")
	}

	s.CycleCategory(-1) // back to All
	s.CycleCategory(-1) // wraps to Go
	out = screenText(s, 100, 30)
	if !strings.Contains(out, "Find All References") || strings.Contains(out, "Quit Zed") {
		t.Fatalf("expected wrap to Go:\n%s", out)
	}
}

func TestSearchFiltersAndClears(t *testing.T) {
	s, in := newTestScreen(t)

	in.SetValue("terminal")
	out := screenText(s, 100, 30)
	if !strings.Contains(out, "Terminal Panel") || !strings.Contains(out, "New Terminal Panel") {
		t.Fatalf("search results missing:\n%s", out)
	}
	if strings.Contains(out, "Open settings") {
		t.Error("non-matching rows should be hidden")
	}

	in.SetValue("zzzzzz")
	if out = screenText(s, 100, 30); !strings.Contains(out, "No shortcuts match") {
		t.Fatalf("empty state missing:\n%s", out)
	}

	s.ClearSearch()
	if in.Value() != "" {
		t.Fatal("ClearSearch left text in the field")
	}
	if out = screenText(s, 100, 30); !strings.Contains(out, "Open settings") {
		t.Fatal("clearing search should restore the list")
	}
}

func TestTabClearsActiveSearch(t *testing.T) {
	s, in := newTestScreen(t)
	in.SetValue("save")
	s.CycleCategory(+1)
	if in.Value() != "" {
		t.Fatal("changing category should clear the search")
	}
}

func TestSelectionScrollsIntoView(t *testing.T) {
	s, _ := newTestScreen(t)
	// Short terminal: only a few rows fit, so moving down must scroll.
	const w, h = 100, 16
	for i := 0; i < 20; i++ {
		s.Move(+1)
	}
	out := screenText(s, w, h)
	if strings.Contains(out, "Open settings") {
		t.Fatalf("first row should have scrolled off:\n%s", out)
	}
	if !strings.Contains(out, "21/56") {
		t.Fatalf("position read-out should say 21/56:\n%s", out)
	}

	// Move clamps at both ends.
	for i := 0; i < 200; i++ {
		s.Move(+1)
	}
	if out = screenText(s, w, h); !strings.Contains(out, "56/56") {
		t.Fatalf("should clamp at last row:\n%s", out)
	}
	for i := 0; i < 200; i++ {
		s.Move(-1)
	}
	if out = screenText(s, w, h); !strings.Contains(out, "1/56") {
		t.Fatalf("should clamp at first row:\n%s", out)
	}
}

func TestTwoStepChordRendering(t *testing.T) {
	s, _ := newTestScreen(t)
	s.CycleCategory(+1) // Zed
	s.CycleCategory(+1) // File
	s.CycleCategory(+1) // Edit
	s.CycleCategory(+1) // Selection
	s.CycleCategory(+1) // View
	out := screenText(s, 110, 30)
	if !strings.Contains(out, "then") || !strings.Contains(out, "↑") {
		t.Fatalf("two-step chords should show 'then' and arrow glyphs:\n%s", out)
	}
}

func TestTooSmallTerminal(t *testing.T) {
	s, _ := newTestScreen(t)
	if out := screenText(s, 40, 8); !strings.Contains(out, "Terminal too small") {
		t.Fatalf("expected size warning:\n%s", out)
	}
}

func TestNarrowTerminalDoesNotPanic(t *testing.T) {
	s, in := newTestScreen(t)
	for _, size := range [][2]int{{64, 12}, {65, 13}, {70, 40}, {200, 60}} {
		_ = renderLines(s, size[0], size[1])
		in.SetValue("ctrl")
		_ = renderLines(s, size[0], size[1])
		in.SetValue("")
		s.CycleCategory(+1)
	}
}

func TestNarrowTerminalDropsCategoryColumnAndElides(t *testing.T) {
	s, _ := newTestScreen(t)

	wide := screenText(s, 100, 24)
	if !strings.Contains(wide, "CATEGORY") {
		t.Fatal("wide terminal should show the category column")
	}

	narrow := screenText(s, 70, 24)
	if strings.Contains(narrow, "CATEGORY") {
		t.Fatalf("narrow terminal should drop the category column:\n%s", narrow)
	}
	if !strings.Contains(narrow, "ACTION") {
		t.Fatalf("action header should be intact:\n%s", narrow)
	}
	if !strings.Contains(narrow, "…") {
		t.Fatalf("clipped action text should end in an ellipsis:\n%s", narrow)
	}
}

func TestFit(t *testing.T) {
	for _, c := range []struct {
		in   string
		w    int
		want string
	}{
		{"Save", 10, "Save"}, {"Open Extensions store", 10, "Open Exte…"},
		{"abc", 3, "abc"}, {"abcd", 1, "…"}, {"abcd", 0, "abcd"},
	} {
		if got := fit(c.in, c.w); got != c.want {
			t.Errorf("fit(%q,%d) = %q, want %q", c.in, c.w, got, c.want)
		}
	}
}

type fakeNav struct{ nexts int }

func (f *fakeNav) Next()                        { f.nexts++ }
func (f *fakeNav) Prev()                        {}
func (f *fakeNav) Enter() bool                  { return false }
func (f *fakeNav) Exit()                        {}
func (f *fakeNav) Current() framework.Focusable { return nil }

func press(s *Screen, r rune) {
	_, _ = s.HandleInput(framework.Event{Key: framework.KeyRune, Rune: r})
}
func pressKey(s *Screen, k framework.Key) {
	_, _ = s.HandleInput(framework.Event{Key: k})
}

func TestBrowseKeys(t *testing.T) {
	s, _ := newTestScreen(t)
	nav := &fakeNav{}
	s.SetContext(framework.AppContext{Focus: nav})
	const w, h = 100, 20
	renderLines(s, w, h) // sets the page size

	press(s, 'j')
	press(s, 'j')
	if out := screenText(s, w, h); !strings.Contains(out, "3/56") {
		t.Fatalf("j j should select row 3:\n%s", out)
	}
	press(s, 'k')
	if out := screenText(s, w, h); !strings.Contains(out, "2/56") {
		t.Fatal("k should move up")
	}
	press(s, 'G')
	if out := screenText(s, w, h); !strings.Contains(out, "56/56") {
		t.Fatal("G should jump to the last row")
	}
	press(s, 'g')
	if out := screenText(s, w, h); !strings.Contains(out, "1/56") {
		t.Fatal("g should jump to the first row")
	}
	press(s, 'd')
	if out := screenText(s, w, h); strings.Contains(out, " 1/56") {
		t.Fatal("d should page down")
	}
	pressKey(s, framework.KeyEnd)
	if out := screenText(s, w, h); !strings.Contains(out, "56/56") {
		t.Fatal("End should jump to the last row")
	}
	pressKey(s, framework.KeyHome)

	// categories: 1 = All, 2 = Zed ... 7 = Go
	press(s, '7')
	if out := screenText(s, w, h); !strings.Contains(out, "Find All References") || strings.Contains(out, "Quit Zed") {
		t.Fatalf("7 should open Go:\n%s", out)
	}
	press(s, '2')
	if out := screenText(s, w, h); !strings.Contains(out, "Quit Zed") {
		t.Fatal("2 should open Zed")
	}
	press(s, '9') // out of range: ignored, still Zed
	if out := screenText(s, w, h); !strings.Contains(out, "Quit Zed") {
		t.Fatal("out-of-range number should be ignored")
	}
	press(s, 'l')
	if out := screenText(s, w, h); !strings.Contains(out, "Save All") {
		t.Fatal("l should go to File")
	}
	pressKey(s, framework.KeyLeft)
	pressKey(s, framework.KeyLeft)
	if out := screenText(s, w, h); !strings.Contains(out, "All shortcuts") {
		t.Fatalf("← ← from File should reach All:\n%s", out)
	}

	// '/' hands focus to the search field via the navigator.
	press(s, '/')
	if nav.nexts != 1 {
		t.Fatalf("/ should call Nav().Next() once, got %d", nav.nexts)
	}
}

func TestHandleRuneReportsUnknown(t *testing.T) {
	s, _ := newTestScreen(t)
	if s.handleRune('z') || s.handleRune('0') {
		t.Fatal("unmapped runes must report false")
	}
	if !s.handleRune('j') {
		t.Fatal("mapped rune must report true")
	}
}
