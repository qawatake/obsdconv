package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("引数の個数が不正です")
	}

	filename := os.Args[1]

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		log.Fatalf("os.Create failed %v", err)
	}
	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatalf("file.Stat failed %v", err)
	}

	content := make([]byte, fileinfo.Size())
	_, err = file.Read(content)
	if err != nil {
		log.Fatalf("file.Write failed: %v", err)
	}

	frontMatter, body := splitMarkdown([]rune(string(content)))
	newContent := removeTags(body)
	title := getH1(newContent)
	fmt.Printf("Title: %v\n", title)
	fmt.Printf("Front Matter: <<\n%v>>\n", string(frontMatter))

	newFile, err := os.Create("new." + filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer newFile.Close()
	newFile.Write([]byte(string(newContent)))
}

func removeTags(content []rune) []rune {
	newContent := make([]rune, 0, len(content))

	id := 0
	for id < len(content) {
		if content[id] == '#' && id < len(content)-1 && content[id+1] != '#' && (unicode.IsLetter(content[id+1]) || unicode.IsNumber(content[id+1])) {
			p := id
			for p < len(content) && !unicode.IsSpace(rune(content[p])) {
				p++
			}
			id = p
			continue
		}

		newContent = append(newContent, content[id])
		id++
	}
	return newContent
}

func getH1(content []rune) string {
	if !(len(content) >= 2 && content[0] == '#' && content[1] == ' ') {
		return ""
	}

	c := content[2:]
	// 冒頭の空白をスキップ
	for len(c) > 0 {
		if !(c[0] == ' ' || c[0] == '\t') {
			break
		}
		c = c[1:]
	}

	// タイトルを取得
	titleEnd := 0
	id := 0
	for id < len(c) {
		if c[id] == '\r' || c[id] == '\n' {
			break
		}
		id++
		if unicode.IsPrint(c[id-1]) && !unicode.IsSpace(c[id-1]) {
			titleEnd = id
		}
	}
	title := string(c[:titleEnd])
	return title
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
	}
	body = append(body, '\n')
	return frontMatter, body
}
