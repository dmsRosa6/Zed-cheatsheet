# Zed-cheatsheet

A terminal cheat sheet for [Zed](https://zed.dev)'s keyboard shortcuts, built
with the [Glyph](https://github.com/dmsRosa6/Glyph) TUI framework for Go.

The content is the [cheatsheets.zip Zed page](https://cheatsheets.zip/zed)
(Linux / Windows bindings), reorganised into an interactive, searchable view.

```
  Zed Cheat Sheet                                                  Linux / Windows  ·  56 shortcuts
  Search split                               / edit search   Esc clear

 ╭─ Categories ─────╮ ╭─ Search “split” ──────────────────────────────────────────────────────────╮
 │  All           8 │ │ SHORTCUT                    ACTION                             CATEGORY   │
 │  Zed           0 │ │───────────────────────────────────────────────────────────────────────────│
 │  File          0 │ │▌ Ctrl + K  then  ↑          Split Up                           View       │
 │  Edit          0 │ │  Ctrl + K  then  ↓          Split Down                         View       │
 │  Selection     0 │ │  Ctrl + K  then  Ctrl + ↑   Focus Split Buffer Up              View       │
 │  View          8 │ │  …                                                                        │
 ╰──────────────────╯ ╰───────────────────────────────────────────────────────── 1/8 ─╯
  j/k ↑↓ select   h/l ←→ 1-7 category   / search   Tab focus   Esc clear   Ctrl+C quit
```

## ⚠️ Disclaimer

This is for personal use and mostly vibe coded using my own framework.

## Run it

Requires Go 1.24+.

```sh
go run .            # open the TUI
go run . -list      # print everything as plain text
go run . -search "split up"   # plain-text search (no TUI)
```

or `make run` / `make build` (binary lands in `bin/`).

## Keys

Two things can have keyboard focus, and `Tab` / `Shift+Tab` switches between
them (Glyph's `FocusManager`): the **table** (browse mode, the default) and the
**search field**.

| Browse mode (table focused)   | Action                                      |
| ----------------------------- | ------------------------------------------- |
| `j` / `k` or `↓` / `↑`        | Move the selection (the table scrolls)      |
| `d` / `u`                     | Half a page down / up                       |
| `g` / `G` or `Home` / `End`   | First / last row                            |
| `h` / `l`, `[` / `]`, `←` / `→` | Previous / next category                  |
| `1`–`7`                       | Jump to All, Zed, File, Edit, Selection, View, Go |
| `/`                           | Focus the search field                      |

| Search field focused | Action                                              |
| -------------------- | --------------------------------------------------- |
| *type*               | Filter live across every category                   |
| `Enter`              | Back to the table to browse the results             |
| `Esc`                | Clear the search and go back to the table           |
| `↑` / `↓`            | Still move the selection                            |

Anywhere: `Tab` / `Shift+Tab` switch focus, `Esc` clears the search, `Ctrl+C` quits.

Search matches every word you type against the action, the keys and the
category, so `shift` finds all Shift chords, `go to` finds the navigation
commands, and `terminal` finds both terminal panels. While searching, the
sidebar shows how many matches each category has. Changing category clears the search.

Two-step chords are drawn the way Zed treats them: `Ctrl+K then ↑` means press
`Ctrl+K`, release, then press `↑` (not three keys held at once).

## Layout

```
main.go                     flags (-list, -search) and entry point
internal/data/shortcuts.go  the cheat sheet itself + search
internal/ui/app.go          builds the Glyph app, widgets and global key bindings
internal/ui/screen.go       custom focusable widget: sidebar, table, browse-mode keys
internal/ui/pen.go          small drawing helpers (frames, text, key caps)
internal/ui/theme.go        colours
```

### How it uses Glyph

- `app.App` with `FullTerminal` and `render.OnDemandMode()`: the screen redraws
  only when a key changes something.
- Focus: the table (`Screen`, which embeds `mixin.FocusableNode`) and the
  `primitive.TextInput` search field are the two focusable widgets. Glyph
  dispatches non-structural keys to the focused widget first, so browse keys
  are `BindAction`s on `Screen` and typing goes to the field. `Tab`, `Enter`
  and `Esc` are structural keys, which Glyph hands to the global
  `App.BindKey` bindings first; those call `ctx.Nav().Next()` / `Prev()` to
  move focus. `↑`/`↓` are global too, so they work from either focus.
- `primitive.StaticText` for the title, labels and footer. The right-hand note
  and the footer use `framework.Anchor` (right / bottom), so Glyph re-positions
  them on every resize.
- `Screen` is a custom drawable built on `mixin.Node`, the pattern
  Glyph's `devguide.md` describes for custom widgets. It sizes itself from the
  buffer every frame, borrows `primitive.Rounded` for its frame runes, and
  guards its state with a mutex because input and rendering run on different
  goroutines.

## Updating the cheat sheet

Everything lives in `internal/data/shortcuts.go`. Add a row with
`sc("Action name", "Ctrl", "Shift", "P")`; use `Then` between the two steps of
a chord (`sc("Split Up", "Ctrl", "K", Then, "Up")`). `go test ./...` checks the
row count (56 on the source page) and the search behaviour.

## Notes

- Bindings are the **Linux / Windows** ones from the source page. On macOS
  Zed mostly uses `Cmd` where this sheet says `Ctrl`.
- The source page's subtitle says "50 keyboard shortcuts" but its tables
  contain 56 rows; the app counts the rows.
- Glyph is pinned to its `v0.1.1` release in `go.mod`.
- The module path `github.com/example/zed-cheatsheet` is a placeholder:
  find-and-replace it in `go.mod` and the Go imports with your own.
- Needs a real terminal at least 64×12. In a pipe or CI use `-list`.
