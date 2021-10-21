package main

import (
	"testing"
)

func TestUnescaped(t *testing.T) {
	cases := []struct {
		name      string
		argRaw    []byte
		argSubstr string
		want      bool
	}{
		{name: "\\ x 1", argRaw: []byte("\\$"), argSubstr: "$", want: false},
		{name: "\\ x 2", argRaw: []byte("\\\\$"), argSubstr: "$", want: true},
		{name: "\\ x 3", argRaw: []byte("\\\\\\$"), argSubstr: "$", want: false},
	}

	for _, tt := range cases {
		if got := unescaped(tt.argRaw, len(tt.argRaw)-1, tt.argSubstr); got != tt.want {
			t.Errorf("[ERROR | %s] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestPrecededBy(t *testing.T) {
	cases := []struct {
		argRaw []byte
		argSs  []string
		want   bool
	}{
		{
			argRaw: []byte("######"),
			argSs:  []string{"##"},
			want:   true,
		},
		{
			argRaw: []byte("\\#"),
			argSs:  []string{"\\"},
			want:   true,
		},
		{
			argRaw: []byte("x $"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []byte("x\t$"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []byte("x$"),
			argSs:  []string{" ", "\t"},
			want:   false,
		},
		{
			argRaw: []byte("x\n\n$"),
			argSs:  []string{" ", "\t", "\n\n", "\r\n\r\n"},
			want:   true,
		},
		{
			argRaw: []byte("x\r\n\r\n$"),
			argSs:  []string{" ", "\t", "\n\n", "\r\n\r\n"},
			want:   true,
		},
	}

	for _, tt := range cases {
		if got := precededBy(tt.argRaw, len(tt.argRaw)-1, tt.argSs); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v", got, tt.want)
		}
	}
}

func TestFollowedBy(t *testing.T) {
	cases := []struct {
		argRaw []byte
		argSs  []string
		want   bool
	}{
		{
			argRaw: []byte("$ x"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []byte("$\tx"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []byte("$x"),
			argSs:  []string{" ", "\t"},
			want:   false,
		},
		{
			argRaw: []byte("$\n\nx"),
			argSs:  []string{" ", "\t", "\n\n", "\r\n\r\n"},
			want:   true,
		},
		{
			argRaw: []byte("$\r\n\r\nx"),
			argSs:  []string{" ", "\t", "\n\n", "\r\n\r\n"},
			want:   true,
		},
	}

	for _, tt := range cases {
		if got := followedBy(tt.argRaw, 0, tt.argSs); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v", got, tt.want)
		}
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
		name  string
		input string
		want  string
	}{
		{
			name:  "tag: simple",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n",
		},
		{
			name:  "tag: escaped by \\",
			input: "# This is a markdown file for test.\n \\#todo #123 \n## h2\n",
			want:  "# This is a markdown file for test.\n #todo  \n## h2\n",
		},
		{
			name:  "tag: multiple #'s",
			input: "# This is a markdown file for test.\n##todo #123 ###456\n",
			want:  "# This is a markdown file for test.\n##todo  ###456\n",
		},
		{
			name:  "tag: inline code block",
			input: "# This is a markdown file for test.\n`#todo` #123 \n## h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n`#todo`  \n## h2\n### h3\n#### h4\n",
		},
		{
			name:  "tag: code block",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n```\n#todo #123\n```\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n```\n#todo #123\n```\n",
		},
		{
			name:  "tag: code block with specified lang",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n```python\n#todo #123\n```\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n```python\n#todo #123\n```\n",
		},
		{
			name:  "tag: code block with long line",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n````\n```\n#todo #123\n````\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n````\n```\n#todo #123\n````\n",
		},
		{
			name:  "tag: inline math",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n $#todo$\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n $#todo$\n",
		},
		{
			name:  "tag: inline math not working",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n $ #todo $\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n $  $\n",
		},
		{
			name:  "tag: display math block",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n$$\n #todo #123\n$$\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n$$\n #todo #123\n$$\n",
		},
		{
			name:  "tag: href",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n [#test](https://google.com#fragment)\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n [#test](https://google.com#fragment)\n",
		},
		{
			name:  "internal link: simple",
			input: "# This is a markdown file for test.\n [[211020-142030]] \n##  h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n [211020-142030]({{< ref \"211020-142030.md\" >}}) \n##  h2\n### h3\n#### h4\n",
		},
		{
			name:  "internal link: empty",
			input: "# This is a markdown file for test.\n [[]] \n##  h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n [[]] \n##  h2\n### h3\n#### h4\n",
		},
		{
			name:  "internal link: blank",
			input: "# This is a markdown file for test.\n [[ ]] \n##  h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n  \n##  h2\n### h3\n#### h4\n",
		},
		{
			name:  "internal link: display name",
			input: "# This is a markdown file for test.\n [[211020-142030 | This is a test | yes | ]] \n##  h2\n### h3\n#### h4\n",
			want:  "# This is a markdown file for test.\n [This is a test | yes |]({{< ref \"211020-142030.md\" >}}) \n##  h2\n### h3\n#### h4\n",
		},
		{
			name:  "internal link: fragment",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n [[211020-165952#fragment]]\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n [211020-165952 > fragment]({{< ref \"211020-165952.md#fragment\" >}})\n",
		},
		{
			name:  "internal link: fragment with displayname",
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n [[211020-165952#fragment | This is a test]]",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n [This is a test]({{< ref \"211020-165952.md#fragment\" >}})\n",
		},
	}

	for _, tt := range cases {
		if got := replace([]rune(tt.input)); string(got) != tt.want {
			t.Errorf("[ERROR] %s\n\tgot:  %q\n\twant: %q", tt.name, string(got), tt.want)
		}
	}
}
