package services

import (
	_ "embed"
	"strings"
)

// Wordlists for the three-slug username format. Each file is a newline-
// separated list of lowercase ASCII words. The spec (nnix.com/projects/
// cherubgyre/) calls for the shape `angel-type-vortex-city`; we name the
// files accordingly so future edits are obvious.

//go:embed wordlists/angels.txt
var angelsRaw string

//go:embed wordlists/types.txt
var typesRaw string

//go:embed wordlists/cities.txt
var citiesRaw string

var (
	angelWords = splitWordlist(angelsRaw)
	typeWords  = splitWordlist(typesRaw)
	cityWords  = splitWordlist(citiesRaw)
)

// UsernameCombinations returns the total number of distinct usernames the
// current wordlists can produce. Exposed for tests and ops tooling.
func UsernameCombinations() int {
	return len(angelWords) * len(typeWords) * len(cityWords)
}

func splitWordlist(raw string) []string {
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		w := strings.TrimSpace(line)
		if w == "" {
			continue
		}
		out = append(out, w)
	}
	return out
}
