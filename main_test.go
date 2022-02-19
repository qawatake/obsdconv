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

	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
)

func TestRun(t *testing.T) {
	sampleDir := "sample"
	testdataDir := filepath.Join("testdata", "run")
	src := "src"
	dst := "dst"
	tmp := "tmp"
	tgt := "tgt"
	errMsgs := map[convert.ErrKind]string{
		convert.ERR_KIND_UNEXPECTED:                       "unexpected error",
		convert.ERR_KIND_INVALID_INTERNAL_LINK_CONTENT:    "invalid internal link content",
		convert.ERR_KIND_NO_REF_SPECIFIED_IN_OBSIDIAN_URL: "no ref specified in obsidian url",
		convert.ERR_KIND_UNEXPECTED_HREF:                  "unexpected href",
		convert.ERR_KIND_INVALID_SHORTHAND_OBSIDIAN_URL:   "invalid shorthand obsidian url",
		convert.ERR_KIND_PATH_NOT_FOUND:                   "path not found",
	}

	cases := []struct {
		name            string
		cmdflags        map[string]string
		version         string
		wantDstDir      string
		wantVersionText string
		wantErrKinds    []convert.ErrKind
	}{
		{
			name: "-version",
			cmdflags: map[string]string{
				FLAG_VERSION: "1",
			},
			version:         "1.0.0",
			wantVersionText: "v1.0.0",
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
		{
			name: "-obs (ignore folders)",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "obs_ignore", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "obs_ignore", tmp),
				FLAG_OBSIDIAN_USAGE: "1",
			},
			wantDstDir: filepath.Join(testdataDir, "obs_ignore", dst),
		},
		{
			name: "-std -strictref=0 (ignore folders)",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "std_strictref0_ignore", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "std_strictref0_ignore", tmp),
				FLAG_STANDARD_USAGE: "1",
				FLAG_STRICT_REF:     "0",
			},
			wantDstDir: filepath.Join(testdataDir, "std_strictref0_ignore", dst),
		},
		{
			name: "-std (ignore folders)",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "std_ignore", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "std_ignore", tmp),
				FLAG_STANDARD_USAGE: "1",
			},
			wantDstDir: filepath.Join(testdataDir, "std_ignore", dst),
			wantErrKinds: []convert.ErrKind{
				convert.ERR_KIND_PATH_NOT_FOUND,
			},
		},
		{
			name: "-obs -synctag",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "obs_synctag", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "obs_synctag", tmp),
				FLAG_OBSIDIAN_USAGE: "1",
				FLAG_SYNC_TAGS:      "1",
			},
			wantDstDir: filepath.Join(testdataDir, "obs_synctag", dst),
		},
		{
			name: "-alias",
			cmdflags: map[string]string{
				FLAG_SOURCE:       filepath.Join(testdataDir, "alias", src),
				FLAG_DESTINATION:  filepath.Join(testdataDir, "alias", tmp),
				FLAG_COPY_ALIASES: "1",
			},
			wantDstDir: filepath.Join(testdataDir, "alias", dst),
		},
		{
			name: "-synctlal",
			cmdflags: map[string]string{
				FLAG_SOURCE:             filepath.Join(testdataDir, "synctlal", src),
				FLAG_DESTINATION:        filepath.Join(testdataDir, "synctlal", tmp),
				FLAG_SYNC_TITLE_ALIASES: "1",
			},
			wantDstDir: filepath.Join(testdataDir, "synctlal", dst),
		},
		{
			name: "-tgt -std -strictref=0",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "tgt", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "tgt", tmp),
				FLAG_TARGET:         filepath.Join(testdataDir, "tgt", tgt),
				FLAG_STANDARD_USAGE: "1",
				FLAG_STRICT_REF:     "0",
			},
			wantDstDir: filepath.Join(testdataDir, "tgt", dst),
		},
		{
			name: "-obs -remapmkey",
			cmdflags: map[string]string{
				FLAG_SOURCE:          filepath.Join(testdataDir, "obs_remapkey", src),
				FLAG_DESTINATION:     filepath.Join(testdataDir, "obs_remapkey", tmp),
				FLAG_OBSIDIAN_USAGE:  "1",
				FLAG_REMAP_META_KEYS: "aliases:xaliases,image:meta_image,x:",
			},
			wantDstDir: filepath.Join(testdataDir, "obs_remapkey", dst),
		},
		{
			name: "-obs -filter",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(testdataDir, "obs_filter", src),
				FLAG_DESTINATION:    filepath.Join(testdataDir, "obs_filter", tmp),
				FLAG_OBSIDIAN_USAGE: "1",
				FLAG_FILTER:         "(key1||!key2)&&key3",
			},
			wantDstDir: filepath.Join(testdataDir, "obs_filter", dst),
		},
	}

	for _, tt := range cases {
		// config を設定
		config := new(configuration)
		flagset := flag.NewFlagSet(fmt.Sprintf("TestSetFlags | %s", tt.name), flag.ExitOnError)
		initFlags(flagset, config)
		for cmdname, cmdvalue := range tt.cmdflags { // flag.Parse() に相当
			flagset.Set(cmdname, cmdvalue)
		}
		setConfig(flagset, config)

		gotVersionText, gotBufferredErrs, err := run(tt.version, config)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected err occurred: %v", tt.name, err)
		}

		// check version text
		if gotVersionText != tt.wantVersionText {
			t.Errorf("[ERROR | version // %s] got: %q, want: %q", tt.name, gotVersionText, tt.wantVersionText)
			continue
		}

		// check non-interfering errors
		caught := false
		for id := 0; ; id++ {
			if id == len(gotBufferredErrs) {
				if id < len(tt.wantErrKinds) {
					t.Errorf("[ERROR | %s] expected err did not occurred: %s", tt.name, errMsgs[tt.wantErrKinds[id]])
					caught = true
				}
				break
			}
			if id == len(tt.wantErrKinds) {
				if id < len(gotBufferredErrs) {
					t.Fatalf("[FATAL | %s] unexpected buffered err occurred: %v", tt.name, gotBufferredErrs[id])
				}
				break
			}

			gotErr := gotBufferredErrs[id]
			wantErrKind := tt.wantErrKinds[id]
			e, ok := errors.Cause(gotErr).(convert.ErrTransform)
			if !ok {
				t.Fatalf("[Fatalf | %s] unexpected buffered err occurred: %v", tt.name, gotErr)
			}
			if e.Kind() != wantErrKind {
				t.Errorf("[ERROR | buffered err - %s]\n got: %v\nwant:%s", tt.name, gotErr, errMsgs[wantErrKind])
				caught = true
				break
			}
		}
		if caught {
			continue
		}

		// check generated directries
		if msg, err := compareDirContent(config.dst, tt.wantDstDir); err != nil {
			t.Fatalf("[FATAL | content // %s] unexpected error occurred: %v", tt.name, err)
		} else if msg != "" {
			t.Errorf("[ERROR | content // %s] %s", tt.name, msg)
			continue
		}

		// remove generated directory
		if err := os.RemoveAll(config.dst); err != nil {
			t.Fatalf("[FATAL | %s] failed to remove generated directory %s", tt.name, config.dst)
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

func TestCheckFilter(t *testing.T) {
	cases := []struct {
		fm     map[interface{}]interface{}
		filter string
		want   bool
	}{
		{
			fm: map[interface{}]interface{}{
				"key1": true,
				"key2": true,
			},
			filter: "key1&&key2",
			want:   true,
		},
		{
			fm: map[interface{}]interface{}{
				"key1": true,
				"key2": true,
				"key3": false,
			},
			filter: "key1&&key2||key3",
			want:   true,
		},
		{
			fm: map[interface{}]interface{}{
				"key1": true,
				"key2": false,
				"key3": false,
			},
			filter: "key1&&key2||key3",
			want:   false,
		},
		{
			fm: map[interface{}]interface{}{
				"key1": false,
				"key2": true,
				"key3": false,
			},
			filter: "key1&&key2||key3",
			want:   false,
		},
		{
			fm: map[interface{}]interface{}{
				"key1": false,
				"key2": false,
				"key3": true,
			},
			filter: "key1&&key2||key3",
			want:   true,
		},
		{
			fm: map[interface{}]interface{}{
				"key1": true,
				"key2": true,
				"key3": false,
			},
			filter: "key1&&!key2||key3",
			want:   false,
		},
		{
			fm: map[interface{}]interface{}{
				"key1": false,
				"key2": false,
				"key3": true,
			},
			filter: "key1&&(key2||key3)",
			want:   false,
		},
	}

	for _, tt := range cases {
		got, err := checkFilter(tt.fm, tt.filter)
		if err != nil {
			t.Fatalf("[FATAL] unexpected error occurred\n%v\nfm: %v\nfilter: %s", err, tt.fm, tt.filter)
		}
		if got != tt.want {
			t.Errorf("[ERROR] filter: %s\nfm: %v", tt.filter, tt.fm)
		}
	}
}
