package scan

import "testing"

func TestScanTag(t *testing.T) {
	cases := []struct {
		raw         []rune
		ptr         int
		wantAdvance int
		wantTag     string
	}{
		{
			raw:         []rune("#todo"),
			ptr:         0,
			wantAdvance: 5,
			wantTag:     "todo",
		},
		{
			raw:         []rune("#_todo"),
			ptr:         0,
			wantAdvance: 6,
			wantTag:     "_todo",
		},
		{
			raw:         []rune("# todo"),
			ptr:         0,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			raw:         []rune("#\ttodo"),
			ptr:         0,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			raw:         []rune("\\#todo"),
			ptr:         1,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			raw:         []rune("##todo"),
			ptr:         0,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			raw:         []rune("#book/comic"),
			ptr:         0,
			wantAdvance: 11,
			wantTag:     "book/comic",
		},
		{
			raw:         []rune("x#todo"),
			ptr:         1,
			wantAdvance: 0,
			wantTag:     "",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotTag := ScanTag(tt.raw, tt.ptr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR] got: %d, want: %d", gotAdvance, tt.wantAdvance)
		}
		if gotTag != tt.wantTag {
			t.Errorf("[ERROR] got: %q, want: %q", gotTag, tt.wantTag)
		}
	}
}

func TestScanRepeat(t *testing.T) {
	cases := []struct {
		raw       []rune
		argSubstr string
		want      int
	}{
		{
			raw:       []rune("###x"),
			argSubstr: "#",
			want:      3,
		},
		{
			raw:       []rune("----"),
			argSubstr: "-",
			want:      4,
		},
		{
			raw:       []rune("$$x$$"),
			argSubstr: "$",
			want:      2,
		},
	}

	for _, tt := range cases {
		if got := scanRepeat(tt.raw, 0, tt.argSubstr); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v", got, tt.want)
		}
	}
}

