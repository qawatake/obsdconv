package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestTagRemover(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		want []rune
	}{
		{name: "simple", raw: []rune("# H1 #todo #obsidian\n## H2\n"), want: []rune("# H1  \n## H2\n")},
		{name: "escaped", raw: []rune("# H1 #todo \\#obsidian\n## H2\n"), want: []rune("# H1  \\#obsidian\n## H2\n")},
		{name: "##", raw: []rune("# H1 #todo ##obsidian\n## H2\n"), want: []rune("# H1  ##obsidian\n## H2\n")},
	}

	c := NewTagRemover()

	for _, tt := range cases {
		if got := c.Convert(tt.raw); string(got) != string(tt.want) {
			t.Errorf("[ERROR | %v]\n\tgot: %q\n\twant: %q", tt.name, string(got), string(tt.want))
		}
	}
}

func TestTagFinder(t *testing.T) {
	cases := []struct {
		name     string
		raw      []rune
		wantTags []string
	}{
		{name: "simple", raw: []rune("# H1 #todo #obsidian\n## H2\n"), wantTags: []string{"todo", "obsidian"}},
		{name: "escaped", raw: []rune("# H1 #todo \\#obsidian\n## H2\n"), wantTags: []string{"todo"}},
		{name: "##", raw: []rune("# H1 #todo ##obsidian\n## H2\n"), wantTags: []string{"todo"}},
	}

	for _, tt := range cases {
		tags := make(map[string]struct{})
		c := NewTagFinder(tags)
		if got := c.Convert(tt.raw); string(got) != string(tt.raw) {
			t.Errorf("[ERROR | ouput - %v]\n\tgot: %q\n\twant: %q", tt.name, string(got), string(tt.raw))
		}
		for _, tag := range tt.wantTags {
			if _, ok := tags[tag]; !ok {
				t.Errorf("[ERROR]| %s] tag not found: %s", tt.name, tag)
			} else {
				delete(tags, tag)
			}
		}
		if len(tags) > 0 {
			for tag := range tags {
				t.Errorf("[ERROR | %s] unexpected tag found: %s", tt.name, tag)
			}
		}
	}
}

func TestTitleFinder(t *testing.T) {
	cases := []struct {
		name      string
		raw       []rune
		wantTitle string
	}{
		{name: "simple", raw: []rune("# H1 #todo #obsidian\n# Second H1\n## H2\n"), wantTitle: "H1 #todo #obsidian"},
		{name: "preceded by \\t", raw: []rune("\t# H1 #todo #obsidian\n# Second H1\n## H2\n"), wantTitle: "Second H1"},
		{name: "preceded by a letter", raw: []rune("x# H1 #todo #obsidian\n# Second H1\n## H2\n"), wantTitle: "Second H1"},
		{name: "escaped", raw: []rune("\\# H1 #todo #obsidian\n# Second H1\n## H2\n"), wantTitle: "Second H1"},
		{name: "separated by \\r\\n", raw: []rune("# H1 #todo #obsidian\r\n# Second H1\n## H2\n"), wantTitle: "H1 #todo #obsidian"},
		{name: "immediate \\r\\n", raw: []rune("#\r\n not H1 #todo #obsidian\n# Second H1\n## H2\n"), wantTitle: "Second H1"},
	}

	for _, tt := range cases {
		gotTitle := ""
		c := NewTitleFinder(&gotTitle)
		if got := c.Convert(tt.raw); string(got) != string(tt.raw) {
			t.Errorf("[ERROR | output - %v]\n\tgot: %q\n\twant: %q", tt.name, got, tt.raw)
		}
		if gotTitle != tt.wantTitle {
			t.Errorf("[ERROR | title - %v] got: %q, want: %q", tt.name, gotTitle, tt.wantTitle)
		}
	}
}

func TestLinkConverter(t *testing.T) {
	const (
		TEST_LINK_CONVERTER_VAULT_DIR = "testdata/linkconverter/"
	)
	cases := []struct {
		name  string
		vault string
		raw   []rune
		want  []rune
	}{
		{name: "simple - external",
			vault: "external",
			raw:   []rune("[google](https://google.com)"),
			want:  []rune("[google](https://google.com)")},
		{name: "filename - external",
			vault: "external/filename",
			raw:   []rune("[211026](test.md)"),
			want:  []rune("[211026](test.md)")},
		{name: "ref is fileId (filename with the extension removed) - external",
			vault: "external/fileid",
			raw:   []rune("[211026](test)"),
			want:  []rune("[211026](test.md)")},
		{name: "ref is fileId with fragments - external",
			vault: "external/fragments",
			raw:   []rune("[211026](test#section)"),
			want:  []rune("[211026](test.md#section)")},
		{name: "obsidian url - external",
			vault: "external/obsidianurl",
			raw:   []rune("[open obsidian note](obsidian://open?vault=obsidian&file=test)"),
			want:  []rune("[open obsidian note](test.md)")},
		{name: "escaped japanese obsidian url - external",
			vault: "external/escaped_obsidianurl",
			raw:   []rune("[日本語のテスト](obsidian://open?vault=obsidian&file=%E3%83%86%E3%82%B9%E3%83%88)"),
			want:  []rune("[日本語のテスト](テスト.md)")},
	}

	for _, tt := range cases {
		c := NewLinkConverter(TEST_LINK_CONVERTER_VAULT_DIR + tt.vault)
		if got := c.Convert(tt.raw); string(got) != string(tt.want) {
			t.Errorf("[ERROR | %v]\n\tgot: %q\n\twant: %q", tt.name, string(got), string(tt.want))
			fmt.Println(filepath.Dir(TEST_LINK_CONVERTER_VAULT_DIR+tt.vault), filepath.Base(TEST_LINK_CONVERTER_VAULT_DIR+tt.vault))
		}
	}
}

func TestCommentEraser(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		want []rune
	}{
		{
			name: "simple",
			raw:  []rune("%%x%%"),
			want: []rune(""),
		},
		{
			name: "long bracket",
			raw:  []rune("%%%x%%%"),
			want: []rune(""),
		},
		{
			name: "longer closing",
			raw:  []rune("%%x%%%"),
			want: []rune("%"),
		},
		{
			name: "longer closing with \\n",
			raw:  []rune("%%\nx\n%%%"),
			want: []rune("%"),
		},
		{
			name: "escaped closing",
			raw:  []rune("%%x\\%%"),
			want: []rune(""),
		},
		{
			name: "escaped opening",
			raw:  []rune("\\%%x%%"),
			want: []rune("\\%%x"),
		},
		{
			name: "no closing",
			raw:  []rune("%%x"),
			want: []rune(""),
		},
	}

	c := NewCommentEraser()

	for _, tt := range cases {
		if got := c.Convert(tt.raw); string(got) != string(tt.want) {
			t.Errorf("[ERROR | %v]\n\tgot: %q\n\twant: %q", tt.name, string(got), string(tt.want))
		}
	}
}
