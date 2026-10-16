// Package data holds the Zed keyboard shortcut reference.
//
// Source: https://cheatsheets.zip/zed (Linux / Windows bindings).
package data

import "strings"

// Then separates the steps of a two-step chord inside Shortcut.Keys. Zed's
// split bindings are "press Ctrl+K, then press Up", not three keys held
// together, so {"Ctrl","K",Then,"Up"} keeps that distinction.
const Then = "then"

// Shortcut is a key chord and what it does. Each element of Keys is one key
// pressed together with its neighbours, unless it is Then.
type Shortcut struct {
	Keys   []string
	Action string
}

// Chord renders the keys as text: "Ctrl+Shift+P" or "Ctrl+K then Up".
func (s Shortcut) Chord() string {
	var b strings.Builder
	for i, k := range s.Keys {
		switch {
		case k == Then:
			b.WriteString(" then ")
		case i > 0 && s.Keys[i-1] != Then:
			b.WriteString("+" + k)
		default:
			b.WriteString(k)
		}
	}
	return b.String()
}

// Category groups related shortcuts under a heading.
type Category struct {
	Name      string
	Shortcuts []Shortcut
}

func sc(action string, keys ...string) Shortcut { return Shortcut{Keys: keys, Action: action} }

// All returns every category in display order.
func All() []Category {
	return []Category{
		{Name: "Zed", Shortcuts: []Shortcut{
			sc("Open settings", "Ctrl", ","),
			sc("Open Extensions store", "Ctrl", "Shift", "X"),
			sc("Quit Zed", "Ctrl", "Q"),
		}},
		{Name: "File", Shortcuts: []Shortcut{
			sc("New file", "Ctrl", "N"),
			sc("New Window", "Ctrl", "Shift", "N"),
			sc("Open folder", "Ctrl", "O"),
			sc("Open remote folder", "Ctrl", "Alt", "Shift", "O"),
			sc("Save", "Ctrl", "S"),
			sc("Save As...", "Ctrl", "Shift", "S"),
			sc("Save All", "Ctrl", "Alt", "S"),
			sc("Close Window", "Ctrl", "Shift", "W"),
		}},
		{Name: "Edit", Shortcuts: []Shortcut{
			sc("Undo", "Ctrl", "Z"),
			sc("Redo", "Ctrl", "Shift", "Z"),
			sc("Cut", "Ctrl", "X"),
			sc("Copy", "Ctrl", "C"),
			sc("Paste", "Ctrl", "V"),
			sc("Find in project", "Ctrl", "Shift", "F"),
			sc("Toggle line comment", "Ctrl", "/"),
		}},
		{Name: "Selection", Shortcuts: []Shortcut{
			sc("Select all", "Ctrl", "A"),
			sc("Expand selection", "Alt", "Shift", "Right"),
			sc("Shrink selection", "Alt", "Shift", "Left"),
			sc("Add cursor above", "Alt", "Shift", "Up"),
			sc("Add cursor below", "Alt", "Shift", "Down"),
			sc("Select next occurrence", "Ctrl", "Shift", "Down"),
			sc("Move line up", "Alt", "Up"),
			sc("Move line down", "Alt", "Down"),
			sc("Duplicate selection", "Ctrl", "Shift", "D"),
		}},
		{Name: "View", Shortcuts: []Shortcut{
			sc("Toggle Left Dock", "Ctrl", "B"),
			sc("Toggle Right Dock", "Ctrl", "Alt", "B"),
			sc("Toggle Bottom Dock", "Ctrl", "J"),
			sc("Close All Docks", "Ctrl", "Alt", "Y"),
			sc("Split Up", "Ctrl", "K", Then, "Up"),
			sc("Split Down", "Ctrl", "K", Then, "Down"),
			sc("Split Left", "Ctrl", "K", Then, "Left"),
			sc("Split Right", "Ctrl", "K", Then, "Right"),
			sc("Focus Split Buffer Up", "Ctrl", "K", Then, "Ctrl", "Up"),
			sc("Focus Split Buffer Down", "Ctrl", "K", Then, "Ctrl", "Down"),
			sc("Focus Split Buffer Left", "Ctrl", "K", Then, "Ctrl", "Left"),
			sc("Focus Split Buffer Right", "Ctrl", "K", Then, "Ctrl", "Right"),
			sc("Project Panel", "Ctrl", "Shift", "E"),
			sc("Outline Panel", "Ctrl", "Shift", "B"),
			sc("Collab Panel", "Ctrl", "Shift", "C"),
			sc("Terminal Panel", "Ctrl", "`"),
			sc("New Terminal Panel", "Ctrl", "~"),
			sc("Diagnostics", "Ctrl", "Shift", "M"),
			sc("Switch Tabs", "Ctrl", "Tab"),
		}},
		{Name: "Go", Shortcuts: []Shortcut{
			sc("Back", "Ctrl", "Alt", "-"),
			sc("Command Palette", "Ctrl", "Shift", "P"),
			sc("Go to File", "Ctrl", "E"),
			sc("Go to Symbol in Editor", "Ctrl", "Shift", "O"),
			sc("Go to Line/Column", "Ctrl", "G"),
			sc("Go to Definition", "F12"),
			sc("Go to Type Definition", "Ctrl", "F12"),
			sc("Find All References", "Alt", "Shift", "F12"),
			sc("Next Problem", "F8"),
			sc("Previous Problem", "Shift", "F8"),
		}},
	}
}

// Match is a shortcut paired with the category it belongs to, used for
// search results that span categories.
type Match struct {
	Category string
	Shortcut
}

// Search returns every shortcut whose action, chord or category contains all
// whitespace-separated terms in query (case-insensitive). An empty query
// matches nothing; callers decide what to show in that case.
func Search(cats []Category, query string) []Match {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return nil
	}
	var out []Match
	for _, c := range cats {
		for _, s := range c.Shortcuts {
			hay := strings.ToLower(s.Action + " " + s.Chord() + " " + strings.Join(s.Keys, " ") + " " + c.Name)
			ok := true
			for _, t := range terms {
				if !strings.Contains(hay, t) {
					ok = false
					break
				}
			}
			if ok {
				out = append(out, Match{Category: c.Name, Shortcut: s})
			}
		}
	}
	return out
}

// Total counts shortcuts across all categories.
func Total(cats []Category) int {
	n := 0
	for _, c := range cats {
		n += len(c.Shortcuts)
	}
	return n
}
