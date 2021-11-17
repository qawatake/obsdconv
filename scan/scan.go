package scan

import (
	"strings"
	"unicode"
)

func ScanInlineCode(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "`") && len(raw[ptr:]) > 1) {
		return 0
	}

	adv := indexInRunes(raw[ptr+1:], "`")
	if adv < 0 {
		return 0
	}
	cur := ptr + 1 + adv
	if precededBy(raw, cur, []string{"\n\n", "\r\n\r\n"}) {
		return 0
	} else {
		return cur - ptr + 1
	}
}

func ScanInlineMath(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "$") && !followedBy(raw, ptr, []string{" ", "\t"})) {
		return 0
	}

	cur := ptr + 1
	for cur < len(raw)-1 && !unescaped(raw, cur, "$") {
		adv := indexInRunes(raw[cur+1:], "$")
		if adv < 0 {
			return 0
		}
		cur += 1 + adv
	}
	if cur >= len(raw) {
		return 0
	}
	if precededBy(raw, cur, []string{"\n\n", "\r\n\r\n", " ", "\t"}) {
		return 0
	}
	if cur == ptr+1 {
		return 0
	}
	if unescaped(raw, cur, "$") {
		return cur - ptr + 1
	}
	return 0
}

func ScanRepeat(raw []rune, ptr int, substr string) (advance int) {
	length := len([]rune(substr))
	cur := ptr
	next := cur + length
	for len(raw[cur:]) >= length && string(raw[cur:next]) == substr {
		cur = next
		next += length
	}
	return cur - ptr
}

func ScanTag(raw []rune, ptr int) (advance int, tag string) {
	if !(unescaped(raw, ptr, "#") && len(raw[ptr:]) > 1) {
		return 0, ""
	}

	if !(unicode.IsLetter(raw[ptr+1]) || unicode.IsNumber(raw[ptr+1]) || raw[ptr+1] == '_') {
		return 0, ""
	}

	cur := ptr + 1
	for cur < len(raw) && (unicode.IsLetter(raw[cur]) || unicode.IsNumber(raw[cur]) || strings.ContainsRune("-_/", raw[cur])) {
		cur++
	}

	return cur - ptr, string(raw[ptr+1 : cur])
}

func ScanInternalLink(raw []rune, ptr int) (advance int, content string) {
	if !(unescaped(raw, ptr, "[[") && len(raw[ptr:]) >= 5) {
		return 0, ""
	}

	cur := ptr + 2
	adv := indexInRunes(raw[cur:], "]]")
	if adv <= 0 {
		return 0, ""
	}

	content = strings.Trim(string(raw[cur:cur+adv]), " \t")
	if strings.ContainsAny(content, "\r\n") {
		return 0, ""
	}
	cur += adv + 2
	advance = cur - ptr
	return advance, content
}

func ScanEmbeds(raw []rune, ptr int) (advance int, content string) {
	if !unescaped(raw, ptr, "![[") {
		return 0, ""
	}
	cur := ptr + 1
	advance, content = ScanInternalLink(raw, cur)
	if advance == 0 {
		return 0, ""
	}
	cur += advance
	return cur - ptr, content
}

func validURI(uri string) bool {
	return !strings.ContainsAny(uri, " \t\r\n")
}

func scanExternalLinkHead(raw []rune, ptr int) (advance int, displayName string) {
	if !(unescaped(raw, ptr, "[") && len(raw[ptr:]) >= 2) {
		return 0, ""
	}
	cur := ptr + 1 // "[" の次
	for {
		adv := indexInRunes(raw[cur:], "]")
		if adv < 0 {
			return 0, ""
		}
		cur += adv // "["
		if unescaped(raw, cur, "]") {
			advance = cur + 1 - ptr
			content := string(raw[ptr+1 : cur])
			if !validExternalLinkHeadContent(content) {
				return 0, ""
			}
			displayName = strings.Trim(content, " \t\r\n")
			return advance, displayName
		}
		cur++ // "]" の次
	}
}

func validExternalLinkHeadContent(content string) bool {
	return !strings.Contains(content, "\r\n\r\n") && !strings.Contains(content, "\n\n")
}

func scanExternalLinkTail(raw []rune, ptr int) (advance int, ref string) {
	if !(unescaped(raw, ptr, "(") && len(raw[ptr:]) >= 2) {
		return 0, ""
	}
	cur := ptr + 1 // "(" の次
	for {
		adv := indexInRunes(raw[cur:], ")")
		if adv < 0 {
			return 0, ""
		}
		cur += adv // ")"
		if unescaped(raw, cur, ")") {
			advance = cur + 1 - ptr
			content := string(raw[ptr+1 : cur])
			if !validExternalLinkTailContent(content) {
				return 0, ""
			}
			uri := strings.Trim(string(content), " \t\r\n")
			if !validURI(uri) {
				return 0, ""
			}
			ref = uri
			return advance, ref
		}
		cur++ // ")" の次
	}

}

