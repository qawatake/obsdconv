package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/qawatake/obsdconv/convert"
)

func TestSplitMarkdown(t *testing.T) {
	cases := []struct {
		input           string
		wantFrontMatter string
		wantBody        string
	}{
		{
			input:           "---\ntitle: \"This is a test\"\n---\n# This is a test\n",
			wantFrontMatter: "title: \"This is a test\"\n",
			wantBody:        "# This is a test\n",
		},
		{
			input:           "---\ntitle: \"This is a test\"\n # This is a test.\n",
			wantFrontMatter: "",
			wantBody:        "---\ntitle: \"This is a test\"\n # This is a test.\n",
		},
		{
			input:           "---\ntitle: \"This is a test ---\"\n---\n# This is a test.\n",
			wantFrontMatter: "title: \"This is a test ---\"\n",
			wantBody:        "# This is a test.\n",
		},
		{
			input:           "---\n---\n# This is a test\n",
			wantFrontMatter: "",
			wantBody:        "# This is a test\n",
		},
	}

	for _, tt := range cases {
		gotFrontMatter, gotBody := splitMarkdown([]rune(tt.input))
		if string(gotFrontMatter) != tt.wantFrontMatter {
			t.Errorf("[ERROR] got %q, want: %q", string(gotFrontMatter), tt.wantFrontMatter)
		}
		if string(gotBody) != tt.wantBody {
			t.Errorf("[ERROR] got: %q, want: %q", string(gotBody), tt.wantBody)
		}
	}
}

