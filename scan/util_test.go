package scan

import "testing"

func TestRuneIndex(t *testing.T) {
	cases := []struct {
		s      string
		substr string
		wantId int
	}{
		{
			s:      "[[abc]]",
			substr: "]]",
			wantId: 5,
		},
		{
			s:      "[[abc]]",
			substr: "]]]",
			wantId: -1,
		},
		{
			s:      "[[abc]]",
			substr: "[[",
			wantId: 0,
		},
	}

	for _, tt := range cases {
		gotId := indexInRunes([]rune(tt.s), tt.substr)
		if gotId != tt.wantId {
			t.Errorf("[ERROR] got: %d, want: %d with s: %q, substr: %q", gotId, tt.wantId, tt.s, tt.substr)
		}
	}
}

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
