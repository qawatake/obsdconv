package main

import (
	"fmt"
	"log"
	"os"
)

const (
	RuneDollar = 0x24 // $
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("引数の個数が不正です")
	}

	filename := os.Args[1]

	file, err := os.Open(filename)
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

func replace(raw []rune) (output []rune) {
	output = make([]rune, 0)
	cur := 0
	for cur < len(raw) {
		if advance := scanCodeBlock(raw, cur); advance > 0 {
			next := cur + advance
			output = append(output, raw[cur:next]...)
			cur = next
			continue
		}

		if advance := scanComment(raw, cur); advance > 0 {
			next := cur + advance
			output = append(output, raw[cur:next]...)
			cur = next
			continue
		}

		if advance := scanMathBlock(raw, cur); advance > 0 {
			next := cur + advance
			output = append(output, raw[cur:next]...)
			cur = next
			continue
		}

		if advance, _, _ := scanExternalLink(raw, cur); advance > 0 {
			next := cur + advance
			output = append(output, raw[cur:next]...)
			cur = next
			continue
		}

		if advance, content := scanInternalLink(raw, cur); advance > 0 {
			if content == "" { // [[ ]] はスキップ
				cur += advance
				continue
			}
			link := genHugoLink(content)
			output = append(output, []rune(link)...)
			cur += advance
			continue
		}

		if advance := scanInlineCode(raw, cur); advance > 0 {
			next := cur + advance
			output = append(output, raw[cur:next]...)
			cur = next
			continue
		}

		if advance := scanInlineMath(raw, cur); advance > 0 {
			next := cur + advance
			output = append(output, raw[cur:next]...)
			cur = next
			continue
		}

		if advance := scanRepeat(raw, cur, "#"); advance > 1 {
			next := cur + advance
			output = append(output, raw[cur:next]...)
			cur = next
			continue
		}

		if advance, _ := scanTag(raw, cur); advance > 0 {
			next := cur + advance
			cur = next
			continue
		}

		// 普通の文字
		output = append(output, raw[cur])
		cur++
	}
	return output
}