func validExternalLinkTailContent(content string) bool {
	return !strings.Contains(content, "\r\n\r\n") && !strings.Contains(content, "\n\n")
}

func ScanExternalLink(raw []rune, ptr int) (advance int, displayName string, ref string) {
	cur := ptr
	adv, displayName := scanExternalLinkHead(raw, cur)
	if adv == 0 {
		return 0, "", ""
	}
	cur += adv
	adv, ref = scanExternalLinkTail(raw, cur)
	if adv == 0 {
		return 0, "", ""
	}
	cur += adv
	return cur - ptr, displayName, ref
}

func ScanComment(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "%%") && len(raw) >= 2) {
		return 0
	}
	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "%")))
	cur := ptr + length // opening の %% の直後
	adv := indexInRunes(raw[cur:], strings.Repeat("%", length))
	if adv < 0 {
		return len(raw) - ptr
	}

	cur += adv + length // closing %% の直後
	return cur - ptr
}

func validMathBlockClosing(raw []rune, openPtr int, closingPtr int) bool {
	if !unescaped(raw, closingPtr, "$$") {
		return false
	}

	// 後ろに何もなければ OK
	if strings.Trim(string(raw[closingPtr+2:]), " \r\n") == "" {
		return true
	}

	// inline だったら OK
	if !strings.ContainsRune(string(raw[openPtr:closingPtr]), '\n') {
		return true
	}

	posLineFeed := indexInRunes(raw[closingPtr+2:], "\n")
	if posLineFeed < 0 {
		return false
	}

	remaining := string(raw[closingPtr+2 : closingPtr+2+posLineFeed])
	remaining = strings.Trim(remaining, " \r\n")
	return remaining == ""
}

func ScanMathBlock(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "$$") && len(raw) >= 3) {
		return 0
	}

	cur := ptr + 2
	for {
		adv := indexInRunes(raw[cur:], "$$")
		if adv < 0 {
			if raw[len(raw)-1] == '\n' {
				return len(raw) - ptr
			} else {
				return 0
			}
		}

		cur += adv
		if validMathBlockClosing(raw, ptr, cur) {
			return cur + 2 - ptr
		}
		cur += 2
	}
}

func ScanCodeBlock(raw []rune, ptr int) (advance int) {
	if !unescaped(raw, ptr, "```") {
		return 0
	}
	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "`")))
	cur := ptr + length
	adv := indexInRunes(raw[cur:], "```")
	if adv < 0 {
		return len(raw) - ptr
	}
	cur += adv // closing の "```" の最初の "`"
	closingLength := len(raw[cur:]) - len([]rune(strings.TrimLeft(string(raw[cur:]), "`")))

	// inline の場合は opening bracket と closing bracket は同じ長さでなければならない
	if !strings.ContainsRune(string(raw[ptr:cur+closingLength]), '\n') {
		if length != closingLength {
			return 0
		} else {
			return cur + closingLength - ptr
		}
	} else { // 複数行の場合は, closing bracket の長さは opening bracket の長さ以上でなければいけない
		adv := indexInRunes(raw[cur:], strings.Repeat("`", length))
		if adv < 0 {
			return len(raw) - ptr
		}
		cur += adv
		closingLength := len(raw[cur:]) - len([]rune(strings.TrimLeft(string(raw[cur:]), "`")))
		return cur + closingLength - ptr
	}
}

func ScanHeader(raw []rune, ptr int) (advance int, level int, headertext string) {
	if !(unescaped(raw, ptr, "#")) {
		return
	}
	// 前の改行まではスペースしか入っちゃいけない
	for back := ptr; back > 0; {
		back--
		if raw[back] == '\n' {
			break
		} else if raw[back] == ' ' {
			continue
		} else {
			return 0, 0, ""
		}
	}

	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "#")))
	cur := ptr + length // "#" の直後
	if cur >= len(raw) || !(raw[cur] == ' ' || raw[cur] == '\n' || (len(raw) >= 2 && string(raw[cur:cur+2]) == "\r\n")) {
		return 0, 0, ""
	}

	adv := indexInRunes(raw[cur:], "\n")
	if adv < 0 {
		advance = len(raw[ptr:])
		level = length
		headertext = strings.Trim(string(raw[cur:]), " \t\r\n")
		return advance, length, headertext
	}

	advance = cur + adv + 1 - ptr // \r\n や \n を含む
	level = length
	headertext = strings.Trim(string(raw[cur:ptr+advance]), " \t\r\n")
	return advance, level, headertext
}

func ScanEscaped(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "\\") && len(raw[ptr:]) > 1) {
		return 0
	}
	return 2
}

func ScanNormalComment(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "<!--")) {
		return 0
	}
	adv := indexInRunes(raw[ptr:], "-->")
	if adv < 0 {
		return len(raw) - ptr
	}
	cur := ptr + adv + 3
	return cur - ptr
}
