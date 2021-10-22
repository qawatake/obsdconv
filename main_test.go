package main

import (
	"testing"
)

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
			want:  "# This is a markdown file for test.\n \\#todo  \n## h2\n",
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
			input: "# This is a markdown file for test.\n#todo #123 \n## h2\n### h3\n#### h4\n [[211020-165952#fragment | This is a test]]\n",
			want:  "# This is a markdown file for test.\n  \n## h2\n### h3\n#### h4\n [This is a test]({{< ref \"211020-165952.md#fragment\" >}})\n",
		},
	}

	for _, tt := range cases {
		got, err := replace([]rune(tt.input))
		if err != nil {
			t.Errorf("[FAIL] %v", err)
		}
		if string(got) != tt.want {
			t.Errorf("[ERROR] %s\n\tgot:  %q\n\twant: %q", tt.name, string(got), tt.want)
		}
	}
}
