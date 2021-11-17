package scan

import (
	"strings"
)

func indexInRunes(rns []rune, substr string) int {
	pos := strings.Index(string(rns), substr)
	if pos < 0 {
		return -1
	}
	return len([]rune(string(rns)[:pos]))
}