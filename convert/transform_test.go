package convert

import "testing"

func TestTransformExternalLink(t *testing.T) {
	const (
		TEST_TRANSFORM_EXTERNAL_LINK_ROOT_DIR = "../testdata/transformexternallink/"
	)
	cases := []struct {
		name string
		root string
		raw  []rune
		want []rune
	}{
		{name: "simple external", root: ".", raw: []rune("[google](https://google.com)"), want: []rune("[google](https://google.com)")},
		{name: "filename", root: "filename", raw: []rune("[211024](test.md)"), want: []rune("[211024](test.md)")},
		{name: "ref is fileId (filename with the extension removed)", root: "fileid", raw: []rune("[211024](test)"), want: []rune("[211024](test.md)")},
		{name: "ref is fileId with fragments", root: "fragments", raw: []rune("[211024](test#section)"), want: []rune("[211024](test.md#section)")},
		{name: "obsidian url", root: "obsidianurl", raw: []rune("[open obsidian note](obsidian://open?vault=obsidian&file=test)"), want: []rune("[open obsidian note](test.md)")},
		{name: "escaped japanese obsidian url", root: "escaped_obsidianurl", raw: []rune("[日本語のテスト](obsidian://open?vault=obsidian&file=%E3%83%86%E3%82%B9%E3%83%88)"), want: []rune("[日本語のテスト](テスト.md)")},
	}

	for _, tt := range cases {
		_, got, err := TransformExternalLinkFunc(TEST_TRANSFORM_EXTERNAL_LINK_ROOT_DIR+tt.root)(tt.raw, 0)
		if err != nil {
			t.Fatalf("[FATAL] | %v] unexpected error ocurred: %v", tt.name, err)
		}
		if string(got) != string(tt.want) {
			t.Errorf("[ERROR | %v]\ngot: %q, want: %q", tt.name, string(got), string(tt.want))
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
