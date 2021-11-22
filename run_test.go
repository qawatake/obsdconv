package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestRun(t *testing.T) {
	rootDir := "sample"
	src := "src"
	dst := "dst"
	tmp := "tmp"
	cases := []struct {
		name        string
		cmdflags    map[string]string
		argVersion  string
		wantDstDir  string
		wantVersion string
	}{
		{
			name: "-version",
			cmdflags: map[string]string{
				FLAG_VERSION: "1",
			},
			argVersion:  "1.0.0",
			wantVersion: "v1.0.0\n",
		},
		{
			name: "-obs",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(rootDir, "obs", src),
				FLAG_DESTINATION:    filepath.Join(rootDir, "obs", tmp),
				FLAG_OBSIDIAN_USAGE: "1",
			},
			wantDstDir: filepath.Join(rootDir, "obs", dst),
		},
	}

	for _, tt := range cases {
		// flags を設定
		flags := new(flagBundle)
		flagset := flag.NewFlagSet(fmt.Sprintf("TestSetFlags | %s", tt.name), flag.ExitOnError)
		initFlags(flagset, flags)
		for cmdname, cmdvalue := range tt.cmdflags { // flag.Parse() に相当
			flagset.Set(cmdname, cmdvalue)
		}
		setFlags(flagset, flags)

		versionBuf := new(bytes.Buffer)
		err := run(tt.argVersion, flags, versionBuf)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected err occurred: %v", tt.name, err)
		}
		if gotVersion := versionBuf.String(); gotVersion != "" {
			if gotVersion != tt.wantVersion {
				t.Errorf("[ERROR | version // %s] got: %q, want: %q", tt.name, gotVersion, tt.wantVersion)
			}
			continue
		}

		if !equalDirContent(flags.dst, tt.wantDstDir) {
			t.Fatalf("[ERROR | content // %s]", tt.name)
		}
	}
}

func equalDirContent(dir1, dir2 string) bool {
	return true
}

func equalFileContent(file1, file2 io.Reader) (eq bool, err error) {
	// // check if file1 and files are not directories
	// if info, err := os.Lstat(file1); err != nil {
	// 	return false, errors.Wrapf(err, "failed to get info about %s", file1)
	// } else if info.IsDir() {
	// 	return false, fmt.Errorf("%s is a directory", file1)
	// }
	// if info, err := os.Lstat(file2); err != nil {
	// 	return false, errors.Wrapf(err, "failed to get info about %s", file2)
	// } else if info.IsDir() {
	// 	return false, fmt.Errorf("%s is a directory", file2)
	// }

	// f1, err := os.Open(file1)
	// if err != nil {
	// 	return false, errors.Wrapf(err, "failed to open %s", file1)
	// }
	// f2, err := os.Open(file2)
	// if err != nil {
	// 	return false, errors.Wrapf(err, "failed to open %s", file2)
	// }
	const (
		chuncksize = 1000
	)
	chunk1 := make([]byte, chuncksize)
	chunk2 := make([]byte, chuncksize)

	for {
		size1, err1 := file1.Read(chunk1)
		size2, err2 := file2.Read(chunk2)

		if err1 != nil && err1 != io.EOF {
			return false, errors.Wrap(err1, "failed to read file1")
		}
		if err2 != nil && err2 != io.EOF {
			return false, errors.Wrap(err1, "failed to read file2")
		}

		if (err1 == io.EOF && err2 != io.EOF) || (err1 != io.EOF && err2 == io.EOF) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}

		if !bytes.Equal(chunk1[:size1], chunk2[:size2]) {
			return false, nil
		}
	}
}

func TestEqualFileContent(t *testing.T) {
	cases := []struct {
		content1 string
		content2 string
		want     bool
	}{
		{
			content1: "abc\ndef\n",
			content2: "abc\ndef\n",
			want:     true,
		},
		{
			content1: "abc\ndef\n",
			content2: "abc\r\ndef\r\n",
			want:     false,
		},
		{
			content1: "abc\ndef\n",
			content2: "abc\ndef\nghi\n",
			want:     false,
		},
		{
			content1: "abc\ndef\n",
			content2: "abc\n",
			want:     false,
		},
	}

	for _, tt := range cases {
		if got, err := equalFileContent(strings.NewReader(tt.content1), strings.NewReader(tt.content2)); err != nil {
			t.Fatalf("[FATAL] unexpected err occurred: %v", err)
		} else if got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v\ncontent1: %q\ncontent2: %q", got, tt.want, string(tt.content1), string(tt.content2))
		}
	}
}
