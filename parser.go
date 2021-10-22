package main

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

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

func consumeInlineCode(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "`") && len(raw[ptr:]) > 1) {
		return 0
	}

	pos := strings.IndexRune(string(raw[ptr+1:]), '`')
	if pos < 0 {
		return 0
	}
	cur := ptr + 1 + len([]rune(string(raw[ptr+1:])[:pos]))
	if precededBy(raw, cur, []string{"\n\n", "\r\n\r\n"}) {
		return 0
	} else {
		return cur - ptr + 1
	}
}

func consumeInlineMath(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "$") && !followedBy(raw, ptr, []string{" ", "\t"})) {
		return 0
	}

	cur := ptr + 1
	for cur < len(raw)-1 && !unescaped(raw, cur, "$") {
		pos := strings.IndexRune(string(raw[cur+1:]), RuneDollar)
		if pos < 0 {
			return 0
		}
		adv := 1 + len([]rune(string(string(raw[cur:])[:pos])))
		cur += adv
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

func consumeRepeat(raw []rune, ptr int, substr string) (advance int) {
	length := len([]rune(substr))
	cur := ptr
	next := cur + length
	for len(raw[cur:]) >= length && string(raw[cur:next]) == substr {
		cur = next
		next += length
	}
	return cur - ptr
}

func consumeTag(raw []rune, ptr int) (advance int, tag string) {
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

func consumeInternalLink(raw []rune, ptr int) (advance int, content string) {
	if !(unescaped(raw, ptr, "[[") && len(raw[ptr:]) >= 5) {
		return 0, ""
	}

	pos := strings.Index(string(raw[ptr+2:]), "]]")
	if pos <= 0 {
		return 0, ""
	}
	content = strings.Trim(string(string(raw[ptr+2:])[:pos]), " \t")
	if strings.ContainsAny(content, "\r\n") {
		return 0, ""
	}
	advance = 2 + len([]rune(string(string(raw[ptr+2:])[:pos]))) + 2
	return advance, content
}

func validURI(uri string) bool {
	return !strings.ContainsAny(uri, " \t\r\n")
}

func consumeExternalLink(raw []rune, ptr int) (advance int, displayName string, ref string) {
	if !(unescaped(raw, ptr, "[") && len(raw[ptr:]) >= 5) {
		return 0, "", ""
	}
	cur := ptr + 1 // "[" の次
	midpos := strings.Index(string(raw[cur:]), "](")
	if midpos < 0 || midpos+1 >= len(string(raw[cur:]))-1 {
		return 0, "", ""
	}
	next := cur + len([]rune(string(string(raw[cur:])[:midpos]))) // "](" の "["
	displayName = strings.Trim(string(raw[cur:next]), " \t")
	cur = next
	if strings.Contains(displayName, "\r\n\r\n") || strings.Contains(displayName, "\n\n") {
		return 0, "", ""
	} else if !unescaped(raw, cur, "](") {
		return 0, "", ""
	}
	displayName = strings.Trim(displayName, "\r\n") // \r\n や \n 一つを削除
	next += 2                                       // "](" の次
	cur = next
	endpos := strings.IndexRune(string(raw[cur:]), ')')
	if endpos < 0 {
		return 0, "", ""
	}
	next = cur + len([]rune(string(string(raw[cur:])[:endpos]))) // ")"
	ref = strings.Trim(string(raw[cur:next]), " \t")
	cur = next
	if strings.Contains(ref, "\r\n\r\n") || strings.Contains(ref, "\n\n") {
		return 0, "", ""
	} else if !unescaped(raw, cur, ")") {
		return 0, "", ""
	}
	ref = strings.Trim(ref, "\r\n") // \r\n や \n 一つを削除
	if !validURI(ref) {
		return 0, "", ""
	}
	advance = next + 1 - ptr
	return
}

func consumeComment(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "%%") && len(raw) >= 2) {
		return 0
	}
	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "%")))
	cur := ptr + length
	pos := strings.Index(string(raw[cur:]), strings.Repeat("%", length))
	if pos < 0 {
		return len(raw[ptr:]) - ptr
	}
	return cur + len([]rune(string(string(raw[cur:])[:pos]))) + length
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

	posLineFeed := strings.IndexRune(string(raw[closingPtr+2:]), '\n')
	if posLineFeed < 0 {
		return false
	}

	remaining := string(string(raw[closingPtr+2:])[:posLineFeed])
	remaining = strings.Trim(remaining, " \r\n")
	return remaining == ""
}

