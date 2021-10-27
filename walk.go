package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
			if err := process(flags.src, newpath, flags, file); err != nil {
				return handleErr(path, err)
			}
		}
		return nil
	})
	return err
}

func handleErr(path string, err error) error {
	orgErr := errors.Unwrap(err)
	if e, ok := orgErr.(ErrTransform); !ok {
		return fmt.Errorf("[ERROR] path: %s | %v", path, orgErr)
	} else {
		e.SetPath(path)
		if _, ok := e.(ErrInvalidInternalLinkContent); !ok {
			return fmt.Errorf("[ERROR] path: %s, line: %d | %w", e.Path(), e.Line(), e.Cause())
		} else {
			return fmt.Errorf("[ERROR] path: %s, line: %d | invalid internal link content found", e.Path(), e.Line())
		}
	}
}
