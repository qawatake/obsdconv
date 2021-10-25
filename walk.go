package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func walk(src string, dst string) error {
	err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		rpath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		rpath = filepath.Clean(rpath)
		newpath := dst + "/" + rpath
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
			if err := convert(src, newpath, file); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
