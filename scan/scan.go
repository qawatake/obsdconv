package scan

import (
	"net/url"
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

func ScanTag(raw []rune, ptr int) (advance int, tag string) {
	if !(unescaped(raw, ptr, "#") && len(raw[ptr:]) > 1) {
		return 0, ""
	}

	// # の直前はスペースだけ
	if !(precededBy(raw, ptr, []string{" ", "\t", "\n"}) || ptr == 0) {
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

func ScanExternalLinkHead(raw []rune, ptr int) (advance int, displayName string) {
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
			ref = strings.Trim(string(content), " \t\r\n")
			return advance, ref
		}
		cur++ // ")" の次
	}

}

// [google](https:google.com "title") の (https:google.com "title") の部分をスキャン
func ScanExternalLinkTail(raw []rune, ptr int) (advance int, ref string, title string) {
	if !(unescaped(raw, ptr, "(") && len(raw[ptr:]) >= 2) {
		return 0, "", ""
	}
	cur := ptr + 1 // opening の ( の直後

	// " " と \t をスキップ
	for raw[cur] == ' ' || raw[cur] == '\t' {
		cur++
		if cur >= len(raw) {
			return 0, "", ""
		}
	}

	// url を抽出
	adv := scanURL(raw, cur)
	if adv == 0 && !unescaped(raw, cur, ")") {
		return 0, "", ""
	}
	next := cur + adv
	ref = string(raw[cur:next])
	cur = next // url 部分の直後

	// " " と \t をスキップ
	if cur >= len(raw) {
		return 0, "", ""
	}
	for raw[cur] == ' ' || raw[cur] == '\t' {
		cur++
		if cur >= len(raw) {
			return 0, "", ""
		}
	}

	// title を抽出
	adv, title = scanLinkTitle(raw, cur)
	if adv == 0 && !unescaped(raw, cur, ")") {
		return 0, "", ""
	}
	cur += adv // "title" の直後

	if unescaped(raw, cur, ")") {
		cur++ // closing の ) の直後
		return cur - ptr, ref, title
	} else {
		return 0, "", ""
	}
}

// url をスキャン
func scanURL(raw []rune, ptr int) (advance int) {
	cur := ptr
	for cur < len(raw) && isLetterForUrl(raw[cur]) {
		cur++
	}
	urlcandidate := string(raw[ptr:cur])
	if _, err := url.Parse(urlcandidate); err != nil {
		return 0
	}
	return cur - ptr
}

// [google](https:google.com "title") の "title" の部分をスキャン
func scanLinkTitle(raw []rune, ptr int) (advance int, title string) {
	if !(unescaped(raw, ptr, "\"") && len(raw[ptr:]) >= 2) {
		return 0, ""
	}

	cur := ptr + 1 // opening の " の直後
	for cur < len(raw) {
		if unescaped(raw, cur, "\"") {
			title = string(raw[ptr+1 : cur])
			cur++ // closing の " の直後
			return cur - ptr, title
		}
		cur++
	}
	return 0, ""
}

func validExternalLinkTailContent(content string) bool {
	if strings.Contains(content, "\r\n\r\n") || strings.Contains(content, "\n\n") {
		return false
	}
	uri := strings.Trim(content, " \t\r\n")
	return validURI(uri)
}

func ScanExternalLink(raw []rune, ptr int) (advance int, displayName string, ref string, title string) {
	cur := ptr
	adv, displayName := ScanExternalLinkHead(raw, cur)
	if adv == 0 {
		return 0, "", "", ""
	}
	cur += adv
	// adv, ref = scanExternalLinkTail(raw, cur)
	adv, ref, title = ScanExternalLinkTail(raw, cur)
	if adv == 0 {
		return 0, "", "", ""
	}
	cur += adv
	return cur - ptr, displayName, ref, title
}

func ScanExternalLinkVar(raw []rune, ptr int) (advance int) {
	cur := ptr
	adv, _ := ScanExternalLinkHead(raw, cur)
	if adv == 0 {
		return 0
	}
	cur += adv // 前半の closing の ] の直後
	if cur >= len(raw) {
		return 0
	}
	adv, _ = ScanExternalLinkHead(raw, cur)
	if adv == 0 {
		return 0
	}
	cur += adv // 後半の closing の ] の直後
	return cur - ptr
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

	// 終わりの "$$" は行末までに空白しかゆるされない
	cur := closingPtr + 2 // 終わりの "$$" の直後
	adv := indexInRunes(raw[cur:], "\n")
	if adv < 0 {
		return false
	}
	lineTail := cur + adv

	return strings.Trim(string(raw[cur:lineTail]), " \r\n") == ""
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

func scanInlineCodeBlock(raw []rune, ptr int) (advance int) {
	if !unescaped(raw, ptr, "```") {
		return 0
	}

	openingLength := scanRepeat(raw, ptr, "`")
	cur := ptr + openingLength
	adv := indexInRunes(raw[cur:], "```")
	if adv < 0 {
		return len(raw) - ptr
	}
	cur += adv // closing の "```" の最初の "`"

	// inline かどうかチェック
	if strings.ContainsRune(string(raw[ptr:cur]), '\n') {
		return 0
	}

	// opening bracket と closing bracket は同じ長さでなければならない
	closingLength := scanRepeat(raw, cur, "`")
	if openingLength != closingLength {
		return 0
	}

	return cur + closingLength - ptr
}

func scanMultilineCodeBlock(raw []rune, ptr int) (advance int) {
	if !unescaped(raw, ptr, "```") {
		return 0
	}

	openingLength := scanRepeat(raw, ptr, "`")
	cur := ptr + openingLength
	adv := indexInRunes(raw[cur:], "```")
	if adv < 0 {
		return len(raw) - ptr
	}
	cur += adv // closing の "```" の最初の "`"

	if !strings.ContainsRune(string(raw[ptr:cur]), '\n') {
		return 0
	}

	for {
		adv = indexInRunes(raw[cur:], strings.Repeat("`", openingLength))
		if adv < 0 {
			return len(raw) - ptr
		}
		cur += adv
		closingLength := scanRepeat(raw, cur, "`")

		if precededBy(raw, cur, []string{"\n"}) {
			cur += closingLength // closing の "```" の直後 ("`" の個数は3個以上の不定)
			return cur - ptr
		}
		cur += closingLength
	}
}

func ScanCodeBlock(raw []rune, ptr int) (advance int) {
	if advance := scanInlineCodeBlock(raw, ptr); advance > 0 {
		return advance
	}
	if advance := scanMultilineCodeBlock(raw, ptr); advance > 0 {
		return advance
	}
	return 0
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

// scan until line end
func ScanExternalLinkVarDef(raw []rune, ptr int) (advance int, displayName string, ref string) {
	adv, displayName := ScanExternalLinkHead(raw, ptr)
	if adv == 0 {
		return 0, "", ""
	}

	// not preceded by non-\n
	if ptr != 0 && !precededBy(raw, ptr, []string{"\n"}) {
		return 0, "", ""
	}
	cur := ptr + adv // closing ] の直後

	if cur >= len(raw) {
		return 0, "", ""
	}
	if raw[cur] != ':' {
		return 0, "", ""
	}
	cur++

	if cur >= len(raw) {
		return 0, "", ""
	}

	// trim spaces
	for {
		if raw[cur] != ' ' && raw[cur] != '\t' {
			break
		}
		cur++
		if cur >= len(raw) {
			return 0, "", ""
		}
	}

	refHead := cur // ref 部分の先頭
	for {
		if unicode.IsSpace(raw[cur]) {
			break
		}
		// ref 部分に [, ] は許されない
		if raw[cur] == '[' || raw[cur] == ']' {
			return 0, "", ""
		}
		cur++
		if cur >= len(raw) { // ref の後ろがなかった場合
			return cur - ptr, displayName, string(raw[refHead:cur])
		}
	}
	ref = string(raw[refHead:cur])

	// ref の直後が行末かどうかチェック
	// cur = ref の直後
	if raw[cur] == '\n' {
		cur++ // \n の直後
		return cur - ptr, displayName, ref
	} else if unescaped(raw, cur, "\r\n") {
		cur += 2 // \r\n の直後
		return cur - ptr, displayName, ref
	}

	// ref の直後から行末までに文字がないことをチェック
	// cur = ref の直後
	for {
		if raw[cur] == '\n' {
			cur++ // \n の直後
			return cur - ptr, displayName, ref
		}
		if !unicode.IsSpace(raw[cur]) {
			return 0, "", ""
		}
		cur++
		if cur >= len(raw) {
			return cur - ptr, displayName, ref
		}
	}
}

// [google]:https://google.com はひとまとめのグループを形成する
// グループ内部では改行区切り
// グループの直前には空行が必要
func ScanExternalLinkVarDefGroup(raw []rune, ptr int) (advance int, ptrs []int) {
	return 0, nil
}
