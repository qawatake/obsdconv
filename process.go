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
	Process(orgpath, newpath string) (err error)
}

type ProcessorImpl struct {
	flags *flagBundle
	BodyConverter
}

func NewProcessorImpl(flags *flagBundle) *ProcessorImpl {
	p := new(ProcessorImpl)
	db := convert.NewPathDB(flags.src)
	p.BodyConverter = NewBodyConverterImpl(db, flags.cptag, flags.rmtag, flags.cmmt, flags.title, flags.link)
	p.flags = flags
	return p
}

func (p *ProcessorImpl) Process(orgpath, newpath string) error {
	err := p.generate(orgpath, newpath)
	if err == nil {
		return nil
	}

	// 予想済みのエラーの場合は処理を止めずに, エラー出力だけする
	public, debug := handleErr(orgpath, err)
	if public == nil && debug == nil {
		return nil
	}

	if p.flags.debug {
		return debug
	} else {
		return public
	}
}

func (p *ProcessorImpl) generate(orgpath, newpath string) (err error) {
	readFrom, err := os.Open(orgpath)
	if err != nil {
		return errors.Errorf("failed to open %s", orgpath)
	}
	content, err := io.ReadAll(readFrom)
	if err != nil {
		return errors.New("failed to read file")
	}
	readFrom.Close()

	yml, body := splitMarkdown([]rune(string(content)))

	output, title, tags, err := p.ConvertBody(body)
	if err != nil {
		return errors.Wrap(err, "failed to convert")
	}

	yc := NewYamlConverterImpl(p.flags.publishable)
	newtags := make([]string, 0, len(tags))
	for tg := range tags {
		newtags = append(newtags, tg)
	}

	yml, err = yc.convertYAML(yml, title, title, newtags)
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

	fmt.Fprintf(writeTo, "---\n%s---\n%s", string(yml), string(output))
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

func handleErr(path string, err error) (public error, debug error) {
	orgErr := errors.Cause(err)
	e, ok := orgErr.(convert.ErrConvert)
	if !ok {
		e := fmt.Errorf("[FATAL] path: %s | %v", path, err)
		return e, e
	}

	line := e.Line()
	ee, ok := errors.Cause(e.Source()).(convert.ErrTransform)
	if !ok {
		public = fmt.Errorf("[FATAL] path: %s, around line: %d | failed to convert", path, line)
		debug = fmt.Errorf("[FATAL] path: %s, around line: %d | cause of source of ErrConvert does not implement ErrTransform: ErrConvert: %w", path, line, e)
		return public, debug
	}

	if ee.Kind() >= convert.ERR_KIND_UNEXPECTED {
		public = fmt.Errorf("[FATAL] path: %s, around line: %d | failed to convert", path, line)
		debug = fmt.Errorf("[FATAL] path: %s, around line: %d | undefined kind of ErrTransform: ErrTransform: %w", path, line, ee)
		return public, debug
	}

	// 想定済みのエラー
	fmt.Fprintf(os.Stderr, "[ERROR] path: %s, around line: %d | %v\n", path, line, ee)
	return nil, nil
}
