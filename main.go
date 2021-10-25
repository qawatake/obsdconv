package main

import (
	"fmt"
	"log"
	"os"
)

const (
	RuneDollar = 0x24 // "$"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("引数の個数が不正です")
	}

	filename := os.Args[1]
	root := "."

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
	title := ""
	tags := make(map[string]struct{})
	c := NewConverter(root, &title, tags)
	newContent := c.Convert(body)
	fmt.Printf("Title: %v\n", title)
	for key, _ := range tags {
		fmt.Printf("[TAG] %s\n", key)
	}
	fmt.Printf("Front Matter: <<\n%v>>\n", string(frontMatter))

	newFile, err := os.Create("new." + filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer newFile.Close()
	newFile.Write([]byte(string(newContent)))
}

func NewConverter(vault string, title *string, tags map[string]struct{}) *Converter {
	*title = ""

	// タグ削除用の Converter を作成
	remover := new(Converter)
	remover.Set(DefaultMiddleware(scanEscaped))
	remover.Set(DefaultMiddleware(scanCodeBlock))
	remover.Set(TransformComment)
	remover.Set(DefaultMiddleware(scanMathBlock))
	remover.Set(TransformExternalLinkFunc(vault))
	remover.Set(TransformInternalLinkFunc(vault))
	remover.Set(TransformEmbedsFunc(vault))
	remover.Set(DefaultMiddleware(scanInlineMath))
	remover.Set(DefaultMiddleware(scanInlineCode))
	remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, tag := scanTag(raw, ptr)
		if advance == 0 {
			return 0, nil
		}
		tags[tag] = struct{}{}
		return advance, nil
	})
	remover.Set(TransformNone)

	// メインの converter を作成
	c := new(Converter)
	c.Set(DefaultMiddleware(scanEscaped))
	c.Set(DefaultMiddleware(scanCodeBlock))
	c.Set(TransformComment)
	c.Set(DefaultMiddleware(scanMathBlock))
	c.Set(TransformExternalLinkFunc(vault))
	c.Set(TransformInternalLinkFunc(vault))
	c.Set(TransformEmbedsFunc(vault))
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, level, _ := scanHeader(raw, ptr)
		if advance == 0 {
			return 0, nil
		}

		tagRemoved := remover.Convert(raw[ptr : ptr+advance])
		if level == 1 && *title == "" {
			_, _, *title = scanHeader(tagRemoved, 0)
		}
		return len(tagRemoved), tagRemoved
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, tag := scanTag(raw, ptr)
		if advance == 0 {
			return 0, nil
		}
		tags[tag] = struct{}{}
		return advance, nil
	})
	c.Set(TransformNone)
	return c
}