func TestConvertBody(t *testing.T) {
	const (
		test_CONVERT_BODY_DIR = "testdata/convertbody"
	)
	cases := []struct {
		name         string
		rootDir      string
		srcDir       string
		rawFileName  string
		cptag        bool
		rmtag        bool
		cmmt         bool
		title        bool
		link         bool
		wantTitle    string
		wantTags     []string
		wantFileName string
	}{
		{
			name:         "-cptag -title",
			rootDir:      "cptag_title",
			srcDir:       "src",
			rawFileName:  "src.md",
			cptag:        true,
			title:        true,
			wantTitle:    "test source file <<>>",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
		{
			name:         "-rmtag -title -cptag",
			rootDir:      "rmtag_title_cptag",
			srcDir:       "src",
			rawFileName:  "src.md",
			rmtag:        true,
			title:        true,
			cptag:        true,
			wantTitle:    "test source file <<>>",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
		{
			name:         "-rmtag -title -cptag -cmmt",
			rootDir:      "rmtag_title_cptag_cmmt",
			srcDir:       "src",
			rawFileName:  "src.md",
			rmtag:        true,
			title:        true,
			cptag:        true,
			cmmt:         true,
			wantTitle:    "test source file <<>>",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
		{
			name:         "-rmtag -title -cptag -cmmt -link",
			rootDir:      "rmtag_title_cptag_cmmt_link",
			srcDir:       "src",
			rawFileName:  "src.md",
			rmtag:        true,
			title:        true,
			cptag:        true,
			cmmt:         true,
			link:         true,
			wantTitle:    "test source file <<>>",
			wantTags:     []string{"obsidian", "test"},
			wantFileName: "want.md",
		},
	}

	// src ディレクトリ作成
	originalPath := filepath.Join(test_CONVERT_BODY_DIR, "src.md")
	originalFile, err := os.Open(originalPath)
	if err != nil {
		t.Fatalf("[FATAL] failed to open: %v", err)
	}
	originalContent, err := io.ReadAll(originalFile)
	originalFile.Close()
	if err != nil {
		t.Fatalf("[FATAL] failed to read: %v", err)
	}
	for _, tt := range cases {
		srcDirPath := filepath.Join(test_CONVERT_BODY_DIR, tt.rootDir, "src")
		if err := os.RemoveAll(srcDirPath); err != nil {
			t.Fatalf("[FATAL] failed to remove tmp dir at the beginning: %v", err)
		}
		if err := os.Mkdir(srcDirPath, 0o777); err != nil {
			t.Fatalf("[FATAL] failed to create tmp dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(srcDirPath, "image.png"), nil, 0o666); err != nil {
			t.Fatalf("[FATAL] failed to write: %v", err)
		}
		if err := os.WriteFile(filepath.Join(srcDirPath, "test.md"), nil, 0o666); err != nil {
			t.Fatalf("[FATAL] failed to write: %v", err)
		}
		if err := os.WriteFile(filepath.Join(srcDirPath, "src.md"), originalContent, 0o666); err != nil {
			t.Fatalf("[FATAL] failed to write: %v", err)
		}
	}

	// テスト部
	for _, tt := range cases {
		vault := filepath.Join(test_CONVERT_BODY_DIR, tt.rootDir, tt.srcDir)
		db := convert.NewPathDB(vault)
		c := NewBodyConverterImpl(db, tt.cptag, tt.rmtag, tt.cmmt, tt.title, tt.link)

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

		output, gotTitle, gotTags, err := c.ConvertBody([]rune(string(raw)))
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}

		// 取得した title の確認
		if gotTitle != tt.wantTitle {
			t.Errorf("[ERROR | title - %s] got: %q, want: %q", tt.name, gotTitle, tt.wantTitle)
		}

		// 取得した tag のチェック
		for _, tag := range tt.wantTags {
			if _, ok := gotTags[tag]; !ok {
				t.Errorf("[ERROR | tag - %s] tag: %s not found", tt.name, tag)
			}
			delete(gotTags, tag)
		}
		if len(gotTags) > 0 {
			t.Errorf("[ERROR | tag - %s] got unexpected tags: %v", tt.name, gotTags)
		}

		// file content の確認
		wantFileName := filepath.Join(test_CONVERT_BODY_DIR, tt.rootDir, tt.wantFileName)
		wantFile, err := os.Open(wantFileName)
		if err != nil {
			t.Fatalf("[FATAL | %s] failed to open %s: %v", tt.name, wantFileName, err)
		}
		wantText, err := io.ReadAll(wantFile)
		if err != nil {
			t.Fatalf("[FATAL | %s] failed to read: %s", tt.name, wantFileName)
		}
		wantFile.Close()
		if string(output) != string(wantText) {
			gotscanner := bufio.NewScanner(bytes.NewReader([]byte(string(output))))
			wantscanner := bufio.NewScanner(bytes.NewReader(wantText))
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

	// src ディレクトリを削除
	for _, tt := range cases {
		srcDirPath := filepath.Join(test_CONVERT_BODY_DIR, tt.rootDir, "src")
		if err := os.RemoveAll(srcDirPath); err != nil {
			t.Fatalf("[FATAL] failed to remove tmp dir at the end: %v", err)
		}
	}
}

func TestConvertYAML(t *testing.T) {
	cases := []struct {
		name        string
		publishable bool
		raw         []byte
		title       string
		alias       string
		tags        []string
		want        string
	}{
		{
			name: "no overlap",
			raw: []byte(`cssclass: index-page
publish: true`),
			title: "211027",
			alias: "today",
			tags:  []string{"todo", "math"},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "title overlaps",
			raw: []byte(`cssclass: index-page
publish: true
title: 211026`),
			title: "211027",
			alias: "today",
			tags:  []string{"todo", "math"},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "add aliases",
			raw: []byte(`cssclass: index-page
publish: true
aliases:
- today
`),
			title: "211027",
			alias: "birthday",
			tags:  []string{"todo", "math"},
			want: `aliases:
- today
- birthday
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "alias coincides",
			raw: []byte(`cssclass: index-page
publish: true
aliases:
- today
`),
			title: "211027",
			alias: "today",
			tags:  []string{"todo", "math"},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "add tags",
			raw: []byte(`cssclass: index-page
publish: true
tags:
- book
`),
			title: "211027",
			alias: "today",
			tags:  []string{"todo", "math"},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- book
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "tags overlap",
			raw: []byte(`cssclass: index-page
publish: true
tags:
- book
- math
`),
			title: "211027",
			alias: "today",
			tags:  []string{"todo", "math"},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- book
- math
- todo
title: "211027"
`,
		},
		//////////
		{
			name: "publishable",
			raw: []byte(`cssclass: index-page
publish: true
`),
			title:       "211027",
			alias:       "today",
			tags:        []string{"todo", "math"},
			publishable: true,
			want: `aliases:
- today
cssclass: index-page
draft: false
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "not publishable",
			raw: []byte(`cssclass: index-page
publish: false`),
			title:       "211027",
			alias:       "today",
			tags:        []string{"todo", "math"},
			publishable: true,
			want: `aliases:
- today
cssclass: index-page
draft: true
publish: false
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "no publish field",
			raw: []byte(`cssclass: index-page
`),
			title:       "211027",
			alias:       "today",
			tags:        []string{"todo", "math"},
			publishable: true,
			want: `aliases:
- today
cssclass: index-page
draft: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "draft field wins publish field",
			raw: []byte(`cssclass: index-page
publish: true
draft: true`),
			title:       "211027",
			alias:       "today",
			tags:        []string{"todo", "math"},
			publishable: true,
			want: `aliases:
- today
cssclass: index-page
draft: true
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "no publishable flag",
			raw: []byte(`cssclass: index-page
publish: true
`),
			title: "211027",
			alias: "today",
			tags:  []string{"todo", "math"},
			// flags: &flagBundle{},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
	}

	for _, tt := range cases {
		yc := NewYamlConverterImpl(tt.publishable)
		got, err := yc.ConvertYAML(tt.raw, tt.title, tt.alias, tt.tags)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}
		if string(got) != tt.want {
			t.Errorf("[ERROR | %s]\ngot:\n%q\nwant:\n%q", tt.name, string(got), tt.want)
		}
	}
}
