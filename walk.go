package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/qawatake/obsd2hugo/convert"
	"gopkg.in/yaml.v2"
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

func process(vault string, path string, newpath string, flags *flagBundle) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s", path)
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("file.Stat failed %w", err)
	}

	content := make([]byte, fileinfo.Size())
	_, err = file.Read(content)
	if err != nil {
		return fmt.Errorf("file.Write failed: %w", err)
	}

	yml, body := splitMarkdown([]rune(string(content)))
	frontmatter := new(frontMatter)
	title := ""
	tags := make(map[string]struct{})

	newContent, err := converts(body, vault, &title, tags, *flags)
	if err != nil {
		return fmt.Errorf("convert failed: %w", err)
	}

	if err := yaml.Unmarshal(yml, frontmatter); err != nil {
		return fmt.Errorf("yaml.Unmarshal failed: %w", err)
	}
	if flags.title {
		frontmatter.Title = title
	}
	if flags.alias {
		frontmatter.Aliases = append(frontmatter.Aliases, frontmatter.Title)
	}
	if flags.cptag {
		for key := range tags {
			frontmatter.Tags = append(frontmatter.Tags, key)
		}
	}
	yml, err = yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("yaml.Marshal failed: %w", err)
	}

	newfile, err := os.Create(newpath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", newpath, err)
	}
	defer newfile.Close()
	fmt.Fprintf(newfile, "---\n%s---\n%s", string(yml), string(newContent))
	return nil
}

func handleErr(path string, err error) error {
	orgErr := errors.Unwrap(err)
	if e, ok := orgErr.(convert.ErrConvert); !ok {
		return fmt.Errorf("[ERROR] path: %s | %v", path, orgErr)
	} else {
		e.SetPath(path)
		if _, ok := e.(convert.ErrTransform); !ok {
			return fmt.Errorf("[ERROR] path: %s, line: %d | %w", e.Path(), e.Line(), e.Cause())
		} else {
			return fmt.Errorf("[ERROR] path: %s, line: %d | invalid internal link content found", e.Path(), e.Line())
		}
	}
}