func TestScanInlineMath(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		ptr  int
		want int
	}{
		{
			name: "simple",
			raw:  []rune("$x$"),
			ptr:  0,
			want: 3,
		},
		{
			name: "followed by space",
			raw:  []rune("$ x$"),
			ptr:  0,
			want: 0,
		},
		{
			name: "preceded by space",
			raw:  []rune("$x $"),
			ptr:  0,
			want: 0,
		},
		{
			name: "preceded \\n",
			raw:  []rune("$x\n$"),
			ptr:  0,
			want: 4,
		},
		{
			name: "preceded \\n\\n",
			raw:  []rune("$x\n\n$"),
			ptr:  0,
			want: 0,
		},
		{
			name: "escaped",
			raw:  []rune("\\$x$"),
			ptr:  1,
			want: 0,
		},
		{
			name: "no closing",
			raw:  []rune("$x"),
			ptr:  0,
			want: 0,
		},
		{
			name: "include escaped $",
			raw:  []rune("$#todo\\$$"),
			ptr:  0,
			want: 9,
		},
		{
			name: "empty",
			raw:  []rune("$$"),
			ptr:  0,
			want: 0,
		},
	}

	for _, tt := range cases {
		if got := ScanInlineMath(tt.raw, tt.ptr); got != tt.want {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanInlineCode(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		ptr  int
		want int
	}{
		{
			name: "simple",
			raw:  []rune("`#todo`"),
			ptr:  0,
			want: 7,
		},
		{
			name: "backslashed closing",
			raw:  []rune("`#todo\\`"),
			ptr:  0,
			want: 8,
		},
		{
			name: "escaped opening",
			raw:  []rune("\\`#todo`"),
			ptr:  1,
			want: 0,
		},
		{
			name: "preceded by \\n",
			raw:  []rune("`\nx\n`"),
			ptr:  0,
			want: 5,
		},
		{
			name: "preceded by \\n\\n",
			raw:  []rune("`x\n\n`"),
			ptr:  0,
			want: 0,
		},
		{
			name: "no closing",
			raw:  []rune("`x"),
			ptr:  0,
			want: 0,
		},
	}

	for _, tt := range cases {
		if got := ScanInlineCode(tt.raw, tt.ptr); got != tt.want {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanInternalLink(t *testing.T) {
	cases := []struct {
		name        string
		raw         []rune
		ptr         int
		wantAdvance int
		wantContent string
	}{
		{
			name:        "simple",
			raw:         []rune("[[ #todo ]]"),
			ptr:         0,
			wantAdvance: 11,
			wantContent: "#todo",
		},
		{
			name:        "empty",
			raw:         []rune("[[]]"),
			ptr:         0,
			wantAdvance: 0,
			wantContent: "",
		},
		{
			name:        "only spaces",
			raw:         []rune("[[ \t]]"),
			ptr:         0,
			wantAdvance: 6,
			wantContent: "",
		},
		{
			name:        "include \\n",
			raw:         []rune("[[x\n]]"),
			ptr:         0,
			wantAdvance: 0,
			wantContent: "",
		},
		{
			name:        "escaped",
			raw:         []rune("\\[[x]]"),
			ptr:         1,
			wantAdvance: 0,
			wantContent: "",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotContent := ScanInternalLink(tt.raw, tt.ptr)
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

func TestScanExternalLinkHead(t *testing.T) {
	cases := []struct {
		name            string
		raw             []rune
		ptr             int
		wantAdvance     int
		wantDisplayName string
	}{
		{
			name:            "simple",
			raw:             []rune("[ test ]"),
			ptr:             0,
			wantAdvance:     8,
			wantDisplayName: "test",
		},
		{
			name:            "escaped [",
			raw:             []rune("\\[ test ]"),
			ptr:             1,
			wantAdvance:     0,
			wantDisplayName: "",
		},
		{
			name:            "escaped ]",
			raw:             []rune("[ test \\]"),
			ptr:             0,
			wantAdvance:     0,
			wantDisplayName: "",
		},
		{
			name:            "\\n in []",
			raw:             []rune("[ te\nst ]"),
			ptr:             0,
			wantAdvance:     9,
			wantDisplayName: "te\nst",
		},
		{
			name:            "\\n\\n in []",
			raw:             []rune("[ test \n\n]"),
			ptr:             0,
			wantAdvance:     0,
			wantDisplayName: "",
		},
		{
			name:            "[\\]]",
			raw:             []rune("[ test\\] ]"),
			ptr:             0,
			wantAdvance:     10,
			wantDisplayName: "test\\]",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotDisplayName := scanExternalLinkHead(tt.raw, tt.ptr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR | %v]\ngot: %v, want: %v", tt.name, gotAdvance, tt.wantAdvance)
		}
		if gotDisplayName != tt.wantDisplayName {
			t.Errorf("[ERROR | %v]\ngot: %q, want: %q", tt.name, gotDisplayName, tt.wantDisplayName)
		}
	}
}

func TestScanUrl(t *testing.T) {
	cases := []struct {
		raw         []rune
		wantAdvance int
	}{
		{
			raw:         []rune("https://google.com "),
			wantAdvance: 18,
		},
		{
			raw:         []rune("https://google.com)"),
			wantAdvance: 18,
		},
		{
			raw:         []rune("https://google.com\n"),
			wantAdvance: 18,
		},
	}

	for _, tt := range cases {
		if gotAdvance := scanURL(tt.raw, 0); gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR] got: %d, want: %d with raw: %q", gotAdvance, tt.wantAdvance, string(tt.raw))
		}

	}
}

func TestScanLinkTitle(t *testing.T) {
	cases := []struct {
		raw         []rune
		wantAdvance int
		wantTitle   string
	}{
		{
			raw:         []rune(`"title"`),
			wantAdvance: 7,
			wantTitle:   "title",
		},
		{
			raw:         []rune(`"ti\"tle"`),
			wantAdvance: 9,
			wantTitle:   `ti\"tle`,
		},
	}

	for _, tt := range cases {
		gotAdvance, gotTitle := scanLinkTitle(tt.raw, 0)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR] got: %d, want: %d with raw: %q", gotAdvance, tt.wantAdvance, tt.raw)
		}
		if gotTitle != tt.wantTitle {
			t.Errorf("[ERROR] got: %s, want: %s with raw: %q", gotTitle, tt.wantTitle, tt.raw)
		}
	}
}

func TestScanExternalLinkTail(t *testing.T) {
	cases := []struct {
		name        string
		raw         []rune
		ptr         int
		wantAdvance int
		wantRef     string
		wantTitle   string
	}{
		{
			name:        "simple",
			raw:         []rune("( https://google.com#fragment )"),
			ptr:         0,
			wantAdvance: 31,
			wantRef:     "https://google.com#fragment",
		},
		{
			name:        "escaped (",
			raw:         []rune("\\( https://google.com#fragment )"),
			ptr:         0,
			wantAdvance: 0,
			wantRef:     "",
		},
		{
			name:        "escaped )",
			raw:         []rune("( https://google.com#fragment \\)"),
			ptr:         0,
			wantAdvance: 0,
			wantRef:     "",
		},
		{
			name:        "\\n in ()",
			raw:         []rune("(https://google.com\n)"),
			ptr:         0,
			wantAdvance: 0,
			wantRef:     "",
		},
		{
			name:        "\\n\\n in ()",
			raw:         []rune("(https://google.com\n\n)"),
			ptr:         0,
			wantAdvance: 0,
			wantRef:     "",
		},
		{
			name:        "ref contains spaces",
			raw:         []rune("(https://g\noogle.com)"),
			ptr:         0,
			wantAdvance: 0,
			wantRef:     "",
		},
		{
			name:        "(\\))",
			raw:         []rune("(https://google.com\\))"),
			ptr:         0,
			wantAdvance: 0,
			wantRef:     "",
		},
		{
			name:        "with title",
			raw:         []rune("(https://google.com \"google\")"),
			ptr:         0,
			wantAdvance: 29,
			wantRef:     "https://google.com",
			wantTitle:   "google",
		},
		{
			name:        "fileId",
			raw:         []rune("(test)"),
			ptr:         0,
			wantAdvance: 6,
			wantRef:     "test",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotRef, gotTitle := ScanExternalLinkTail(tt.raw, tt.ptr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR | advance - %v]\ngot: %v, want: %v", tt.name, gotAdvance, tt.wantAdvance)
		}
		if gotRef != tt.wantRef {
			t.Errorf("[ERROR | ref - %v]\ngot: %q, want: %q", tt.name, gotRef, tt.wantRef)
		}
		if gotTitle != tt.wantTitle {
			t.Errorf("[ERROR | title - %v] got: %s, want: %s", tt.name, gotTitle, tt.wantTitle)
		}
	}
}

func TestScanExternalLink(t *testing.T) {
	cases := []struct {
		name            string
		raw             []rune
		ptr             int
		wantAdvance     int
		wantDisplayName string
		wantRef         string
	}{
		{
			name:            "] (",
			raw:             []rune("[ test ] (https://google.com)"),
			ptr:             0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
		{
			name:            "[]]()",
			raw:             []rune("[ test] ]( https://google.com#fragment )"),
			ptr:             0,
			wantAdvance:     0,
			wantDisplayName: "",
			wantRef:         "",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotDisplayName, gotRef, _ := ScanExternalLink(tt.raw, tt.ptr)
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
		name string
		raw  []rune
		ptr  int
		want int
	}{
		{
			name: "simple",
			raw:  []rune("%%x%%"),
			ptr:  0,
			want: 5,
		},
		{
			name: "long bracket",
			raw:  []rune("%%%x%%%"),
			ptr:  0,
			want: 7,
		},
		{
			name: "longer closing",
			raw:  []rune("%%x%%%"),
			ptr:  0,
			want: 5,
		},
		{
			name: "longer closing with \\n",
			raw:  []rune("%%\nx\n%%%"),
			ptr:  0,
			want: 7,
		},
		{
			name: "seemingly escaped closing",
			raw:  []rune("%%x\\%%"),
			ptr:  0,
			want: 6,
		},
		{
			name: "escaped opening",
			raw:  []rune("\\%%x%%"),
			ptr:  1,
			want: 0,
		},
		{
			name: "no closing",
			raw:  []rune("%%x"),
			ptr:  0,
			want: 3,
		},
	}

	for _, tt := range cases {
		if got := ScanComment(tt.raw, tt.ptr); got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestValidMathBlockClosing(t *testing.T) {
	cases := []struct {
		name          string
		raw           []rune
		argOpeningPtr int
		argClosingPtr int
		want          bool
	}{
		{
			name:          "escaped",
			raw:           []rune("$$\\$$"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          false,
		},
		{
			name:          "no remaining",
			raw:           []rune("$$\nx\n$$"),
			argOpeningPtr: 0,
			argClosingPtr: 5,
			want:          true,
		},
		{
			name:          "only spaces and \\n are remaining",
			raw:           []rune("$$\nx\n$$   \n\nxxxx"),
			argOpeningPtr: 0,
			argClosingPtr: 5,
			want:          true,
		},
		{
			name:          "ramaining \\t or letter or number or...",
			raw:           []rune("$$\nx\n$$\t\nx"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          false,
		},
		{
			name:          "inline",
			raw:           []rune("$$x$$\t"),
			argOpeningPtr: 0,
			argClosingPtr: 3,
			want:          true,
		},
	}

	for _, tt := range cases {
		if got := validMathBlockClosing(tt.raw, tt.argOpeningPtr, tt.argClosingPtr); got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanMathBlock(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		ptr  int
		want int
	}{
		{
			name: "simple",
			raw:  []rune("$$\nx\n$$"),
			ptr:  0,
			want: 7,
		},
		{
			name: "inline",
			raw:  []rune("$$x$$"),
			ptr:  0,
			want: 5,
		},
		{
			name: "escaped opening",
			raw:  []rune("\\$$x$$"),
			ptr:  1,
			want: 0,
		},
		{
			name: "escaped closing",
			raw:  []rune("$$x\\$$"),
			ptr:  0,
			want: 0,
		},
		{
			name: "preceded by \\n and followed by other than space or \\n",
			raw:  []rune("$$x\n$$\t$$"),
			ptr:  0,
			want: 9,
		},
		{
			name: "inline and followed by other than space or \\n",
			raw:  []rune("$$x$$x"),
			ptr:  0,
			want: 5,
		},
		{
			name: "no closing but ended with \\n",
			raw:  []rune("$$x\n"),
			ptr:  0,
			want: 4,
		},
		{
			name: "no closing and ended with other than \\n",
			raw:  []rune("$$x"),
			ptr:  0,
			want: 0,
		},
	}

	for _, tt := range cases {
		if got := ScanMathBlock(tt.raw, tt.ptr); got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestScanInlineCodeBlock(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		ptr  int
		want int
	}{
		{
			name: "matched brackets",
			raw:  []rune("````f(x)=x````"),
			ptr:  0,
			want: 14,
		},
		{
			name: "longer closing brackets",
			raw:  []rune("```f(x)=x````"),
			ptr:  0,
			want: 0,
		},
		{
			name: "longer opening brackets",
			raw:  []rune("````f(x)=x```"),
			ptr:  0,
			want: 0,
		},
		{
			name: "no closing brackets",
			raw:  []rune("```f(x)=x"),
			ptr:  0,
			want: 9,
		},
		{
			name: "escaped opening brackets",
			raw:  []rune("\\```f(x)=x```"),
			ptr:  1,
			want: 0,
		},
		{
			name: "seemingly escaped closing brackets",
			raw:  []rune("```f(x)=x\\```"),
			ptr:  0,
			want: 13,
		},
	}

	for _, tt := range cases {
		if got := scanInlineCodeBlock(tt.raw, tt.ptr); got != tt.want {
			t.Errorf("[ERROR | %s] got: %d, want: %d", tt.name, got, tt.want)
		}
	}
}

func TestScanMultilineCodeBlock(t *testing.T) {
	cases := []struct {
		name string
		raw  []rune
		ptr  int
		want int
	}{
		{
			name: "simple",
			raw:  []rune("```\nf(x)=x\n```"),
			ptr:  0,
			want: 14,
		},
		{
			name: "long bracket",
			raw:  []rune("````\nf(x)=x\n````"),
			ptr:  0,
			want: 16,
		},
		{
			name: "longer closing",
			raw:  []rune("```\nf(x)=x\n````"),
			ptr:  0,
			want: 15,
		},
		{
			name: "no closing",
			raw:  []rune("```\nf(x)=x\n"),
			ptr:  0,
			want: 11,
		},
		{
			name: "escaped",
			raw:  []rune("\\```\nf(x)=x\n```"),
			ptr:  0,
			want: 0,
		},
		{
			name: "sandwitch",
			raw:  []rune("````\n```\nf(x)=x\n````"),
			ptr:  0,
			want: 20,
		},
		{
			name: "indented closing",
			raw:  []rune("```\nf(x)=x\n\t```\n```"),
			ptr:  0,
			want: 19,
		},
	}

	for _, tt := range cases {
		if got := scanMultilineCodeBlock(tt.raw, tt.ptr); got != tt.want {
			t.Errorf("[ERROR | %s] got: %d, want: %d with raw: %q", tt.name, got, tt.want, string(tt.raw))
		}
	}
}

func TestScanHeader(t *testing.T) {
	cases := []struct {
		name           string
		raw            []rune
		ptr            int
		wantAdvance    int
		wantLevel      int
		wantHeaderText string
	}{
		{name: "simple h1", raw: []rune("# This is a header 1\n"), ptr: 0, wantAdvance: 21, wantLevel: 1, wantHeaderText: "This is a header 1"},
		{name: "simple h6", raw: []rune("###### This is a header 6\n"), ptr: 0, wantAdvance: 26, wantLevel: 6, wantHeaderText: "This is a header 6"},
		{name: "preceded by \\t", raw: []rune("\t# This is not a header 1\n"), ptr: 1, wantAdvance: 0, wantLevel: 0, wantHeaderText: ""},
		{name: "preceded by a letter", raw: []rune("x # This is not a header 1\n"), ptr: 2, wantAdvance: 0, wantLevel: 0, wantHeaderText: ""},
		{name: "preceded by \\n and spaces", raw: []rune("\n # This is not a header 1\n"), ptr: 2, wantAdvance: 25, wantLevel: 1, wantHeaderText: "This is not a header 1"},
		{name: "escaped", raw: []rune("\\# This is not a header 1\n"), ptr: 1, wantAdvance: 0, wantLevel: 0, wantHeaderText: ""},
		{name: "tag", raw: []rune("#This is a tag"), ptr: 0, wantAdvance: 0, wantLevel: 0, wantHeaderText: ""},
		{name: "separated by \\r\\n", raw: []rune("# This is a header 1\r\n"), ptr: 0, wantAdvance: 22, wantLevel: 1, wantHeaderText: "This is a header 1"},
		{name: "immediate \\r\\n", raw: []rune("#\r\n"), ptr: 0, wantAdvance: 3, wantLevel: 1, wantHeaderText: ""},
	}

	for _, tt := range cases {
		gotAdvance, gotLevel, gotHeaderText := ScanHeader(tt.raw, tt.ptr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR | advance - %s] got: %d, want: %d", tt.name, gotAdvance, tt.wantAdvance)
			continue
		}
		if gotLevel != tt.wantLevel {
			t.Errorf("[ERROR | level - %s] got: %d, want: %d", tt.name, gotLevel, tt.wantLevel)
			continue
		}
		if gotHeaderText != tt.wantHeaderText {
			t.Errorf("[ERROR | header text - %s] got: %q, want: %q", tt.name, gotHeaderText, tt.wantHeaderText)
		}
	}
}

func TestScanNormalComment(t *testing.T) {
	cases := []struct {
		name        string
		raw         []rune
		ptr         int
		wantAdvance int
	}{
		{
			name:        "simple",
			raw:         []rune("<!--\ncomment\n-->"),
			ptr:         0,
			wantAdvance: 16,
		},
		{
			name:        "escaped",
			raw:         []rune("\\<!--\ncomment\n-->"),
			ptr:         1,
			wantAdvance: 0,
		},
	}

	for _, tt := range cases {
		gotAdvance := ScanNormalComment(tt.raw, tt.ptr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR | %s] got: %q, want: %q with input: %q", tt.name, gotAdvance, tt.wantAdvance, tt.raw)
		}
	}
}
