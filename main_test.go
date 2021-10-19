package main

import "testing"

func TestRemoveTags(t *testing.T) {
	input := "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n"
	want := "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n"
	got := string(removeTags([]rune(input)))
	if got != want {
		t.Errorf("[ERROR] got: %q, want: %q", got, want)
	}
}
