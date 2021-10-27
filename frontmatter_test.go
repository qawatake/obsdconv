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
			name: "add aliases",
			raw: []byte(`cssclass: index-page
publish: true
aliases:
- today
`),
			frontmatter: frontMatter{
				Title: "211027",
				Alias: "birthday",
				Tags:  []string{"todo", "math"},
			},
			want: `aliases:
- today
- birthday
cssclass: index-page
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		{
			name: "alias coincides",
			raw: []byte(`cssclass: index-page
publish: true
aliases:
- today
`),
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
			name: "add tags",
			raw: []byte(`cssclass: index-page
publish: true
tags:
- book
`),
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
- book
- todo
- math
title: "211027"
`,
		},
		{
			name: "tags overlap",
			raw: []byte(`cssclass: index-page
publish: true
tags:
- book
- math
`),
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
- book
- math
- todo
title: "211027"
`,
		},
	}

	for _, tt := range cases {
		got, err := convertYAML(tt.raw, tt.frontmatter)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}
		if string(got) != tt.want {
			t.Errorf("[ERROR | %s]\ngot:\n%q\nwant:\n%q", tt.name, string(got), tt.want)
		}
	}
}
