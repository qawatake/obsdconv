package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	sampleDir := "sample"
	testdataDir := filepath.Join("testdata", "run")
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
			name: "[SAMPLE] -obs",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(sampleDir, "obs", src),
				FLAG_DESTINATION:    filepath.Join(sampleDir, "obs", tmp),
				FLAG_OBSIDIAN_USAGE: "1",
			},
			wantDstDir: filepath.Join(sampleDir, "obs", dst),
		},
		{
			name: "[SAMPLE] -std",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(sampleDir, "std", src),
				FLAG_DESTINATION:    filepath.Join(sampleDir, "std", tmp),
				FLAG_STANDARD_USAGE: "1",
			},
			wantDstDir: filepath.Join(sampleDir, "std", dst),
		},
		{
			name: "[SAMPLE] -std -rmh1",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(sampleDir, "std_rmh1", src),
				FLAG_DESTINATION:    filepath.Join(sampleDir, "std_rmh1", tmp),
				FLAG_STANDARD_USAGE: "1",
				FLAG_REMOVE_H1:      "1",
			},
			wantDstDir: filepath.Join(sampleDir, "std_rmh1", dst),
		},
		{
			name: "[SAMPLE] -std -pub",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(sampleDir, "std_pub", src),
				FLAG_DESTINATION:    filepath.Join(sampleDir, "std_pub", tmp),
				FLAG_STANDARD_USAGE: "1",
				FLAG_PUBLISHABLE:    "1",
			},
			wantDstDir: filepath.Join(sampleDir, "std_pub", dst),
		},
		{
			name: "-std -strictref=0",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "std_strictref0", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "std_strictref0", tmp),
				FLAG_STANDARD_USAGE: "1",
				FLAG_STRICT_REF:     "0",
			},
			wantDstDir: filepath.Join(testdataDir, "std_strictref0", dst),
		},
		{
			name: "-obs",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "obs", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "obs", tmp),
				FLAG_OBSIDIAN_USAGE: "1",
			},
			wantDstDir: filepath.Join(testdataDir, "obs", dst),
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

		if msg, err := compareDirContent(flags.dst, tt.wantDstDir); err != nil {
			t.Fatalf("[FATAL | content // %s] unexpected error occurred: %v", tt.name, err)
		} else if msg != "" {
			t.Errorf("[ERROR | content // %s] %s", tt.name, msg)
		}

		// remove generated directory
		if err := os.RemoveAll(flags.dst); err != nil {
			t.Fatalf("[FATAL | %s] failed to remove generated directory %s", tt.name, flags.dst)
		}
	}
}

// if contents fo two directories are the same, msg = ""
func compareDirContent(dir1, dir2 string) (msg string, err error) {
	const (
		capacity = 100
	)
	data1 := make([]struct {
		path string
		info fs.FileInfo
	}, 0, capacity)
	data2 := make([]struct {
		path string
		info fs.FileInfo
	}, 0, capacity)

	err = filepath.Walk(dir1, func(path string, info fs.FileInfo, err error) error {
		data1 = append(data1, struct {
			path string
			info fs.FileInfo
		}{
			path: path,
			info: info,
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(dir2, func(path string, info fs.FileInfo, err error) error {
		data2 = append(data2, struct {
			path string
			info fs.FileInfo
		}{
			path: path,
			info: info,
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(data1) != len(data2) {
		return fmt.Sprintf("number of files in directories are diffrent - %s: %d, %s: %d", dir1, len(data1), dir2, len(data2)), nil
	}

	for id := 1; id < len(data1); id++ { // exclude root directory
		d1 := data1[id]
		d2 := data2[id]
		path1, err := filepath.Rel(dir1, d1.path)
		if err != nil {
			log.Fatal(err)
		}
		path2, err := filepath.Rel(dir2, d2.path)
		if err != nil {
			log.Fatal(err)
		}
		if path1 != path2 {
			return fmt.Sprintf("paths are different - %s vs %s", d1.path, d2.path), nil
		}

		// directories
		if d1.info.IsDir() && d2.info.IsDir() {
			if msg, err := compareDirContent(d1.path, d2.path); err != nil {
				log.Fatal(err)
			} else if msg != "" {
				return msg, nil
			}
			continue
		}

		// regular file and directory
		if d1.info.IsDir() || d2.info.IsDir() {
			if d1.info.IsDir() {
				return fmt.Sprintf("%s is a file but %s is a directory", d2.path, d1.path), nil
			}
			if d2.info.IsDir() {
				return fmt.Sprintf("%s is a file but %s is a directory", d1.path, d2.path), nil
			}
		}

		var scanner1 *bufio.Scanner
		var scanner2 *bufio.Scanner
		if b, err := os.ReadFile(d1.path); err != nil {
			log.Fatal(err)
		} else {
			scanner1 = bufio.NewScanner(bytes.NewReader(b))
		}
		if b, err := os.ReadFile(d2.path); err != nil {
			log.Fatal(err)
		} else {
			scanner2 = bufio.NewScanner(bytes.NewReader(b))
		}

		line := 1
		for {
			if !scanner1.Scan() {
				if scanner2.Scan() {
					return fmt.Sprintf("path:%s, line: %d, more lines than %s", d2.path, line, d1.path), nil
				}
				break
			}
			if !scanner2.Scan() {
				if scanner1.Scan() {
					return fmt.Sprintf("path:%s, line: %d, more lines than %s", d1.path, line, d2.path), nil
				}
				break
			}
			if !bytes.Equal(scanner1.Bytes(), scanner2.Bytes()) {
				return fmt.Sprintf("line: %d in %s and %s are different:\n%q\n%q", line, d1.path, d2.path, scanner1.Text(), scanner2.Text()), nil
			}
			line++
		}
	}
	return "", nil
}
