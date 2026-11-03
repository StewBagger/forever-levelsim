package levelsim

import (
	"sort"
	"strings"
)

// TrimTalents spends a full talent string's points up to a level's budget, the way a character
// levelling into that build spends them: its deepest tree first, then the next, each tree in the
// string's own order, which is row by row from the top. A tree's row is reached only once the rows
// above it are spent in full, so the trimmed string is as legal as the one it came from.
func TrimTalents(talents string, points int32) string {
	trees := strings.Split(talents, "-")
	spent := make([]int32, len(trees))
	for i, tree := range trees {
		for _, c := range tree {
			spent[i] += c - '0'
		}
	}
	order := make([]int, len(trees))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return spent[order[a]] > spent[order[b]] })

	out := make([][]byte, len(trees))
	for _, i := range order {
		for _, c := range trees[i] {
			take := min(c-'0', points)
			points -= take
			out[i] = append(out[i], byte('0'+take))
		}
	}
	parts := make([]string, len(out))
	for i, tree := range out {
		parts[i] = strings.TrimRight(string(tree), "0")
	}
	return strings.TrimRight(strings.Join(parts, "-"), "-")
}
