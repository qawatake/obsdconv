package main

import "testing"

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
