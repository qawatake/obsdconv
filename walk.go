package main

import (
	"fmt"
	"io"
	"os"

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

	var frontmatter frontMatter
	if p.flags.title {
		frontmatter.title = bodyOutput.title
	}
	if p.flags.alias {
		frontmatter.alias = bodyOutput.title
	}
	if p.flags.cptag {
		for key := range bodyOutput.tags {
			frontmatter.tags = append(frontmatter.tags, key)
		}
	}
	yml, err = convertYAML(yml, frontmatter, p.flags)
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
