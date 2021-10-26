package main

import (
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
	cases := []struct {
		name  string
		vault string
		raw   []rune
		want  []rune
	}{
		{},
		{},
	}

	for _, tt := range cases {
		c := NewLinkConverter(tt.vault)
		if got := c.Convert(tt.raw); string(got) != string(tt.want) {
			t.Errorf("[ERROR | %v]\n\tgot: %q\n\twant: %q", tt.name, got, tt.want)
		}
	}
}

func TestCommentEraser(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		want []rune
	}{
		{},
		{},
	}

	c := NewCommentEraser()

	for _, tt := range cases {
		if got := c.Convert(tt.raw); string(got) != string(tt.want) {
			t.Errorf("[ERROR | %v]\n\tgot: %q\n\twant: %q", tt.name, got, tt.want)
		}
	}
}
