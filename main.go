package main

import (
	"fmt"
	"log"
	"os"
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

	newContent := removeTags([]rune(string(content)))

	newFile, err := os.Create("new." + filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer newFile.Close()
	fmt.Println(string(newContent))
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
