package main

import (
	"testing"
)

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
			input:           "\n\n---\ntitle: \"This is a test\"\n---\n# This is a test\n",
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

func TestReplace(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n",
		},
		{
			input: "# This is a markdown file for test.\n \\#todo #123 \n## h2\n",
			want:  "# This is a markdown file for test.\n #todo  \n## h2\n",
		},
		{
			input: "# This is a markdown file for test.\n##todo #123 ###456\n",
			want:  "# This is a markdown file for test.\n##todo  ###456\n",
		},
		{
			input: "# This is a markdown file for test.\n`#todo` #123 \n## h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n`#todo`  \n## h2\n### h3\n#### h4\n",
		},
		{
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n```\n#todo #123\n```\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n```\n#todo #123\n```\n",
		},
		{
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n $#todo$\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n $#todo$\n",
		},
		{
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n $ #todo $\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n $  $\n",
		},
	}

	for _, tt := range cases {
		if got := replace([]rune(tt.input)); string(got) != tt.want {
			t.Errorf("[ERROR] got: %q, want: %q", string(got), tt.want)
		}
	}
}
