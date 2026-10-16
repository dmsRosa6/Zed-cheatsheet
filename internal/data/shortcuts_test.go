package data

import "testing"

func TestTotalMatchesSource(t *testing.T) {
	// 3 + 8 + 7 + 9 + 19 + 10 rows on https://cheatsheets.zip/zed
	if got := Total(All()); got != 56 {
		t.Fatalf("Total = %d, want 56", got)
	}
}

func TestSearch(t *testing.T) {
	cats := All()

	if got := Search(cats, "  "); got != nil {
		t.Fatalf("blank query should match nothing, got %d", len(got))
	}

	// Terms match independently, so "Focus Split Buffer Up" qualifies too.
	got := Search(cats, "split up")
	if len(got) != 2 || got[0].Action != "Split Up" {
		t.Fatalf("'split up' = %+v", got)
	}

	found := false
	for _, m := range Search(cats, "ctrl shift p") {
		if m.Action == "Command Palette" {
			found = true
		}
	}
	if !found {
		t.Fatal("searching a chord did not find Command Palette")
	}

	if got := Search(cats, "GO TO"); len(got) < 5 {
		t.Fatalf("case-insensitive 'go to' matched only %d", len(got))
	}

	if got := Search(cats, "selection"); len(got) < 9 {
		t.Fatalf("category name should be searchable, got %d", len(got))
	}
}
