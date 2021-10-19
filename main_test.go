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

func TestGetH1(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "#     This is a markdown file for test.   \t \n # This is a second title. \n ## h2",
			want:  "This is a markdown file for test.",
		},
		{
			input: "#       \n # this is a second title.",
			want:  "",
		},
	}

	for _, tt := range cases {
		if got := getH1([]rune(tt.input)); got != tt.want {
			t.Errorf("[ERROR] got: %q, want: %q", got, tt.want)
		}
	}

}
