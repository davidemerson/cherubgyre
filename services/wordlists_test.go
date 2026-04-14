package services

import (
	"strings"
	"testing"
	"unicode"
)

func TestWordlistsNonEmpty(t *testing.T) {
	if len(angelWords) == 0 {
		t.Fatal("angelWords is empty — go:embed failed or file is empty")
	}
	if len(typeWords) == 0 {
		t.Fatal("typeWords is empty")
	}
	if len(cityWords) == 0 {
		t.Fatal("cityWords is empty")
	}
}

func TestWordlistsHaveCleanEntries(t *testing.T) {
	check := func(name string, words []string) {
		for i, w := range words {
			if w == "" {
				t.Errorf("%s[%d] is empty", name, i)
			}
			if strings.TrimSpace(w) != w {
				t.Errorf("%s[%d] = %q has surrounding whitespace", name, i, w)
			}
			for _, r := range w {
				if !unicode.IsLower(r) && r != '-' {
					t.Errorf("%s[%d] = %q contains non-lowercase/non-hyphen char %q", name, i, w, r)
					break
				}
			}
		}
	}
	check("angelWords", angelWords)
	check("typeWords", typeWords)
	check("cityWords", cityWords)
}

func TestWordlistsHaveNoDuplicates(t *testing.T) {
	check := func(name string, words []string) {
		seen := make(map[string]struct{}, len(words))
		for _, w := range words {
			if _, dup := seen[w]; dup {
				t.Errorf("%s has duplicate entry %q", name, w)
			}
			seen[w] = struct{}{}
		}
	}
	check("angelWords", angelWords)
	check("typeWords", typeWords)
	check("cityWords", cityWords)
}

func TestUsernameCombinationsAboveFloor(t *testing.T) {
	// Floor chosen to prevent a regression where one of the wordlists
	// is accidentally truncated. 1M is still well under what's
	// needed for strong anonymity but the login rate limiter is
	// the primary defense — this floor is about catching dumb edits,
	// not anonymity guarantees.
	const minCombinations = 1_000_000
	got := UsernameCombinations()
	if got < minCombinations {
		t.Errorf("UsernameCombinations() = %d, want >= %d", got, minCombinations)
	}
}
