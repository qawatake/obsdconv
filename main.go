package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
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
		inline := false
		inlineMath := false
		var inlineMathfrom int
		for id < len(line) {
			// インラインブロック内
			if inline {
				if line[id] == '`' {
					inline = false
				}
				newLine = append(newLine, line[id])
				id++
				continue
			}

			// インラインブロックに入る
			if line[id] == '`' {
				inline = true
				newLine = append(newLine, line[id])
				id++
				continue
			}

			// inline math ブロック内
			if inlineMath {
				if line[id] == RuneDollar && id > 0 && !unicode.IsSpace(line[id-1]) {
					inlineMath = false
					newLine = append(newLine, line[inlineMathfrom:id+1]...)
					id++
					continue
				} else if id >= len(line)-1 {
					id = inlineMathfrom
					inlineMath = false
					newLine = append(newLine, RuneDollar)
					id++
					continue
				}
				id++
				continue
			}

			// inline math ブロックに入る
			if line[id] == RuneDollar && id+1 < len(line) && !unicode.IsSpace(line[id+1]) {
				inlineMath = true
				inlineMathfrom = id
				id++
				continue
			}

			// エスケープ
			if line[id] == '\\' && id+1 < len(line) {
				switch line[id+1] {
				case '#':
					newLine = append(newLine, '#')
					id += 2
					continue
				}
			}

			// タグ
			if line[id] == '#' && id+1 < len(line) && unicode.IsGraphic(line[id+1]) && !unicode.IsSpace(line[id+1]) {
				if line[id+1] == '#' { // ###todo はそのまま ###todo として扱われる
					p := id
					for p < len(line) && line[p] == '#' {
						p++
					}
					newLine = append(newLine, line[id:p]...)
					id = p
					continue
				} else {
					p := id + 1
					for p < len(line) && (unicode.IsLetter(line[p]) || unicode.IsNumber(line[p])) {
						p++
					}
					id = p
					continue
				}
				continue
			}

			// [[]]
			if len(line[id:]) >= 5 && string(line[id:id+2]) == "[[" {
				position := strings.Index(string(line[id+2:]), "]]")
				if position < 0 {
					newLine = append(newLine, line[id:id+2]...)
					id += 2
					continue
				}

				if string(line[id:id+4]) == "[[]]" {
					newLine = append(newLine, line[id:id+4]...)
					id += 4
					continue
				}

				name := strings.Trim(string(string(line[id+2:])[:position]), " \t")
				id +=  2 + len([]rune(string(string(line[id+2:])[:position]))) + 2
				if name != "" {
					path := findPath(name)
					newLine = append(newLine, []rune(fmt.Sprintf("[%s]({{< ref %s >}})", name, path))...)
				}
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
