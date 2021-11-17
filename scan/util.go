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

func unescaped(raw []rune, ptr int, substr string) bool {
	length := len([]rune(substr))
	if len(raw[ptr:]) < length {
		return false
	}
	if string(raw[ptr:ptr+length]) != substr {
		return false
	}
	if ptr > 0 && unescaped(raw, ptr-1, "\\") {
		return false
	}
	return true
}

func precededBy(raw []rune, ptr int, ss []string) bool {
	for _, substr := range ss {
		if ptr >= len(substr) && string(raw[ptr-len([]rune(substr)):ptr]) == substr {
			return true
		}
	}
	return false
}

func followedBy(raw []rune, ptr int, ss []string) bool {
	for _, substr := range ss {
		if len(raw[ptr+1:]) >= len(substr) && string(raw[ptr+1:ptr+1+len([]rune(substr))]) == substr {
			return true
		}
	}
	return false
}
