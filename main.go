package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	RuneDollar = 0x24 // $
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
	newContent := replace(body)
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

func replace(content []rune) []rune {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	newContent := make([]rune, 0)
	inCodeBlock := false
	inMathBlock := false
	for scanner.Scan() {
		newLine := make([]rune, 0)
		line := []rune(scanner.Text())

		// コードブロック内
		if inCodeBlock {
			if strings.Trim(string(line), " \t") == "```" {
				inCodeBlock = false
			}
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		// コードブロックに入る
		if strings.Trim(string(line), " \t") == "```" {
			inCodeBlock = true
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		// math ブロック内
		if inMathBlock {
			if strings.Trim(string(line), " \t") == "$$" {
				inMathBlock = false
			}
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		// math ブロックに入る
		if strings.Trim(string(line), " \t") == "$$" {
			inMathBlock = true
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		id := 0
		for id < len(line) {

			// inline ブロック
			if advance := consumeInlineBlock(line[id:]); advance > 0 {
				newLine = append(newLine, line[id:id+advance]...)
				id += advance
				continue
			}

			// inline math
			if advance := consumeInlineMath(line[id:]); advance > 0 {
				newLine = append(newLine, line[id:id+advance]...)
				id += advance
				continue
			}

			// エスケープ
			if advance, escaped := consumeEscaped(line[id:]); advance > 0 {
				newLine = append(newLine, escaped...)
				id += advance
				continue
			}

			if advance := consumeRepeat(line[id:], "#"); advance > 1 {
				newLine = append(newLine, line[id:id+advance]...)
				id += advance
				continue
			}

			if advance, _ := consumeTag(line[id:]); advance > 0 {
				id += advance
				continue
			}

			// internl link [[]]
			if advance, content := consumeInternalLink(line[id:]); advance > 0 {
				if content == "" { // [[ ]] はスキップ
					id += advance
					continue
				}
				link := genHugoLink(content)
				newLine = append(newLine, []rune(link)...)
				id += advance
				continue
			}

			// 普通の文字
			newLine = append(newLine, line[id])
			id++
		}
		newContent = append(newContent, newLine...)
		newContent = append(newContent, '\n')
	}
	return newContent
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
