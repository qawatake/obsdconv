package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
)

type Processor interface {
	Process(path, newpath string) (err error)
}

type ProcessorImpl struct {
	flags *flagBundle
	BodyConverter
}

func NewProcessorImpl(flags *flagBundle) *ProcessorImpl {
	bc := new(BodyConverterImpl)
	bc.db = convert.NewPathDB(flags.src)
	bc.flags = flags
	p := new(ProcessorImpl)
	p.flags = flags
	p.BodyConverter = bc
	return p
}

func (p *ProcessorImpl) Process(path, newpath string) (err error) {
	readFrom, err := os.Open(path)
	if err != nil {
		return errors.Errorf("failed to open %s", path)
	}
	content, err := io.ReadAll(readFrom)
	if err != nil {
		return errors.New("failed to read file")
	}
	readFrom.Close()

	yml, body := splitMarkdown([]rune(string(content)))

	bodyOutput, err := p.ConvertBody(body)
	if err != nil {
		return errors.Wrap(err, "failed to convert")
	}

	yc := NewYamlConverterImpl(p.flags)
	if p.flags.title {
		yc.title = bodyOutput.title
	}
	if p.flags.alias {
		yc.alias = bodyOutput.title
	}
	if p.flags.cptag {
		for key := range bodyOutput.tags {
			yc.tags = append(yc.tags, key)
		}
	}
	yml, err = yc.convertYAML(yml)
	if err != nil {
		return errors.Wrap(err, "failed to convert yaml")
	}

	// os.Create によってファイルの内容は削除されるので,
	// 変換がすべて正常に行われた後で, 書き込み先のファイルを開く
	writeTo, err := os.Create(newpath)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s", newpath)
	}
	defer writeTo.Close()

	fmt.Fprintf(writeTo, "---\n%s---\n%s", string(yml), string(bodyOutput.text))
	return nil
}

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
