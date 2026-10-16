// Package ui builds the terminal interface with Glyph.
package ui

import (
	"errors"
	"fmt"

	"github.com/dmsRosa6/glyph/app"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/term"

	"github.com/example/zed-cheatsheet/internal/data"
)

const (
	searchX     = 9
	searchWidth = 34
)

// Run builds the UI and blocks until the user quits (Ctrl+C).
func Run(cats []data.Category) error {
	// Glyph's NewApp swallows the "no terminal" error from its canvas and hands
	// back an app whose Canvas is nil, so check up front and fail politely.
	if _, err := term.TermSize(); err != nil {
		return errors.New("stdout is not a terminal; use -list for plain text")
	}

	a, err := app.NewApp(app.AppConfig{
		Bg:             Background,
		Fg:             Foreground,
		RenderMode:     render.OnDemandMode(), // only redraw when something changes
		FullTerminal:   true,                  // track the terminal's size
		DisableFileLog: true,                  // don't litter the cwd with logs/
	})
	if err != nil {
		return fmt.Errorf("start UI: %w", err)
	}
	if a.Canvas == nil {
		return errors.New("start UI: Glyph could not create a canvas")
	}

	// Search field: the one focusable widget, so it always receives typing.
	search, err := primitive.NewTextInput(&geom.Point{X: searchX, Y: 1}, searchWidth, primitive.TextInputConfig{
		Style: framework.Style{Bg: colInputBg, Fg: colText},
		Layer: 1,
	})
	if err != nil {
		return err
	}

	screen, err := NewScreen(cats, search)
	if err != nil {
		return err
	}

	title, err := primitive.NewStaticText(&geom.Point{X: 2, Y: 0}, primitive.StaticTextConfig{
		Value: "Zed Cheat Sheet", Fg: colAccent, Layer: 1,
	})
	if err != nil {
		return err
	}

	// Anchored to the right edge: Glyph recomputes its X on every resize.
	note, err := primitive.NewStaticText(&geom.Point{}, primitive.StaticTextConfig{
		Value:  fmt.Sprintf("Linux / Windows  ·  %d shortcuts ", data.Total(cats)),
		Fg:     colDim,
		Layer:  1,
		Anchor: framework.Anchor{V: framework.Start, H: framework.End},
	})
	if err != nil {
		return err
	}

	// Anchored to the bottom edge.
	footer, err := primitive.NewStaticText(&geom.Point{X: 2, Y: 0}, primitive.StaticTextConfig{
		Value:  "j/k ↑↓ select   h/l ←→ 1-7 category   / search   Tab focus   Esc clear   Ctrl+C quit",
		Fg:     colDim,
		Layer:  1,
		Anchor: framework.Anchor{V: framework.End},
	})
	if err != nil {
		return err
	}

	for _, w := range []framework.Drawable{screen, search, title, note, footer} {
		a.Canvas.AddShape(w)
	}

	// Focus order is [screen, search]; the screen starts focused (browse mode).
	// Global bindings: Tab/Enter/Esc are "structural" keys and run global-first;
	// Up/Down aren't claimed by the search field, so they fall through to here
	// from either focus.
	act := func(fn func(framework.AppContext) bool) app.AppActionFunc {
		return func(ctx framework.AppContext, _ framework.Event) (bool, error) {
			return fn(ctx), nil
		}
	}
	searching := func(ctx framework.AppContext) bool { return ctx.Nav().Current() == search }

	a.BindKey(framework.KeyTab, act(func(ctx framework.AppContext) bool { ctx.Nav().Next(); return true }))
	a.BindKeyMod(framework.KeyTab, framework.ModShift, act(func(ctx framework.AppContext) bool { ctx.Nav().Prev(); return true }))
	a.BindKey(framework.KeyEnter, act(func(ctx framework.AppContext) bool {
		if searching(ctx) {
			ctx.Nav().Next() // done typing: browse the results
			return true
		}
		return false
	}))
	a.BindKey(framework.KeyEsc, act(func(ctx framework.AppContext) bool {
		screen.ClearSearch()
		if searching(ctx) {
			ctx.Nav().Next()
		}
		return true
	}))
	a.BindKey(framework.KeyDown, act(func(framework.AppContext) bool { screen.Move(+1); return true }))
	a.BindKey(framework.KeyUp, act(func(framework.AppContext) bool { screen.Move(-1); return true }))

	a.Run()
	return nil
}
