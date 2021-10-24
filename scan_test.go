package main

import "testing"

func TestUnescaped(t *testing.T) {
	cases := []struct {
		name      string
		argRaw    []rune
		argSubstr string
		want      bool
	}{
		{name: "\\ x 1", argRaw: []rune("\\$"), argSubstr: "$", want: false},
		{name: "\\ x 2", argRaw: []rune("\\\\$"), argSubstr: "$", want: true},
		{name: "\\ x 3", argRaw: []rune("\\\\\\$"), argSubstr: "$", want: false},
	}

	for _, tt := range cases {
		if got := unescaped(tt.argRaw, len(tt.argRaw)-1, tt.argSubstr); got != tt.want {
			t.Errorf("[ERROR | %s] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestPrecededBy(t *testing.T) {
	cases := []struct {
		argRaw []rune
		argSs  []string
		want   bool
	}{
		{
			argRaw: []rune("######"),
			argSs:  []string{"##"},
			want:   true,
		},
		{
			argRaw: []rune("\\#"),
			argSs:  []string{"\\"},
			want:   true,
		},
		{
			argRaw: []rune("x $"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []rune("x\t$"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []rune("x$"),
			argSs:  []string{" ", "\t"},
			want:   false,
		},
		{
			argRaw: []rune("x\n\n$"),
			argSs:  []string{" ", "\t", "\n\n", "\r\n\r\n"},
			want:   true,
		},
		{
			argRaw: []rune("x\r\n\r\n$"),
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
		argRaw []rune
		argSs  []string
		want   bool
	}{
		{
			argRaw: []rune("$ x"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []rune("$\tx"),
			argSs:  []string{" ", "\t"},
			want:   true,
		},
		{
			argRaw: []rune("$x"),
			argSs:  []string{" ", "\t"},
			want:   false,
		},
		{
			argRaw: []rune("$\n\nx"),
			argSs:  []string{" ", "\t", "\n\n", "\r\n\r\n"},
			want:   true,
		},
		{
			argRaw: []rune("$\r\n\r\nx"),
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

func TestScanTag(t *testing.T) {
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
		gotAdvance, gotTag := scanTag(tt.argRaw, tt.argPtr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR] got: %v, want: %v", gotAdvance, tt.wantAdvance)
		}
		if gotTag != tt.wantTag {
			t.Errorf("[ERROR] got: %v, want: %v", gotTag, tt.wantTag)
		}
	}
}

func TestScanRepeat(t *testing.T) {
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
		if got := scanRepeat(tt.argRaw, 0, tt.argSubstr); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v", got, tt.want)
		}
	}
}

func TestScanInlineMath(t *testing.T) {
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
		{
			name:   "empty",
			argRaw: []rune("$$"),
			argPtr: 0,
			want:   0,
		},
	}

	for _, tt := range cases {
		if got := scanInlineMath(tt.argRaw, tt.argPtr); got != tt.want {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanInlineCode(t *testing.T) {
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
		if got := scanInlineCode(tt.argRaw, tt.argPtr); got != tt.want {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanInternalLink(t *testing.T) {
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
		gotAdvance, gotContent := scanInternalLink(tt.argRaw, tt.argPtr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, gotAdvance, tt.wantAdvance)
		}
		if gotContent != tt.wantContent {
			t.Errorf("[ERROR | %v]\ngot: %q, want: %q", tt.name, gotContent, tt.wantContent)
		}
	}
}

func TestValidURI(t *testing.T) {
	cases := []struct {
		arg  string
		want bool
	}{
		{arg: "https://google.com", want: true},
		{arg: "https://\ngoogle.com", want: false},
		{arg: "https://\tgoogle.com", want: false},
		{arg: "https://\r\ngoogle.com", want: false},
		{arg: "https:// google.com", want: false},
	}

	for _, tt := range cases {
		if got := validURI(tt.arg); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v", got, tt.want)
		}
	}
}

func TestScanExternalLink(t *testing.T) {
	cases := []struct {
		name            string
		argRaw          []rune
		argPtr          int
		wantAdvance     int
		wantDisplayName string
		wantRef         string
	}{
		{
			name:            "simple",
			argRaw:          []rune("[ test ]( https://google.com#fragment )"),
			argPtr:          0,
			wantAdvance:     39,
			wantDisplayName: "test",
			wantRef:         "https://google.com#fragment",
		},
		{
			name:            "escaped [",
			argRaw:          []rune("\\[ test ]( https://google.com#fragment )"),
			argPtr:          1,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "escaped ]",
			argRaw:          []rune("[ test \\]( https://google.com#fragment )"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "escaped (",
			argRaw:          []rune("[ test ]\\( https://google.com#fragment )"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "escaped )",
			argRaw:          []rune("[ test ]( https://google.com#fragment \\)"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "\\n in []",
			argRaw:          []rune("[ te\nst ]( https://google.com#fragment )"),
			argPtr:          0,
			wantAdvance:     40,
			wantDisplayName: "te\nst",
			wantRef:         "https://google.com#fragment",
		},
		{
			name:            "\\n\\n in []",
			argRaw:          []rune("[ test \n\n](https://google.com)"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "\\n in ()",
			argRaw:          []rune("[ test ](https://google.com\n)"),
			argPtr:          0,
			wantAdvance:     29,
			wantDisplayName: "test",
			wantRef:         "https://google.com",
		},
		{
			name:            "\\n\\n in ()",
			argRaw:          []rune("[ test ](https://google.com\n\n)"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "ref contains spaces",
			argRaw:          []rune("[ test ](https://g\noogle.com)"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "] (",
			argRaw:          []rune("[ test ] (https://google.com)"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "[]]()",
			argRaw:          []rune("[ test] ]( https://google.com#fragment )"),
			argPtr:          0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "[\\]]()",
			argRaw:          []rune("[ test\\] ]( https://google.com#fragment )"),
			argPtr:          0,
			wantAdvance:     41,
			wantDisplayName: "test\\]",
			wantRef:         "https://google.com#fragment",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotDisplayName, gotRef := scanExternalLink(tt.argRaw, tt.argPtr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, gotAdvance, tt.wantAdvance)
		}
		if gotDisplayName != tt.wantDisplayName {
			t.Errorf("[ERROR | %v]\ngot: %q, want: %q", tt.name, gotDisplayName, tt.wantDisplayName)
		}
		if gotRef != tt.wantRef {
			t.Errorf("[ERROR | %v]\ngot: %q, want: %q", tt.name, gotRef, tt.wantRef)
		}
	}
}

func TestScanComment(t *testing.T) {
	cases := []struct {
		name   string
		argRaw []rune
		argPtr int
		want   int
	}{
		{
			name:   "simple",
			argRaw: []rune("%%x%%"),
			argPtr: 0,
			want:   5,
		},
		{
			name:   "long bracket",
			argRaw: []rune("%%%x%%%"),
			argPtr: 0,
			want:   7,
		},
		{
			name:   "longer closing",
			argRaw: []rune("%%x%%%"),
			argPtr: 0,
			want:   5,
		},
		{
			name:   "escaped closing",
			argRaw: []rune("%%x\\%%"),
			argPtr: 0,
			want:   6,
		},
		{
			name:   "escaped opening",
			argRaw: []rune("\\%%x%%"),
			argPtr: 1,
			want:   0,
		},
		{
			name:   "no closing",
			argRaw: []rune("%%x"),
			argPtr: 0,
			want:   3,
		},
	}

	for _, tt := range cases {
		if got := scanComment(tt.argRaw, tt.argPtr); got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestValidMathBlockClosing(t *testing.T) {
	cases := []struct {
		name          string
		argRaw        []rune
		argOpeningPtr int
		argClosingPtr int
		want          bool
	}{
		{
			name:          "escaped",
			argRaw:        []rune("$$\\$$"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          false,
		},
		{
			name:          "no remaining",
			argRaw:        []rune("$$x$$"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          true,
		},
		{
			name:          "only spaces and \\n are remaining",
			argRaw:        []rune("$$x$$   \n\nxxxx"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          true,
		},
		{
			name:          "ramaining \\t or letter or number or...",
			argRaw:        []rune("$$\nx\n$$\t\nx"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          false,
		},
		{
			name:          "inline",
			argRaw:        []rune("$$x$$\t"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          true,
		},
	}

	for _, tt := range cases {
		if got := validMathBlockClosing(tt.argRaw, tt.argOpeningPtr, tt.argClosingPtr); got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanMathBlock(t *testing.T) {
	cases := []struct {
		name   string
		argRaw []rune
		argPtr int
		want   int
	}{
		{
			name:   "simple",
			argRaw: []rune("$$\nx\n$$"),
			argPtr: 0,
			want:   7,
		},
		{
			name:   "inline",
			argRaw: []rune("$$x$$"),
			argPtr: 0,
			want:   5,
		},
		{
			name:   "escaped opening",
			argRaw: []rune("\\$$x$$"),
			argPtr: 1,
			want:   0,
		},
		{
			name:   "escaped closing",
			argRaw: []rune("$$x\\$$"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "preceded by \\n and followed by other than space or \\n",
			argRaw: []rune("$$x\n$$\t$$"),
			argPtr: 0,
			want:   9,
		},
		{
			name:   "inline and followed by other than space or \\n",
			argRaw: []rune("$$x$$x"),
			argPtr: 0,
			want:   5,
		},
		{
			name:   "no closing but ended with \\n",
			argRaw: []rune("$$x\n"),
			argPtr: 0,
			want:   4,
		},
		{
			name:   "no closing and ended with other than \\n",
			argRaw: []rune("$$x"),
			argPtr: 0,
			want:   0,
		},
	}

	for _, tt := range cases {
		if got := scanMathBlock(tt.argRaw, tt.argPtr); got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanCodeBlock(t *testing.T) {
	cases := []struct {
		name   string
		argRaw []rune
		argPtr int
		want   int
	}{
		{
			name:   "simple",
			argRaw: []rune("```\nf(x)=x\n```"),
			argPtr: 0,
			want:   14,
		},
		{
			name:   "long bracket",
			argRaw: []rune("````\nf(x)=x\n````"),
			argPtr: 0,
			want:   16,
		},
		{
			name:   "longer closing",
			argRaw: []rune("```\nf(x)=x\n````"),
			argPtr: 0,
			want:   15,
		},
		{
			name:   "no closing",
			argRaw: []rune("```\nf(x)=x\n"),
			argPtr: 0,
			want:   11,
		},
		{
			name:   "inline with matched brackets",
			argRaw: []rune("````f(x)=x````"),
			argPtr: 0,
			want:   14,
		},
		{
			name:   "inline with longer closing brackets",
			argRaw: []rune("```f(x)=x````"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "inline with longer opening brackets",
			argRaw: []rune("````f(x)=x```"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "escaped",
			argRaw: []rune("\\```\nf(x)=x\n```"),
			argPtr: 0,
			want:   0,
		},
		{
			name:   "sandwitch",
			argRaw: []rune("````\n```\nf(x)=x\n````"),
			argPtr: 0,
			want:   20,
		},
	}

	for _, tt := range cases {
		if got := scanCodeBlock(tt.argRaw, tt.argPtr); got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}
