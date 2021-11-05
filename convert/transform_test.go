package convert

import (
	"path/filepath"
	"testing"
)

func TestTransformExternalLink(t *testing.T) {
	testTransformExternalLinkRootDir := filepath.Join("testdata", "transformexternallink")
	cases := []struct {
		name             string
		root             string
		displayName      string
		ref              string
		wantExternalLink string
	}{
		{name: "simple external", root: ".", displayName: "google", ref: "https://google.com", wantExternalLink: "[google](https://google.com)"},
		{name: "filename", root: "filename", displayName: "211024", ref: "test.md", wantExternalLink: "[211024](test.md)"},
		{name: "ref is fileId (filename with the extension removed)", root: "fileid", displayName: "211024", ref: "test", wantExternalLink: "[211024](test.md)"},
		{name: "ref is fileId with fragments", root: "fragments", displayName: "211024", ref: "test#section", wantExternalLink: "[211024](test.md#section)"},
		{name: "obsidian url", root: "obsidianurl", displayName: "open obsidian note", ref: "obsidian://open?vault=obsidian&file=test", wantExternalLink: "[open obsidian note](test.md)"},
		{name: "escaped japanese obsidian url", root: "escaped_obsidianurl", displayName: "日本語のテスト", ref: "obsidian://open?vault=obsidian&file=%E3%83%86%E3%82%B9%E3%83%88", wantExternalLink: "[日本語のテスト](テスト.md)"},
	}

	for _, tt := range cases {
		db := NewPathDB(filepath.Join(testTransformExternalLinkRootDir, tt.root))
		transformer := &ExternalLinkTransformerImpl{PathDB: db}
		got, err := transformer.TransformExternalLink(tt.displayName, tt.ref)
		if err != nil {
			t.Fatalf("[FATAL] | %v] unexpected error ocurred: %v", tt.name, err)
		}
		if got != tt.wantExternalLink {
			t.Errorf("[ERROR | %v]\ngot: %q, want: %q", tt.name, got, tt.wantExternalLink)
		}
	}
}

func TestCurrentLine(t *testing.T) {
	cases := []struct {
		raw  []rune
		ptr  int
		want int
	}{
		{raw: []rune("a\nb\nc\nX"), ptr: 6, want: 4},
		{raw: []rune("a\n\n\n\\n\\n\nX"), ptr: 9, want: 5},
	}
	for _, tt := range cases {
		if got := currentLine(tt.raw, tt.ptr); got != tt.want {
			t.Errorf("[ERROR] got: %d, want: %d with input %q", got, tt.want, string(tt.raw))
		}
	}
}
