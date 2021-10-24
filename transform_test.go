package main

import "testing"

func TestTransformExternalLink(t *testing.T) {
	cases := []struct {
		name string
		root string
		raw  []rune
		want []rune
	}{
		{name: "simple external", root: ".", raw: []rune("[google](https://google.com)"), want: []rune("[google](https://google.com)")},
		{name: "filename", root: "testdata/transformexternallink/filename", raw: []rune("[211024](test.md)"), want: []rune("[211024](test.md)")},
		{name: "ref is fileId (filename with the extension removed)", root: "testdata/transformexternallink/fileid", raw: []rune("[211024](test)"), want: []rune("[211024](test.md)")},
		{name: "obsidian url", root: "testdata/findpath/simple", raw: []rune("[open obsidian note](obsidian://open?file=test)"), want: []rune("[open obsidian note](test.md)")},
		{name: "escaped japanese obsidian url", root: "test/data/findpath/simple", raw: []rune("[日本語のテスト](obsidian://open?file=%E3%83%86%E3%82%B9%E3%83%88)"), want: []rune("[日本語のテスト](テスト.md)")},
	}

	for _, tt := range cases {
		if _, got := TransformExternalLinkFunc(tt.root)(tt.raw, 0); string(got) != string(tt.want) {
			t.Errorf("[ERROR | %v]\ngot: %q, want: %q", tt.name, string(got), string(tt.want))
		}
	}
}
