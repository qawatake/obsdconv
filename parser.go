package main

import (
	"strings"
	"unicode"
)

func consumeInlineMath(line []rune) (advance int) {
	if !(line[0] == RuneDollar && 1 < len(line) && !unicode.IsSpace(line[1])) {
		return 0
	}

	cur := 1
	for cur < len(line) && line[cur] != RuneDollar {
		cur++
	}
	if cur == len(line) {
		return 0
	}
	return cur + 1
}

func consumeEscaped(line []rune) (advance int, escaped []rune) {
	if line[0] == '\\' && 1 < len(line) {
		switch line[1] {
		case '#':
			advance = 2
			escaped = []rune("#")
			return advance, escaped
		}
	}
	return 0, nil
}

func consumeRepeat(line []rune, substr string) (advance int) {
	cur := 0
	length := len([]rune(substr))
	for len(line[cur:]) >= length && string(line[cur:cur+length]) == substr {
		cur += length
	}
	return cur
}

func consumeTag(line []rune) (advance int, tag string) {
	if !(line[0] == '#' && 1 < len(line) && unicode.IsGraphic(line[1]) && !unicode.IsSpace(line[1])) {
		return 0, ""
	}

	if !(unicode.IsLetter(line[1]) || unicode.IsNumber(line[1]) || strings.ContainsRune("-_/", line[1])) {
		return 0, ""
	}

	cur := 1
	for cur < len(line) && (unicode.IsLetter(line[cur]) || unicode.IsNumber(line[cur]) || strings.ContainsRune("-_/", line[cur])) {
		cur++
	}
	return cur, string(line[1:cur])
}