func consumeMathBlock(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "$$") && len(raw) >= 3) {
		return 0
	}

	cur := ptr + 2
	for {
		pos := strings.Index(string(raw[cur:]), "$$")
		if pos < 0 {
			if raw[len(raw)-1] == '\n' {
				return len(raw) - ptr
			} else {
				return 0
			}
		}

		cur += len([]rune(string(string(raw[cur:])[:pos])))
		if validMathBlockClosing(raw, ptr, cur) {
			return cur + 2 - ptr
		}
		cur += 2
	}
}

func consumeCodeBlock(raw []rune, ptr int) (advance int) {
	if !unescaped(raw, ptr, "```") {
		return 0
	}
	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "`")))
	cur := ptr + length
	pos := strings.Index(string(raw[cur:]), "```")
	if pos < 0 {
		return len(raw) - ptr
	}
	cur += len([]rune(string(string(raw[cur:])[:pos]))) // closing の "```" の最初の "`"
	closingLength := len(raw[cur:]) - len([]rune(strings.TrimLeft(string(raw[cur:]), "`")))

	// inline の場合は opening bracket と closing bracket は同じ長さでなければならない
	if !strings.ContainsRune(string(raw[ptr:cur+closingLength]), '\n') {
		if length != closingLength {
			return 0
		} else {
			return cur + closingLength - ptr
		}
	} else { // 複数行の場合は, closing bracket の長さは opening bracket の長さ以上でなければいけない
		pos := strings.Index(string(raw[cur:]), strings.Repeat("`", length))
		if pos < 0 {
			return len(raw) - ptr
		}
		cur += len([]rune(string(string(raw[cur:])[:pos])))
		closingLength := len(raw[cur:]) - len([]rune(strings.TrimLeft(string(raw[cur:]), "`")))
		return cur + closingLength - ptr
	}
}

func getH1(content []rune) string {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	if !scanner.Scan() {
		return ""
	}
	line := strings.Trim(scanner.Text(), " \t")
	runes := []rune(line)
	if !(len(runes) >= 3 && string(runes[:2]) == "# ") {
		return ""
	}
	return strings.TrimLeft(line, "# \t")
}

// yaml front matter と本文を切り離す
func splitMarkdown(content []rune) ([]rune, []rune) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	// 冒頭の改行をスキップ
	for scanner.Scan() {
		if scanner.Text() != "" {
			break
		}
	}

	// --- が見つからなかったら front matter なし
	if scanner.Text() != "---" {
		return nil, content
	}

	// --- が見つかるまで front matter に追加していく
	frontMatter := make([]rune, 0)
	endFound := false
	for scanner.Scan() {
		if scanner.Text() == "---" {
			endFound = true
			break
		} else {
			frontMatter = append(frontMatter, []rune(scanner.Text())...)
			frontMatter = append(frontMatter, '\n')
		}
	}

	if !endFound {
		return nil, content
	}

	body := make([]rune, 0)
	for scanner.Scan() {
		body = append(body, []rune(scanner.Text())...)
		body = append(body, '\n')
	}
	return frontMatter, body
}

func findPath(name string) string {
	var filename string
	switch filepath.Ext(name) {
	case "":
		filename = name + ".md"
	case ".md":
		filename = name
	}
	return filename
}

func splitDisplayName(fullname string) (identifier string, displayname string) {
	position := strings.Index(fullname, "|")
	if position < 0 {
		return fullname, ""
	} else {
		identifier := strings.Trim(string(fullname[:position]), " \t")
		displayname := strings.TrimLeft(string(fullname[position:]), "|")
		displayname = strings.Trim(displayname, " \t")
		return identifier, displayname
	}
}

func splitFragment(identifier string) (fileId string, fragment string) {
	position := strings.Index(identifier, "#")
	if position < 0 {
		return identifier, ""
	} else {
		fileId := strings.Trim(string(identifier[:position]), " \t")
		fragment := strings.TrimLeft(string(identifier[position:]), "#")
		fragment = strings.Trim(fragment, " \t")
		return fileId, fragment
	}
}

func genHugoLink(content string) (link string) {
	identifier, displayName := splitDisplayName(content)
	fileId, fragment := splitFragment(identifier)
	path := findPath(fileId)

	if displayName == "" {
		if fragment == "" {
			displayName = fileId
		} else {
			displayName = fmt.Sprintf("%s > %s", fileId, fragment)
		}
	}

	var ref string
	if fragment != "" {
		ref = fmt.Sprintf("%s#%s", path, fragment)
	} else {
		ref = path
	}

	return fmt.Sprintf("[%s]({{< ref \"%s\" >}})", displayName, ref)
}
