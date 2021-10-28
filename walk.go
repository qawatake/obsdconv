package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/qawatake/obsd2hugo/convert"
)

func walk(flags *flagBundle) error {
	err := filepath.Walk(flags.src, func(path string, info fs.FileInfo, err error) error {
		rpath, err := filepath.Rel(flags.src, path)
		if err != nil {
			return err
		}
		rpath = filepath.Clean(rpath)
		newpath := flags.dst + "/" + rpath
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
				return handleErr(path, err)
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

	body, err = converts(body, vault, &title, tags, *flags)
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

func handleErr(path string, err error) error {
	orgErr := errors.Cause(err)
	if e, ok := orgErr.(convert.ErrConvert); !ok {
		return fmt.Errorf("[ERROR] path: %s | %v", path, err)
	} else {
		line := e.Line()
		e := errors.Cause(e.Source())
		if ee, ok := e.(convert.ErrTransform); !ok {
			return fmt.Errorf("[ERROR] path: %s, line: %d | %w", path, line, e)
		} else {
			switch ee.Kind() {
			case convert.ERR_KIND_INVALID_INTERNAL_LINK_CONTENT:
				fmt.Fprintf(os.Stderr, "[ERROR] path: %s, line: %d | invalid internal link content found\n", path, line)
				return nil
			default:
				return fmt.Errorf("[ERROR] path: %s, line: %d | unexpected error occurred: %w", path, line, e)
			}
		}
	}
}
