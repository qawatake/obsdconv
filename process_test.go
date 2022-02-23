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

func TestExamineYaml(t *testing.T) {
	cases := []struct {
		name        string
		publishable bool
		yml         []byte
		want        bool
	}{
		{
			name:        "not publishable",
			publishable: false,
			yml: []byte(`publish: false
draft: true`),
			want: true,
		},
		{
			name:        "publishable && publish: true",
			publishable: true,
			yml:         []byte(`publish: true`),
			want:        true,
		},
		{
			name:        "publishable && publish: false",
			publishable: true,
			yml:         []byte(`publish: false`),
			want:        false,
		},
		{
			name:        "publishable && draft: false",
			publishable: true,
			yml:         []byte(`draft: false`),
			want:        true,
		},
		{
			name:        "publishable && draft: true",
			publishable: true,
			yml:         []byte(`draft: true`),
			want:        false,
		},
	}

	for _, tt := range cases {
		got, err := newYamlExaminatorImpl("", tt.publishable).ExamineYaml(tt.yml)
		if err != nil {
			t.Fatalf("[FATAL | %s] ExamineYaml failed: %v", tt.name, err)
		}
		if got != tt.want {
			t.Errorf("[ERROR | %s] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestConvertBody(t *testing.T) {
	const (
		test_CONVERT_BODY_DIR = "testdata/convertbody"
	)
	cases := []struct {
		name            string
		rootDir         string
		srcDir          string
		rawFileName     string
		cptag           bool
		rmtag           bool
		cmmt            bool
		title           bool
		link            bool
		rmH1            bool
		formatLink      bool
		formatAnchor    string
		pathPrefixRemap map[string]string
		wantTitle       string
		wantTags        []string
		wantFileName    string
	}{
		{
			name:         "-cptag -title",
			rootDir:      "cptag_title",
			srcDir:       "src",
			rawFileName:  "src.md",
			cptag:        true,
			title:        true,
			wantTitle:    "test source file <<  >>",
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
			wantTitle:    "test source file <<  >>",
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
			wantTitle:    "test source file <<  >>",
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
			formatAnchor: convert.FORMAT_ANCHOR_HUGO,
			wantTitle:    "test source file <<  >>",
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
		c := newBodyConverterImpl(db, tt.cptag, tt.rmtag, tt.cmmt, tt.title, tt.link, tt.rmH1, tt.formatLink, tt.formatAnchor, nil)

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

		// output, gotTitle, gotTags, err := c.ConvertBody([]rune(string(raw)))
		output, aux, err := c.ConvertBody([]rune(string(raw)), tt.rawFileName)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}

		gotTitle := ""
		var gotTags map[string]struct{}
		if v, ok := aux.(*bodyConvAuxOutImpl); !ok {
			t.Fatalf("[FATAL | %s] aux (BodyConvAuxOut) cannot converted to bodyConvAuxOutImpl", tt.name)
		} else {
			gotTitle = v.title
			gotTags = v.tags
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
		synctag     bool
		synctlal    bool
		publishable bool
		remap       map[string]string
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
		{
			name:    "sync tags",
			synctag: true,
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
- todo
- math
title: "211027"
`,
		},
		{
			name:     "sync title and aliases",
			synctlal: true,
			raw: []byte(`aliases:
- existing-alias
- existing-title
cssclass: index-page
publish: true
tags:
- book
title: existing-title
`),
			title: "211027",
			alias: "211027",
			want: `aliases:
- existing-alias
- "211027"
cssclass: index-page
publish: true
tags:
- book
title: "211027"
`,
		},
		{
			name:     "sync title and aliases when h1 = existing title = existing alias",
			synctlal: true,
			raw: []byte(`aliases:
- existing-alias
- existing-title
cssclass: index-page
publish: true
tags:
- book
title: existing-title
`),
			title: "existing-title",
			alias: "existing-title",
			want: `aliases:
- existing-alias
- existing-title
cssclass: index-page
publish: true
tags:
- book
title: existing-title
`,
		},
		{
			name: "remap keys in front matter",
			remap: map[string]string{
				"aliases":  "xaliases",
				"cssclass": "",
			},
			raw: []byte(`aliases:
- existing-alias
cssclass: index-page
publish: true
tags:
- book
title: existing-title
`),
			title: "211208",
			alias: "today",
			want: `publish: true
tags:
- book
title: "211208"
xaliases:
- existing-alias
- today
`,
		},
	}

	for _, tt := range cases {
		yc := newYamlConverterImpl(tt.synctag, tt.synctlal, tt.publishable, tt.remap)
		auxinput := newYamlConvAuxInImpl(tt.title, tt.alias, tt.tags)
		got, err := yc.ConvertYAML(tt.raw, auxinput)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}
		if string(got) != tt.want {
			t.Errorf("[ERROR | %s]\ngot:\n%q\nwant:\n%q", tt.name, string(got), tt.want)
		}
	}
}

func TestPassArg(t *testing.T) {
	cases := []struct {
		name       string
		title      bool
		alias      bool
		iter       int
		frombody   bodyConvAuxOutImpl
		wantToyaml yamlConvAuxInImpl
	}{
		{
			name:  "title & alias & tags",
			title: true,
			alias: true,
			iter:  20,
			frombody: bodyConvAuxOutImpl{
				title: "title",
				tags:  map[string]struct{}{"c": {}, "a": {}, "b": {}},
			},
			wantToyaml: yamlConvAuxInImpl{
				title:   "title",
				alias:   "title",
				newtags: []string{"a", "b", "c"},
			},
		},
		{
			name:  "title",
			title: true,
			frombody: bodyConvAuxOutImpl{
				title: "title",
			},
			wantToyaml: yamlConvAuxInImpl{
				title: "title",
			},
		},
		{
			name:  "alias",
			alias: true,
			frombody: bodyConvAuxOutImpl{
				title: "title",
			},
			wantToyaml: yamlConvAuxInImpl{
				alias: "title",
			},
		},
	}

	for _, tt := range cases {
		iter := tt.iter
		if iter == 0 {
			iter = 1
		}
		for range make([]struct{}, iter) {
			passer := newArgPasserImpl(tt.title, tt.alias)
			got, err := passer.PassArg(&tt.frombody)
			if err != nil {
				t.Fatalf("[FATAL] unexpected error occurred: %v", err)
			}
			gotToyaml, _ := got.(*yamlConvAuxInImpl)
			if gotToyaml.title != tt.wantToyaml.title {
				t.Errorf("[ERROR | title - %s] got: %s, want: %s", tt.name, gotToyaml.title, tt.wantToyaml.title)
			}
			if gotToyaml.alias != tt.wantToyaml.alias {
				t.Errorf("[ERROR | alias - %s] got: %s, want: %s", tt.name, gotToyaml.alias, tt.wantToyaml.alias)
			}
			if len(gotToyaml.newtags) != len(tt.wantToyaml.newtags) {
				t.Errorf("[ERROR | tags - %s] got: %s, want: %s", tt.name, gotToyaml.newtags, tt.wantToyaml.newtags)
				return
			}
			for i, gotTag := range gotToyaml.newtags {
				wantTag := tt.wantToyaml.newtags[i]
				if gotTag != wantTag {
					t.Errorf("[ERROR | tags - %s] got: %s, want: %s", tt.name, gotToyaml.newtags, tt.wantToyaml.newtags)
					return
				}
			}
		}
	}

}

func TestParseRemap(t *testing.T) {
	cases := []struct {
		input   string
		want    map[string]string
		wantErr mainErr
	}{
		{
			input: "image:meta_image,aliases:xaliases",
			want: map[string]string{
				"image":   "meta_image",
				"aliases": "xaliases",
			},
		},
		{
			input: "image:meta_image,aliases:",
			want: map[string]string{
				"image":   "meta_image",
				"aliases": "",
			},
		},
		{
			input:   "this is a bad input",
			wantErr: newMainErrf(MAIN_ERR_KIND_INVALID_REMAP_FORMAT, "invalid remap format"),
		},
		{
			input: "",
			want:  nil,
		},
	}

	for _, tt := range cases {
		got, gotErr := parseRemap(tt.input)
		if gotErr != nil {
			if tt.wantErr == nil {
				t.Fatalf("[FATAL] unexpected error occurred: %v with input: %s", gotErr, tt.input)
			}
			if e, ok := gotErr.(mainErr); !ok {
				t.Fatalf("[FATAL] unexpected error occurred: %v with input: %s", gotErr, tt.input)
			} else if e.Kind() != tt.wantErr.Kind() {
				t.Fatalf("[FATAL] unexpected error occurred: %v with input: %s", gotErr, tt.input)
			}
		} else if tt.wantErr != nil {
			t.Errorf("[ERROR] expected error did not occurr: %v with input: %s", tt.wantErr, tt.input)
		}

		tobeskipped := false
		for wantOldKey, wantNewKey := range tt.want {
			if gotNewKey, ok := got[wantOldKey]; !ok {
				t.Errorf("[ERROR] expected key %s missing for input %s", wantOldKey, tt.input)
				tobeskipped = true
				break
			} else if gotNewKey != wantNewKey {
				t.Errorf("[ERROR] new keys for %s are different. got: %s, want: %s for input: %s", wantOldKey, gotNewKey, wantNewKey, tt.input)
			}
			delete(got, wantOldKey)
		}
		if tobeskipped {
			continue
		}
		if len(got) > 0 {
			t.Errorf("[ERROR] unexpected keys found: %v", got)
		}
	}
}
