package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/qawatake/obsdconv/convert"
)

func walk(flags *flagBundle) error {
	err := filepath.Walk(flags.src, func(path string, info fs.FileInfo, err error) error {
		rpath, err := filepath.Rel(flags.src, path)
		if err != nil {
			return err
		}
		newpath := filepath.Join(flags.dst, rpath)
		if info.IsDir() {
			if _, err := os.Stat(newpath); !os.IsNotExist(err) {
				return nil
			}
			if err := os.Mkdir(newpath, 0o777); err == nil {
				return nil
			} else {
				return err
			}
		}
		if filepath.Ext(path) != ".md" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			newfile, err := os.Create(newpath)
			if err != nil {
				return err
			}
			defer newfile.Close()
			io.Copy(newfile, file)
		} else {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if err := process(flags.src, path, newpath, flags); err != nil {
				if public, debug := handleErr(path, err); public != nil || debug != nil {
					if flags.debug {
						return debug
					} else {
						return public
					}
				}

			}
		}
		return nil
	})
	return err
}

func process(vault string, path string, newpath string, flags *flagBundle) (err error) {
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
	title := ""
	tags := make(map[string]struct{})

	body, err = convertBody(body, vault, &title, tags, *flags)
	if err != nil {
		return errors.Wrap(err, "failed to convert")
	}

	var frontmatter frontMatter
	if flags.title {
		frontmatter.title = title
	}
	if flags.alias {
		frontmatter.alias = frontmatter.title
	}
	if flags.cptag {
		for key := range tags {
			frontmatter.tags = append(frontmatter.tags, key)
		}
	}
	yml, err = convertYAML(yml, frontmatter, flags)
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

	fmt.Fprintf(writeTo, "---\n%s---\n%s", string(yml), string(body))
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
