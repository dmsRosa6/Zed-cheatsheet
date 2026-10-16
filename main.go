// Command zed-cheatsheet is a terminal cheat sheet for Zed's keyboard
// shortcuts, built with the Glyph TUI framework.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/example/zed-cheatsheet/internal/data"
	"github.com/example/zed-cheatsheet/internal/ui"
)

func main() {
	list := flag.Bool("list", false, "print the shortcuts as plain text instead of opening the TUI")
	query := flag.String("search", "", "with -list: only show shortcuts matching these words")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-list [-search WORDS]]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	cats := data.All()

	if *list || *query != "" {
		printPlain(cats, *query)
		return
	}

	if err := ui.Run(cats); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// printPlain writes the cheat sheet as aligned text, grouped by category.
func printPlain(cats []data.Category, query string) {
	width := 0
	for _, c := range cats {
		for _, s := range c.Shortcuts {
			width = max(width, len(s.Chord()))
		}
	}

	if strings.TrimSpace(query) != "" {
		matches := data.Search(cats, query)
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "no shortcuts match %q\n", query)
			os.Exit(1)
		}
		for _, m := range matches {
			fmt.Printf("%-*s  %-28s %s\n", width, m.Chord(), m.Action, m.Category)
		}
		return
	}

	for i, c := range cats {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s (%d)\n", c.Name, len(c.Shortcuts))
		for _, s := range c.Shortcuts {
			fmt.Printf("  %-*s  %s\n", width, s.Chord(), s.Action)
		}
	}
}
