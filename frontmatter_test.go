package main

import "testing"

func TestConvertYAML(t *testing.T) {
	cases := []struct {
		name        string
		raw         []byte
		frontmatter frontMatter
		want        string
	}{
		{
			name: "no overlap",
			raw: []byte(`cssclass: index-page
publish: true`),
			frontmatter: frontMatter{
				Title: "211027",
				Alias: "today",
				Tags:  []string{"todo", "math"},
			},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		{
			name: "title overlaps",
			raw: []byte(`cssclass: index-page
publish: true
title: 211026`),
			frontmatter: frontMatter{
				Title: "211027",
				Alias: "today",
				Tags:  []string{"todo", "math"},
			},
			want: `aliases:
- today
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		{
			name: "aliases overlaps",
			raw: []byte(`cssclass: index-page
publish: true
aliases:
- birthday
- today
`),
			frontmatter: frontMatter{
				Title: "211027",
				Alias: "today",
				Tags:  []string{"todo", "math"},
			},
			want: `aliases:
- birthday
- today
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
	}

	for _, tt := range cases {
		got, err := convertYAML(tt.raw, tt.frontmatter)
		if err != nil {
			t.Fatalf("[FATAL] unexpected error occurred: %v", err)
		}
		if string(got) != tt.want {
			t.Errorf("[ERROR]\ngot:\n%q\nwant:\n%q", string(got), tt.want)
		}
	}
}
