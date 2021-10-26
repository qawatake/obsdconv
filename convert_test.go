package main

import "testing"

func TestTagRemover(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		want []rune
	}{
		{},
		{},
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
		tags     map[string]struct{}
		raw      []rune
		wantTags []string
	}{
		{},
		{},
	}

	for _, tt := range cases {
		c := NewTagFinder(tt.tags)
		if got := c.Convert(tt.raw); string(got) != string(tt.raw) {
			t.Errorf("[ERROR | ouput - %v]\n\tgot: %q\n\twant: %q", tt.name, string(got), string(tt.raw))
		}
		for _, tag := range tt.wantTags {
			if _, ok := tt.tags[tag]; !ok {
				t.Errorf("[ERROR]| %s] tag not found: %s", tt.name, tag)
			} else {
				delete(tt.tags, tag)
			}
		}
		if len(tt.tags) > 0 {
			for _, tag := range tt.tags {
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
		{},
		{},
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
