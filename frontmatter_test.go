package main

import "testing"

func TestConvertYAML(t *testing.T) {
	cases := []struct {
		name        string
		raw         []byte
		frontmatter frontMatter
		flags       *flagBundle
		want        string
	}{
		{
			name: "no overlap",
			raw: []byte(`cssclass: index-page
publish: true`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{},
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
		//////////
		{
			name: "title overlaps",
			raw: []byte(`cssclass: index-page
publish: true
title: 211026`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{},
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
		//////////
		{
			name: "add aliases",
			raw: []byte(`cssclass: index-page
publish: true
aliases:
- today
`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "birthday",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{},
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
		//////////
		{
			name: "alias coincides",
			raw: []byte(`cssclass: index-page
publish: true
aliases:
- today
`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{},
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
		//////////
		{
			name: "add tags",
			raw: []byte(`cssclass: index-page
publish: true
tags:
- book
`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{},
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
		//////////
		{
			name: "tags overlap",
			raw: []byte(`cssclass: index-page
publish: true
tags:
- book
- math
`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{},
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
		//////////
		{
			name: "publishable",
			raw: []byte(`cssclass: index-page
publish: true
`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{publishable: true},
			want: `aliases:
- today
cssclass: index-page
draft: false
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "not publishable",
			raw: []byte(`cssclass: index-page
publish: false`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{publishable: true},
			want: `aliases:
- today
cssclass: index-page
draft: true
publish: false
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "no publish field",
			raw: []byte(`cssclass: index-page
`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{publishable: true},
			want: `aliases:
- today
cssclass: index-page
draft: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "draft field wins publish field",
			raw: []byte(`cssclass: index-page
publish: true
draft: true`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{publishable: true},
			want: `aliases:
- today
cssclass: index-page
draft: true
publish: true
tags:
- todo
- math
title: "211027"
`,
		},
		//////////
		{
			name: "no publishable flag",
			raw: []byte(`cssclass: index-page
publish: true
`),
			frontmatter: frontMatter{
				title: "211027",
				alias: "today",
				tags:  []string{"todo", "math"},
			},
			flags: &flagBundle{},
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
	}

	for _, tt := range cases {
		got, err := convertYAML(tt.raw, tt.frontmatter, tt.flags)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}
		if string(got) != tt.want {
			t.Errorf("[ERROR | %s]\ngot:\n%q\nwant:\n%q", tt.name, string(got), tt.want)
		}
	}
}
