package main

import (
	"bufio"
	"strings"
)

// yaml front matter と本文を切り離す
func splitMarkdown(content []rune) (yml []byte, body []rune) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	if !scanner.Scan() {
		return nil, nil
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

	for scanner.Scan() {
		body = append(body, []rune(scanner.Text())...)
		body = append(body, '\n')
	}
	return []byte(string(frontMatter)), body
}
