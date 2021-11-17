package scan

import (
	"strings"
)

func runeIndex(s string, substr string) int {
	pos := strings.Index(s, substr)
	if pos < 0 {
		return -1
	}
	return len([]rune(s[:pos]))
}
