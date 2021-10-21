package main

import "testing"

func TestConsumeTag(t *testing.T) {
	cases := []struct {
		argRaw      []rune
		argPtr      int
		wantAdvance int
		wantTag     string
	}{
		{
			argRaw:      []rune("#todo"),
			argPtr:      0,
			wantAdvance: 5,
			wantTag:     "todo",
		},
		{
			argRaw:      []rune("#_todo"),
			argPtr:      0,
			wantAdvance: 6,
			wantTag:     "_todo",
		},
		{
			argRaw:      []rune("# todo"),
			argPtr:      0,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			argRaw:      []rune("#\ttodo"),
			argPtr:      0,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			argRaw:      []rune("\\#todo"),
			argPtr:      1,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			argRaw:      []rune("##todo"),
			argPtr:      0,
			wantAdvance: 0,
			wantTag:     "",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotTag := consumeTag(tt.argRaw, tt.argPtr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR] got: %v, want: %v", gotAdvance, tt.wantAdvance)
		}
		if gotTag != tt.wantTag {
			t.Errorf("[ERROR] got: %v, want: %v", gotTag, tt.wantTag)
		}
	}
}

func TestConsumeRepeat(t *testing.T) {
	cases := []struct {
		argRaw    []rune
		argSubstr string
		want      int
	}{
		{
			argRaw:    []rune("###x"),
			argSubstr: "#",
			want:      3,
		},
		{
			argRaw:    []rune("----"),
			argSubstr: "-",
			want:      4,
		},
		{
			argRaw:    []rune("$$x$$"),
			argSubstr: "$",
			want:      2,
		},
	}

	for _, tt := range cases {
		if got := consumeRepeat(tt.argRaw, 0, tt.argSubstr); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v", got, tt.want)
		}
	}
}

func TestConsumeInlineMath(t *testing.T) {
	cases := []struct {
		name   string
		argRaw []rune
		argPtr int
		want   int
	}{
		{
			name:   "simple",
			argRaw: []rune("$x$"),
			argPtr: 0,
			want:   3,
		},
		{
			name:   "followed by space",
			argRaw: []rune("$ x$"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "preceded by space",
			argRaw: []rune("$x $"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "preceded \\n",
			argRaw: []rune("$x\n$"),
			argPtr: 0,
			want:   4,
		},
		{
			name:   "preceded \\n\\n",
			argRaw: []rune("$x\n\n$"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "escaped",
			argRaw: []rune("\\$x$"),
			argPtr: 1,
			want:   0,
		},
		{
			name:   "no closing",
			argRaw: []rune("$x"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "include escaped $",
			argRaw: []rune("$#todo\\$$"),
			argPtr: 0,
			want:   9,
		},
	}

	for _, tt := range cases {
		if got := consumeInlineMath(tt.argRaw, tt.argPtr); got != tt.want {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestConsumeInlineCode(t *testing.T) {
	cases := []struct {
		name   string
		argRaw []rune
		argPtr int
		want   int
	}{
		{
			name:   "simple",
			argRaw: []rune("`#todo`"),
			argPtr: 0,
			want:   7,
		},
		{
			name:   "backslashed closing",
			argRaw: []rune("`#todo\\`"),
			argPtr: 0,
			want:   8,
		},
		{
			name:   "escaped opening",
			argRaw: []rune("\\`#todo`"),
			argPtr: 1,
			want:   0,
		},
		{
			name:   "preceded by \\n",
			argRaw: []rune("`\nx\n`"),
			argPtr: 0,
			want:   5,
		},
		{
			name:   "preceded by \\n\\n",
			argRaw: []rune("`x\n\n`"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "no closing",
			argRaw: []rune("`x"),
			argPtr: 0,
			want:   0,
		},
	}

	for _, tt := range cases {
		if got := consumeInlineCode(tt.argRaw, tt.argPtr); got != tt.want {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestConsumeInternalLink(t *testing.T) {
	cases := []struct {
		name        string
		argRaw      []rune
		argPtr      int
		wantAdvance int
		wantContent string
	}{
		{
			name:        "simple",
			argRaw:      []rune("[[ #todo ]]"),
			argPtr:      0,
			wantAdvance: 11,
			wantContent: "#todo",
		},
		{
			name:        "empty",
			argRaw:      []rune("[[]]"),
			argPtr:      0,
			wantAdvance: 0,
			wantContent: "",
		},
		{
			name:        "only spaces",
			argRaw:      []rune("[[ \t]]"),
			argPtr:      0,
			wantAdvance: 6,
			wantContent: "",
		},
		{
			name:        "include \\n",
			argRaw:      []rune("[[x\n]]"),
			argPtr:      0,
			wantAdvance: 0,
			wantContent: "",
		},
		{
			name:        "escaped",
			argRaw:      []rune("\\[[x]]"),
			argPtr:      1,
			wantAdvance: 0,
			wantContent: "",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotContent := consumeInternalLink(tt.argRaw, tt.argPtr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, gotAdvance, tt.wantAdvance)
		}
		if gotContent != tt.wantContent {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, gotContent, tt.wantContent)
		}
	}
}
