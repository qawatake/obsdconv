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

func scanInlineBlock(line []rune) (advance int) {
	if line[0] != '`' {
		return 0
	}

	cur := 1
	for cur < len(line) && line[cur] != '`' {
		cur++
	}
	if cur == len(line) {
		return 0
	}
	return cur + 1
}

func consumeInlineCode(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "`") && len(raw[ptr:]) > 1) {
		return 0
	}

	pos := strings.IndexRune(string(raw[ptr+1:]), '`')
	cur := ptr + 1 + len([]rune(string(raw[ptr+1:])[:pos]))
	if precededBy(raw, cur, []string{"\n\n", "\r\n\r\n"}) {
		return 0
	} else {
		return cur - ptr + 1
	}
}

func scanInlineMath(line []rune) (advance int) {
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

func consumeInlineMath(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "$") && !followedBy(raw, ptr, []string{" ", "\t"})) {
		return 0
	}

	cur := ptr + 1
	for cur < len(raw) && !unescaped(raw, cur, "$") {
		pos := strings.IndexRune(string(raw[cur:]), RuneDollar)
		adv := len([]rune(string(string(raw[cur:])[:pos])))
		cur += adv
	}
	if cur >= len(raw) {
		return 0
	}
	if precededBy(raw, cur, []string{"\n\n", "\r\n\r\n", " ", "\t"}) {
		return 0
	} else if unescaped(raw, cur, "$") {
		return cur - ptr + 1
	}
	return 0
}

func scanEscaped(line []rune) (advance int, escaped []rune) {
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

func scanRepeat(line []rune, substr string) (advance int) {
	cur := 0
	length := len([]rune(substr))
	for len(line[cur:]) >= length && string(line[cur:cur+length]) == substr {
		cur += length
	}
	return cur
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

func scanTag(line []rune) (advance int, tag string) {
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

func scanInternalLink(line []rune) (advance int, content string) {
	if !(len(line) >= 5 && string(line[:2]) == "[[") {
		return 0, ""
	}

	position := strings.Index(string(line[2:]), "]]")
	if position <= 0 {
		return 0, ""
	}
	advance = 2 + len([]rune(string(string(line[2:])[:position]))) + 2
	content = strings.Trim(string(string(line[2:])[:position]), " \t")
	return advance, content
}

func scanExternalLink(line []rune) (advance int, displayName string, ref string) {
	if !(line[0] == '[' && len(line) >= 4) {
		return 0, "", ""
	}
	midPosition := strings.Index(string(line[1:]), "](")
	if midPosition < 0 || midPosition == len(line[1:]) {
		return 0, "", ""
	}

	endPosition := strings.Index(string(string(line[1:])[midPosition+2:]), ")")
	if endPosition < 0 {
		return 0, "", ""
	}

	advance = 1 + midPosition + 2 + endPosition + 1
	displayName = string(line[1:midPosition])
	ref = string(line[midPosition+3 : midPosition+endPosition+3])
	return advance, displayName, ref
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
