package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertBody(t *testing.T) {
	const (
		test_CONVERT_BODY_DIR = "testdata/convertbody"
	)
	cases := []struct {
		name         string
		rootDir      string
		srcDir       string
		rawFileName  string
		flags        flagBundle
		wantTitle    string
		wantTags     []string
		wantFileName string
	}{
		{
			name:        "-cptag -title",
			rootDir:     "cptag_title",
			srcDir:      "src",
			rawFileName: "src.md",
			flags: flagBundle{
				cptag: true,
				title: true,
			},
			wantTitle:    "test source file",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
		{
			name:        "-rmtag -title -cptag",
			rootDir:     "rmtag_title_cptag",
			srcDir:      "src",
			rawFileName: "src.md",
			flags: flagBundle{
				rmtag: true,
				title: true,
				cptag: true,
			},
			wantTitle:    "test source file <<>>",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
		{
			name:        "-rmtag -title -cptag -cmmt",
			rootDir:     "rmtag_title_cptag_cmmt",
			srcDir:      "src",
			rawFileName: "src.md",
			flags: flagBundle{
				rmtag: true,
				title: true,
				cptag: true,
				cmmt:  true,
			},
			wantTitle:    "test source file <<>>",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
		{
			name:        "-rmtag -title -cptag -cmmt -link",
			rootDir:     "rmtag_title_cptag_cmmt_link",
			srcDir:      "src",
			rawFileName: "src.md",
			flags: flagBundle{
				rmtag: true,
				title: true,
				cptag: true,
				cmmt:  true,
				link:  true,
			},
			wantTitle:    "test source file <<>>",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
	}

	for _, tt := range cases {
		vault := filepath.Join(test_CONVERT_BODY_DIR, tt.rootDir, tt.srcDir)
		srcFileName := filepath.Join(vault, tt.rawFileName)
		srcFile, err := os.Open(srcFileName)
		if err != nil {
			t.Fatalf("[FATAL | %s] failed to open %s: %v", tt.name, srcFileName, err)
		}
		raw, err := io.ReadAll(srcFile)
		if err != nil {
			t.Fatalf("[FATAL | %s] failed to read: %s", tt.name, srcFileName)
		}
		srcFile.Close()

		// 取得した title の確認
		title := ""
		tags := make(map[string]struct{})
		gotOutput, err := convertBody([]rune(string(raw)), vault, &title, tags, tt.flags)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}
		if title != tt.wantTitle {
			t.Errorf("[ERROR | title - %s] got: %q, want: %q", tt.name, title, tt.wantTitle)
		}

		// 取得した tag のチェック
		for _, tag := range tt.wantTags {
			if _, ok := tags[tag]; !ok {
				t.Errorf("[ERROR | tag - %s] tag: %s not found", tt.name, tag)
			}
			delete(tags, tag)
		}
		if len(tags) > 0 {
			t.Errorf("[ERROR | tag - %s] got unexpected tags: %v", tt.name, tags)
		}

		// file content の確認
		wantFileName := filepath.Join(test_CONVERT_BODY_DIR, tt.rootDir, tt.wantFileName)
		wantFile, err := os.Open(wantFileName)
		if err != nil {
			t.Fatalf("[FATAL | %s] failed to open %s: %v", tt.name, wantFileName, err)
		}
		wantOutput, err := io.ReadAll(wantFile)
		if err != nil {
			t.Fatalf("[FATAL | %s] failed to read: %s", tt.name, wantFileName)
		}
		wantFile.Close()
		if string(gotOutput) != string(wantOutput) {
			gotscanner := bufio.NewScanner(bytes.NewReader([]byte(string(gotOutput))))
			wantscanner := bufio.NewScanner(bytes.NewReader(wantOutput))
			linenum := 1
			errDisplayed := false
			for gotscanner.Scan() && wantscanner.Scan() {
				if gotscanner.Text() != wantscanner.Text() {
					t.Errorf("[ERROR | %s] got output differs from wanted output in line %d:\n got: %q\nwant: %q", tt.name, linenum, gotscanner.Text(), wantscanner.Text())
					errDisplayed = true
					break
				}
				linenum++
			}
			if !errDisplayed {
				t.Errorf("[ERROR | %s] output differs from wanted output, but couldn't catch the error line", tt.name)
			}
		}
	}
}
